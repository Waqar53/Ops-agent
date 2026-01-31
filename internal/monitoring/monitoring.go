package monitoring

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// MetricType represents different types of metrics
type MetricType string

const (
	MetricCPU      MetricType = "cpu"
	MetricMemory   MetricType = "memory"
	MetricDisk     MetricType = "disk"
	MetricNetwork  MetricType = "network"
	MetricRequests MetricType = "requests"
	MetricLatency  MetricType = "latency"
	MetricErrors   MetricType = "errors"
	MetricCustom   MetricType = "custom"
)

// Metric represents a single metric data point
type Metric struct {
	ID            string                 `json:"id"`
	ProjectID     string                 `json:"project_id"`
	EnvironmentID string                 `json:"environment_id,omitempty"`
	MetricType    MetricType             `json:"metric_type"`
	Name          string                 `json:"name"`
	Value         float64                `json:"value"`
	Unit          string                 `json:"unit"`
	Tags          map[string]string      `json:"tags,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
}

// Alert represents an alert configuration
type Alert struct {
	ID            string                 `json:"id"`
	ProjectID     string                 `json:"project_id"`
	EnvironmentID *string                `json:"environment_id,omitempty"`
	Name          string                 `json:"name"`
	MetricType    MetricType             `json:"metric_type"`
	Condition     string                 `json:"condition"` // >, <, ==, !=
	Threshold     float64                `json:"threshold"`
	Duration      int                    `json:"duration"` // seconds
	Severity      string                 `json:"severity"` // info, warning, critical
	Enabled       bool                   `json:"enabled"`
	Channels      []string               `json:"channels"` // email, slack, pagerduty
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

// AlertInstance represents a triggered alert
type AlertInstance struct {
	ID            string                 `json:"id"`
	AlertID       string                 `json:"alert_id"`
	ProjectID     string                 `json:"project_id"`
	EnvironmentID *string                `json:"environment_id,omitempty"`
	Title         string                 `json:"title"`
	Message       string                 `json:"message"`
	Severity      string                 `json:"severity"`
	Status        string                 `json:"status"` // triggered, acknowledged, resolved
	TriggeredAt   time.Time              `json:"triggered_at"`
	ResolvedAt    *time.Time             `json:"resolved_at,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// MonitoringService handles metrics and alerting
type MonitoringService struct {
	db *sql.DB
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(db *sql.DB) *MonitoringService {
	return &MonitoringService{db: db}
}

// RecordMetric stores a metric data point
func (ms *MonitoringService) RecordMetric(ctx context.Context, metric *Metric) error {
	tagsJSON, _ := json.Marshal(metric.Tags)
	metadataJSON, _ := json.Marshal(metric.Metadata)

	_, err := ms.db.ExecContext(ctx, `
		INSERT INTO metrics (project_id, environment_id, metric_type, name, value, unit, tags, metadata, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, metric.ProjectID, metric.EnvironmentID, metric.MetricType, metric.Name,
		metric.Value, metric.Unit, tagsJSON, metadataJSON, metric.Timestamp)

	// Check alerts after recording metric
	if err == nil {
		go ms.checkAlerts(metric)
	}

	return err
}

// GetMetrics retrieves metrics for a project
func (ms *MonitoringService) GetMetrics(ctx context.Context, projectID string, metricType MetricType, start, end time.Time) ([]Metric, error) {
	query := `
		SELECT id, project_id, environment_id, metric_type, name, value, unit, tags, metadata, timestamp
		FROM metrics
		WHERE project_id = $1 AND metric_type = $2 AND timestamp BETWEEN $3 AND $4
		ORDER BY timestamp DESC
		LIMIT 1000
	`

	rows, err := ms.db.QueryContext(ctx, query, projectID, metricType, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []Metric
	for rows.Next() {
		var m Metric
		var envID sql.NullString
		var tagsJSON, metadataJSON []byte

		err := rows.Scan(&m.ID, &m.ProjectID, &envID, &m.MetricType, &m.Name,
			&m.Value, &m.Unit, &tagsJSON, &metadataJSON, &m.Timestamp)
		if err != nil {
			continue
		}

		if envID.Valid {
			m.EnvironmentID = envID.String
		}
		json.Unmarshal(tagsJSON, &m.Tags)
		json.Unmarshal(metadataJSON, &m.Metadata)

		metrics = append(metrics, m)
	}

	return metrics, nil
}

// GetAggregatedMetrics returns aggregated metrics (avg, min, max)
func (ms *MonitoringService) GetAggregatedMetrics(ctx context.Context, projectID string, metricType MetricType, interval string, start, end time.Time) ([]map[string]interface{}, error) {
	// Aggregate by time interval (e.g., 1 hour, 1 day)
	query := `
		SELECT 
			date_trunc($1, timestamp) as time_bucket,
			AVG(value) as avg_value,
			MIN(value) as min_value,
			MAX(value) as max_value,
			COUNT(*) as count
		FROM metrics
		WHERE project_id = $2 AND metric_type = $3 AND timestamp BETWEEN $4 AND $5
		GROUP BY time_bucket
		ORDER BY time_bucket DESC
	`

	rows, err := ms.db.QueryContext(ctx, query, interval, projectID, metricType, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var timeBucket time.Time
		var avg, min, max float64
		var count int

		if err := rows.Scan(&timeBucket, &avg, &min, &max, &count); err != nil {
			continue
		}

		results = append(results, map[string]interface{}{
			"timestamp": timeBucket,
			"avg":       avg,
			"min":       min,
			"max":       max,
			"count":     count,
		})
	}

	return results, nil
}

// CreateAlert creates a new alert configuration
func (ms *MonitoringService) CreateAlert(ctx context.Context, alert *Alert) error {
	channelsJSON, _ := json.Marshal(alert.Channels)
	metadataJSON, _ := json.Marshal(alert.Metadata)

	return ms.db.QueryRowContext(ctx, `
		INSERT INTO alert_configs (project_id, environment_id, name, metric_type, condition, threshold, duration, severity, enabled, channels, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at
	`, alert.ProjectID, alert.EnvironmentID, alert.Name, alert.MetricType, alert.Condition,
		alert.Threshold, alert.Duration, alert.Severity, alert.Enabled, channelsJSON, metadataJSON).
		Scan(&alert.ID, &alert.CreatedAt)
}

// checkAlerts evaluates alert conditions
func (ms *MonitoringService) checkAlerts(metric *Metric) {
	// Get all enabled alerts for this project and metric type
	rows, err := ms.db.Query(`
		SELECT id, name, condition, threshold, duration, severity, channels
		FROM alert_configs
		WHERE project_id = $1 AND metric_type = $2 AND enabled = true
	`, metric.ProjectID, metric.MetricType)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var alert Alert
		var channelsJSON []byte
		rows.Scan(&alert.ID, &alert.Name, &alert.Condition, &alert.Threshold,
			&alert.Duration, &alert.Severity, &channelsJSON)
		json.Unmarshal(channelsJSON, &alert.Channels)

		// Evaluate condition
		if ms.evaluateCondition(metric.Value, alert.Condition, alert.Threshold) {
			ms.triggerAlert(&alert, metric)
		}
	}
}

// evaluateCondition checks if a metric value meets alert condition
func (ms *MonitoringService) evaluateCondition(value float64, condition string, threshold float64) bool {
	switch condition {
	case ">":
		return value > threshold
	case "<":
		return value < threshold
	case ">=":
		return value >= threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	case "!=":
		return value != threshold
	default:
		return false
	}
}

// triggerAlert creates an alert instance and sends notifications
func (ms *MonitoringService) triggerAlert(alert *Alert, metric *Metric) {
	// Check if alert already triggered recently
	var exists bool
	ms.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM alerts 
			WHERE alert_type = $1 AND project_id = $2 AND status = 'triggered'
			AND triggered_at > NOW() - INTERVAL '5 minutes'
		)
	`, alert.Name, metric.ProjectID).Scan(&exists)

	if exists {
		return // Don't spam alerts
	}

	// Create alert instance
	title := fmt.Sprintf("%s Alert: %s", alert.Severity, alert.Name)
	message := fmt.Sprintf("Metric %s is %.2f (threshold: %.2f)", metric.Name, metric.Value, alert.Threshold)

	metadataJSON, _ := json.Marshal(map[string]interface{}{
		"metric_value": metric.Value,
		"threshold":    alert.Threshold,
		"condition":    alert.Condition,
	})

	ms.db.Exec(`
		INSERT INTO alerts (project_id, environment_id, alert_type, severity, title, message, status, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, metric.ProjectID, metric.EnvironmentID, alert.Name, alert.Severity, title, message, "triggered", metadataJSON)

	// Send notifications
	for _, channel := range alert.Channels {
		go ms.sendNotification(channel, title, message, alert.Severity)
	}
}

// sendNotification sends alert to configured channels
func (ms *MonitoringService) sendNotification(channel, title, message, severity string) {
	// TODO: Implement actual notification sending
	switch channel {
	case "email":
		// Send email
	case "slack":
		// Send Slack message
	case "pagerduty":
		// Create PagerDuty incident
	case "webhook":
		// POST to webhook URL
	}
}

// GetAlerts retrieves alerts for a project
func (ms *MonitoringService) GetAlerts(ctx context.Context, projectID string, status string) ([]AlertInstance, error) {
	query := `
		SELECT id, project_id, environment_id, alert_type, severity, title, message, status, triggered_at, resolved_at, metadata
		FROM alerts
		WHERE project_id = $1
	`
	args := []interface{}{projectID}

	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}

	query += " ORDER BY triggered_at DESC LIMIT 100"

	rows, err := ms.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []AlertInstance
	for rows.Next() {
		var a AlertInstance
		var envID sql.NullString
		var resolvedAt sql.NullTime
		var metadataJSON []byte

		err := rows.Scan(&a.ID, &a.ProjectID, &envID, &a.AlertID, &a.Severity,
			&a.Title, &a.Message, &a.Status, &a.TriggeredAt, &resolvedAt, &metadataJSON)
		if err != nil {
			continue
		}

		if envID.Valid {
			envStr := envID.String
			a.EnvironmentID = &envStr
		}
		if resolvedAt.Valid {
			a.ResolvedAt = &resolvedAt.Time
		}
		json.Unmarshal(metadataJSON, &a.Metadata)

		alerts = append(alerts, a)
	}

	return alerts, nil
}

// ResolveAlert marks an alert as resolved
func (ms *MonitoringService) ResolveAlert(ctx context.Context, alertID string) error {
	_, err := ms.db.ExecContext(ctx, `
		UPDATE alerts
		SET status = 'resolved', resolved_at = NOW()
		WHERE id = $1
	`, alertID)
	return err
}

// GetMetricsSummary returns current metrics summary for dashboard
func (ms *MonitoringService) GetMetricsSummary(ctx context.Context, projectID string) (map[string]interface{}, error) {
	summary := make(map[string]interface{})

	// Get latest metrics for each type
	metricTypes := []MetricType{MetricCPU, MetricMemory, MetricDisk, MetricRequests, MetricLatency, MetricErrors}

	for _, mt := range metricTypes {
		var value float64
		err := ms.db.QueryRowContext(ctx, `
			SELECT value FROM metrics
			WHERE project_id = $1 AND metric_type = $2
			ORDER BY timestamp DESC
			LIMIT 1
		`, projectID, mt).Scan(&value)

		if err == nil {
			summary[string(mt)] = value
		}
	}

	// Get active alerts count
	var alertCount int
	ms.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM alerts
		WHERE project_id = $1 AND status = 'triggered'
	`, projectID).Scan(&alertCount)
	summary["active_alerts"] = alertCount

	return summary, nil
}

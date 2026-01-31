package deployer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// DeploymentHistory tracks all deployments for rollback
type DeploymentHistory struct {
	storagePath string
}

// DeploymentRecord represents a single deployment in history
type DeploymentRecord struct {
	ID             string                 `json:"id"`
	ProjectID      string                 `json:"project_id"`
	Environment    string                 `json:"environment"`
	Version        string                 `json:"version"`
	Image          string                 `json:"image"`
	Strategy       DeploymentStrategy     `json:"strategy"`
	Status         string                 `json:"status"`
	DeployedAt     time.Time              `json:"deployed_at"`
	DeployedBy     string                 `json:"deployed_by"`
	RollbackFrom   string                 `json:"rollback_from,omitempty"`
	Configuration  map[string]interface{} `json:"configuration"`
	Metrics        *DeploymentMetrics     `json:"metrics,omitempty"`
	Duration       time.Duration          `json:"duration"`
	RollbackReason string                 `json:"rollback_reason,omitempty"`
}

// RollbackManager manages deployment rollbacks
type RollbackManager struct {
	history  *DeploymentHistory
	executor *DeploymentExecutor
	monitor  DeploymentMonitor
}

// RollbackTrigger defines when to automatically rollback
type RollbackTrigger struct {
	ErrorRateThreshold    float64       // e.g., 0.05 for 5%
	LatencyThreshold      time.Duration // e.g., 500ms
	FailedRequestsCount   int           // Absolute number of failed requests
	MonitoringWindow      time.Duration // Time window to evaluate metrics
	ConsecutiveFailures   int           // Number of consecutive health check failures
	CPUThreshold          float64       // e.g., 0.95 for 95%
	MemoryThreshold       float64       // e.g., 0.90 for 90%
	CustomMetricThreshold map[string]float64
}

// NewDeploymentHistory creates a new deployment history tracker
func NewDeploymentHistory(storagePath string) *DeploymentHistory {
	return &DeploymentHistory{
		storagePath: storagePath,
	}
}

// NewRollbackManager creates a new rollback manager
func NewRollbackManager(history *DeploymentHistory, executor *DeploymentExecutor, monitor DeploymentMonitor) *RollbackManager {
	return &RollbackManager{
		history:  history,
		executor: executor,
		monitor:  monitor,
	}
}

// RecordDeployment records a deployment in history
func (dh *DeploymentHistory) RecordDeployment(ctx context.Context, record *DeploymentRecord) error {
	if record.ID == "" {
		record.ID = fmt.Sprintf("deploy_%d", time.Now().UnixNano())
	}

	if err := os.MkdirAll(dh.storagePath, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}

	filename := filepath.Join(dh.storagePath, fmt.Sprintf("%s.json", record.ID))
	return os.WriteFile(filename, data, 0644)
}

// GetDeployment retrieves a deployment record by ID
func (dh *DeploymentHistory) GetDeployment(ctx context.Context, deploymentID string) (*DeploymentRecord, error) {
	filename := filepath.Join(dh.storagePath, fmt.Sprintf("%s.json", deploymentID))
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var record DeploymentRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, err
	}

	return &record, nil
}

// ListDeployments lists all deployments for a project/environment
func (dh *DeploymentHistory) ListDeployments(ctx context.Context, projectID, environment string, limit int) ([]*DeploymentRecord, error) {
	files, err := os.ReadDir(dh.storagePath)
	if err != nil {
		return nil, err
	}

	var records []*DeploymentRecord
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dh.storagePath, file.Name()))
		if err != nil {
			continue
		}

		var record DeploymentRecord
		if err := json.Unmarshal(data, &record); err != nil {
			continue
		}

		if record.ProjectID == projectID && (environment == "" || record.Environment == environment) {
			records = append(records, &record)
		}
	}

	// Sort by deployment time (newest first)
	sort.Slice(records, func(i, j int) bool {
		return records[i].DeployedAt.After(records[j].DeployedAt)
	})

	if limit > 0 && len(records) > limit {
		records = records[:limit]
	}

	return records, nil
}

// GetLastSuccessfulDeployment gets the last successful deployment
func (dh *DeploymentHistory) GetLastSuccessfulDeployment(ctx context.Context, projectID, environment string) (*DeploymentRecord, error) {
	deployments, err := dh.ListDeployments(ctx, projectID, environment, 100)
	if err != nil {
		return nil, err
	}

	for _, deployment := range deployments {
		if deployment.Status == "success" {
			return deployment, nil
		}
	}

	return nil, fmt.Errorf("no successful deployment found")
}

// Rollback performs a rollback to a previous deployment
func (rm *RollbackManager) Rollback(ctx context.Context, projectID, environment, targetDeploymentID string) (*DeploymentResult, error) {
	// Get target deployment
	targetDeployment, err := rm.history.GetDeployment(ctx, targetDeploymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target deployment: %w", err)
	}

	// Get current deployment
	currentDeployments, err := rm.history.ListDeployments(ctx, projectID, environment, 1)
	if err != nil || len(currentDeployments) == 0 {
		return nil, fmt.Errorf("failed to get current deployment")
	}
	currentDeployment := currentDeployments[0]

	// Create rollback deployment config
	config := &DeploymentConfig{
		Strategy:           StrategyDirect, // Fast rollback
		Version:            targetDeployment.Version,
		Image:              targetDeployment.Image,
		Replicas:           3, // Default replicas
		HealthCheckURL:     "/health",
		HealthCheckTimeout: 30 * time.Second,
	}

	// Execute rollback deployment
	result, err := rm.executor.Execute(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("rollback deployment failed: %w", err)
	}

	// Record rollback in history
	rollbackRecord := &DeploymentRecord{
		ProjectID:      projectID,
		Environment:    environment,
		Version:        targetDeployment.Version,
		Image:          targetDeployment.Image,
		Strategy:       StrategyDirect,
		Status:         result.Status,
		DeployedAt:     time.Now(),
		DeployedBy:     "system",
		RollbackFrom:   currentDeployment.ID,
		Configuration:  targetDeployment.Configuration,
		Duration:       result.Duration(),
		RollbackReason: "Manual rollback",
	}

	if err := rm.history.RecordDeployment(ctx, rollbackRecord); err != nil {
		return result, fmt.Errorf("failed to record rollback: %w", err)
	}

	return result, nil
}

// RollbackToLastSuccessful rolls back to the last successful deployment
func (rm *RollbackManager) RollbackToLastSuccessful(ctx context.Context, projectID, environment string) (*DeploymentResult, error) {
	lastSuccessful, err := rm.history.GetLastSuccessfulDeployment(ctx, projectID, environment)
	if err != nil {
		return nil, err
	}

	return rm.Rollback(ctx, projectID, environment, lastSuccessful.ID)
}

// MonitorAndAutoRollback monitors a deployment and automatically rolls back if needed
func (rm *RollbackManager) MonitorAndAutoRollback(ctx context.Context, deploymentID string, trigger *RollbackTrigger) error {
	deployment, err := rm.history.GetDeployment(ctx, deploymentID)
	if err != nil {
		return err
	}

	if trigger == nil {
		trigger = &RollbackTrigger{
			ErrorRateThreshold:  0.05, // 5%
			LatencyThreshold:    500 * time.Millisecond,
			FailedRequestsCount: 100,
			MonitoringWindow:    5 * time.Minute,
			ConsecutiveFailures: 3,
			CPUThreshold:        0.95,
			MemoryThreshold:     0.90,
		}
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	consecutiveFailures := 0
	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Check if monitoring window has passed
			if time.Since(startTime) > trigger.MonitoringWindow {
				return nil // Monitoring complete, no issues
			}

			// Get current metrics
			metrics, err := rm.monitor.GetMetrics(ctx, deployment.Version)
			if err != nil {
				consecutiveFailures++
				if consecutiveFailures >= trigger.ConsecutiveFailures {
					return rm.triggerAutoRollback(ctx, deployment, "Consecutive health check failures")
				}
				continue
			}

			consecutiveFailures = 0

			// Check error rate
			if metrics.ErrorRate > trigger.ErrorRateThreshold {
				return rm.triggerAutoRollback(ctx, deployment,
					fmt.Sprintf("Error rate %.2f%% exceeds threshold %.2f%%",
						metrics.ErrorRate*100, trigger.ErrorRateThreshold*100))
			}

			// Check latency
			if metrics.Latency > trigger.LatencyThreshold {
				return rm.triggerAutoRollback(ctx, deployment,
					fmt.Sprintf("Latency %v exceeds threshold %v",
						metrics.Latency, trigger.LatencyThreshold))
			}

			// Check CPU usage
			if metrics.CPUUsage > trigger.CPUThreshold {
				return rm.triggerAutoRollback(ctx, deployment,
					fmt.Sprintf("CPU usage %.2f%% exceeds threshold %.2f%%",
						metrics.CPUUsage*100, trigger.CPUThreshold*100))
			}

			// Check memory usage
			if metrics.MemoryUsage > trigger.MemoryThreshold {
				return rm.triggerAutoRollback(ctx, deployment,
					fmt.Sprintf("Memory usage %.2f%% exceeds threshold %.2f%%",
						metrics.MemoryUsage*100, trigger.MemoryThreshold*100))
			}
		}
	}
}

// triggerAutoRollback triggers an automatic rollback
func (rm *RollbackManager) triggerAutoRollback(ctx context.Context, deployment *DeploymentRecord, reason string) error {
	fmt.Printf("🔴 AUTO-ROLLBACK TRIGGERED: %s\n", reason)

	result, err := rm.RollbackToLastSuccessful(ctx, deployment.ProjectID, deployment.Environment)
	if err != nil {
		return fmt.Errorf("auto-rollback failed: %w", err)
	}

	// Update deployment record with rollback info
	deployment.Status = "rolled_back"
	deployment.RollbackReason = reason
	rm.history.RecordDeployment(ctx, deployment)

	fmt.Printf("✅ Auto-rollback completed successfully in %v\n", result.Duration())
	return nil
}

// GetRollbackHistory gets the rollback history for a project
func (rm *RollbackManager) GetRollbackHistory(ctx context.Context, projectID, environment string) ([]*DeploymentRecord, error) {
	deployments, err := rm.history.ListDeployments(ctx, projectID, environment, 0)
	if err != nil {
		return nil, err
	}

	var rollbacks []*DeploymentRecord
	for _, deployment := range deployments {
		if deployment.RollbackFrom != "" {
			rollbacks = append(rollbacks, deployment)
		}
	}

	return rollbacks, nil
}

// AnalyzeRollbackTrends analyzes rollback patterns
func (rm *RollbackManager) AnalyzeRollbackTrends(ctx context.Context, projectID string, days int) (*RollbackAnalysis, error) {
	cutoff := time.Now().AddDate(0, 0, -days)
	deployments, err := rm.history.ListDeployments(ctx, projectID, "", 0)
	if err != nil {
		return nil, err
	}

	analysis := &RollbackAnalysis{
		TotalDeployments: 0,
		TotalRollbacks:   0,
		RollbackRate:     0,
		CommonReasons:    make(map[string]int),
		ByEnvironment:    make(map[string]int),
	}

	for _, deployment := range deployments {
		if deployment.DeployedAt.Before(cutoff) {
			continue
		}

		analysis.TotalDeployments++

		if deployment.Status == "rolled_back" || deployment.RollbackFrom != "" {
			analysis.TotalRollbacks++

			if deployment.RollbackReason != "" {
				analysis.CommonReasons[deployment.RollbackReason]++
			}

			analysis.ByEnvironment[deployment.Environment]++
		}
	}

	if analysis.TotalDeployments > 0 {
		analysis.RollbackRate = float64(analysis.TotalRollbacks) / float64(analysis.TotalDeployments)
	}

	return analysis, nil
}

// RollbackAnalysis contains rollback trend analysis
type RollbackAnalysis struct {
	TotalDeployments int
	TotalRollbacks   int
	RollbackRate     float64
	CommonReasons    map[string]int
	ByEnvironment    map[string]int
}

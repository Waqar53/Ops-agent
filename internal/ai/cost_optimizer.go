package ai
import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)
type CostOptimizer struct {
	db *sql.DB
}
func NewCostOptimizer(db *sql.DB) *CostOptimizer {
	return &CostOptimizer{db: db}
}
type Recommendation struct {
	ID               string                 `json:"id"`
	ProjectID        string                 `json:"project_id"`
	Type             string                 `json:"type"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	EstimatedSavings float64                `json:"estimated_savings"`
	Confidence       float64                `json:"confidence"`
	Priority         string                 `json:"priority"`
	Action           string                 `json:"action"`
	Metadata         map[string]interface{} `json:"metadata"`
	Status           string                 `json:"status"`
	CreatedAt        time.Time              `json:"created_at"`
}
type UsagePattern struct {
	ResourceType string   `json:"resource_type"`
	AvgUsage     float64  `json:"avg_usage"`
	PeakUsage    float64  `json:"peak_usage"`
	MinUsage     float64  `json:"min_usage"`
	Utilization  float64  `json:"utilization"`
	PeakHours    []int    `json:"peak_hours"`
	LowUsageDays []string `json:"low_usage_days"`
}
type CostForecast struct {
	Period     string  `json:"period"`
	Forecast   float64 `json:"forecast"`
	LowerBound float64 `json:"lower_bound"`
	UpperBound float64 `json:"upper_bound"`
	Confidence float64 `json:"confidence"`
	Trend      string  `json:"trend"`
	GrowthRate float64 `json:"growth_rate"`
}
func (co *CostOptimizer) AnalyzeUsagePatterns(ctx context.Context, projectID string, days int) ([]UsagePattern, error) {
	startDate := time.Now().AddDate(0, 0, -days)
	cpuPattern, err := co.analyzeResourcePattern(ctx, projectID, "cpu", startDate)
	if err == nil {
		if cpuPattern.Utilization < 0.3 {
			co.createRecommendation(ctx, projectID, "rightsize",
				"Downsize CPU resources",
				fmt.Sprintf("CPU utilization is only %.1f%%. Consider reducing instance size.", cpuPattern.Utilization*100),
				cpuPattern.AvgUsage*0.5,
				0.85,
				"high",
				"Reduce instance type from current to smaller size",
				map[string]interface{}{
					"current_utilization": cpuPattern.Utilization,
					"avg_usage":           cpuPattern.AvgUsage,
				})
		}
	}
	memPattern, err := co.analyzeResourcePattern(ctx, projectID, "memory", startDate)
	if err == nil && memPattern.Utilization < 0.4 {
		co.createRecommendation(ctx, projectID, "rightsize",
			"Reduce memory allocation",
			fmt.Sprintf("Memory utilization is only %.1f%%. Consider reducing memory allocation.", memPattern.Utilization*100),
			memPattern.AvgUsage*0.3,
			0.80,
			"medium",
			"Reduce memory allocation by 30-40%",
			map[string]interface{}{
				"current_utilization": memPattern.Utilization,
			})
	}
	co.detectIdleResources(ctx, projectID, startDate)
	co.detectSpotOpportunities(ctx, projectID, startDate)
	patterns := []UsagePattern{}
	if cpuPattern != nil {
		patterns = append(patterns, *cpuPattern)
	}
	if memPattern != nil {
		patterns = append(patterns, *memPattern)
	}
	return patterns, nil
}
func (co *CostOptimizer) analyzeResourcePattern(ctx context.Context, projectID, resourceType string, startDate time.Time) (*UsagePattern, error) {
	rows, err := co.db.QueryContext(ctx, `
		SELECT value, timestamp
		FROM metrics
		WHERE project_id = $1 AND metric_type = $2 AND timestamp > $3
		ORDER BY timestamp ASC
	`, projectID, resourceType, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var values []float64
	var timestamps []time.Time
	hourlyUsage := make(map[int][]float64)
	for rows.Next() {
		var value float64
		var timestamp time.Time
		if err := rows.Scan(&value, &timestamp); err != nil {
			continue
		}
		values = append(values, value)
		timestamps = append(timestamps, timestamp)
		hour := timestamp.Hour()
		hourlyUsage[hour] = append(hourlyUsage[hour], value)
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("no data")
	}
	pattern := &UsagePattern{
		ResourceType: resourceType,
		AvgUsage:     average(values),
		PeakUsage:    max(values),
		MinUsage:     min(values),
	}
	pattern.Utilization = pattern.AvgUsage / 100.0
	peakHours := []int{}
	avgByHour := make(map[int]float64)
	for hour, vals := range hourlyUsage {
		avgByHour[hour] = average(vals)
	}
	overallAvg := pattern.AvgUsage
	for hour, avg := range avgByHour {
		if avg > overallAvg*1.2 {
			peakHours = append(peakHours, hour)
		}
	}
	pattern.PeakHours = peakHours
	return pattern, nil
}
func (co *CostOptimizer) detectIdleResources(ctx context.Context, projectID string, startDate time.Time) {
	var count int
	co.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT environment_id)
		FROM metrics
		WHERE project_id = $1 AND timestamp > $2 AND value < 5
		GROUP BY environment_id
		HAVING COUNT(*) > 100
	`, projectID, startDate).Scan(&count)
	if count > 0 {
		co.createRecommendation(ctx, projectID, "cleanup",
			fmt.Sprintf("Remove %d idle resources", count),
			"Detected resources with consistently low usage that may be idle or unused.",
			float64(count)*50,
			0.90,
			"high",
			"Review and terminate unused resources",
			map[string]interface{}{
				"idle_count": count,
			})
	}
}
func (co *CostOptimizer) detectSpotOpportunities(ctx context.Context, projectID string, startDate time.Time) {
	var envCount int
	co.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM environments
		WHERE project_id = $1 AND type IN ('development', 'staging')
	`, projectID).Scan(&envCount)
	if envCount > 0 {
		co.createRecommendation(ctx, projectID, "spot",
			"Use spot instances for non-production",
			fmt.Sprintf("Save up to 90%% on %d non-production environments by using spot instances.", envCount),
			float64(envCount)*100*0.9,
			0.75,
			"high",
			"Migrate development and staging to spot instances",
			map[string]interface{}{
				"environment_count":         envCount,
				"potential_savings_percent": 90,
			})
	}
}
func (co *CostOptimizer) ForecastCosts(ctx context.Context, projectID string, period string) (*CostForecast, error) {
	days := 30
	if period == "60d" {
		days = 60
	} else if period == "90d" {
		days = 90
	}
	startDate := time.Now().AddDate(0, 0, -days)
	rows, err := co.db.QueryContext(ctx, `
		SELECT DATE(recorded_at) as date, SUM(cost) as daily_cost
		FROM usage_records
		WHERE organization_id IN (
			SELECT organization_id FROM projects WHERE id = $1
		) AND recorded_at > $2
		GROUP BY DATE(recorded_at)
		ORDER BY date ASC
	`, projectID, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var dailyCosts []float64
	for rows.Next() {
		var date time.Time
		var cost float64
		if err := rows.Scan(&date, &cost); err != nil {
			continue
		}
		dailyCosts = append(dailyCosts, cost)
	}
	if len(dailyCosts) < 7 {
		return nil, fmt.Errorf("insufficient data for forecasting")
	}
	avgCost := average(dailyCosts)
	trend := calculateTrend(dailyCosts)
	forecastDays := 30
	if period == "60d" {
		forecastDays = 60
	} else if period == "90d" {
		forecastDays = 90
	}
	forecast := avgCost * float64(forecastDays)
	if trend > 0 {
		forecast *= (1 + trend)
	}
	lowerBound := forecast * 0.8
	upperBound := forecast * 1.2
	trendDirection := "stable"
	if trend > 0.05 {
		trendDirection = "increasing"
	} else if trend < -0.05 {
		trendDirection = "decreasing"
	}
	return &CostForecast{
		Period:     period,
		Forecast:   forecast,
		LowerBound: lowerBound,
		UpperBound: upperBound,
		Confidence: 0.75,
		Trend:      trendDirection,
		GrowthRate: trend * 100,
	}, nil
}
func (co *CostOptimizer) GetRecommendations(ctx context.Context, projectID string, status string) ([]Recommendation, error) {
	query := `
		SELECT id, project_id, type, title, description, estimated_savings, 
		       confidence, priority, action, metadata, status, created_at
		FROM cost_recommendations
		WHERE project_id = $1
	`
	args := []interface{}{projectID}
	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}
	query += " ORDER BY estimated_savings DESC, created_at DESC"
	rows, err := co.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var recommendations []Recommendation
	for rows.Next() {
		var r Recommendation
		var metadataJSON []byte
		err := rows.Scan(&r.ID, &r.ProjectID, &r.Type, &r.Title, &r.Description,
			&r.EstimatedSavings, &r.Confidence, &r.Priority, &r.Action,
			&metadataJSON, &r.Status, &r.CreatedAt)
		if err != nil {
			continue
		}
		json.Unmarshal(metadataJSON, &r.Metadata)
		recommendations = append(recommendations, r)
	}
	return recommendations, nil
}
func (co *CostOptimizer) createRecommendation(ctx context.Context, projectID, recType, title, description string,
	savings, confidence float64, priority, action string, metadata map[string]interface{}) error {
	metadataJSON, _ := json.Marshal(metadata)
	_, err := co.db.ExecContext(ctx, `
		INSERT INTO cost_recommendations 
		(project_id, type, title, description, estimated_savings, confidence, priority, action, metadata, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, projectID, recType, title, description, savings, confidence, priority, action, metadataJSON, "pending")
	return err
}
func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
func max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	maxVal := values[0]
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
}
func min(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	minVal := values[0]
	for _, v := range values {
		if v < minVal {
			minVal = v
		}
	}
	return minVal
}
func calculateTrend(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	n := float64(len(values))
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumX2 := 0.0
	for i, y := range values {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	avgY := sumY / n
	if avgY == 0 {
		return 0
	}
	return slope / avgY
}
func (co *CostOptimizer) ApplyRecommendation(ctx context.Context, recommendationID string) error {
	_, err := co.db.ExecContext(ctx, `
		UPDATE cost_recommendations
		SET status = 'applied'
		WHERE id = $1
	`, recommendationID)
	return err
}
func (co *CostOptimizer) DismissRecommendation(ctx context.Context, recommendationID string) error {
	_, err := co.db.ExecContext(ctx, `
		UPDATE cost_recommendations
		SET status = 'dismissed'
		WHERE id = $1
	`, recommendationID)
	return err
}

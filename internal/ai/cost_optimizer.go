package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// CostOptimizer provides AI-powered cost optimization recommendations
type CostOptimizer struct {
	db *sql.DB
}

// NewCostOptimizer creates a new cost optimizer
func NewCostOptimizer(db *sql.DB) *CostOptimizer {
	return &CostOptimizer{db: db}
}

// Recommendation represents a cost optimization recommendation
type Recommendation struct {
	ID               string                 `json:"id"`
	ProjectID        string                 `json:"project_id"`
	Type             string                 `json:"type"` // rightsize, spot, reserved, cleanup, schedule
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	EstimatedSavings float64                `json:"estimated_savings"` // USD per month
	Confidence       float64                `json:"confidence"`        // 0-1
	Priority         string                 `json:"priority"`          // low, medium, high
	Action           string                 `json:"action"`            // What to do
	Metadata         map[string]interface{} `json:"metadata"`
	Status           string                 `json:"status"` // pending, applied, dismissed
	CreatedAt        time.Time              `json:"created_at"`
}

// UsagePattern represents detected usage patterns
type UsagePattern struct {
	ResourceType string   `json:"resource_type"`
	AvgUsage     float64  `json:"avg_usage"`
	PeakUsage    float64  `json:"peak_usage"`
	MinUsage     float64  `json:"min_usage"`
	Utilization  float64  `json:"utilization"`    // 0-1
	PeakHours    []int    `json:"peak_hours"`     // Hours of day (0-23)
	LowUsageDays []string `json:"low_usage_days"` // Days of week
}

// CostForecast represents predicted costs
type CostForecast struct {
	Period     string  `json:"period"`      // 30d, 60d, 90d
	Forecast   float64 `json:"forecast"`    // USD
	LowerBound float64 `json:"lower_bound"` // Best case
	UpperBound float64 `json:"upper_bound"` // Worst case
	Confidence float64 `json:"confidence"`  // 0-1
	Trend      string  `json:"trend"`       // increasing, decreasing, stable
	GrowthRate float64 `json:"growth_rate"` // % per month
}

// AnalyzeUsagePatterns analyzes resource usage patterns
func (co *CostOptimizer) AnalyzeUsagePatterns(ctx context.Context, projectID string, days int) ([]UsagePattern, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	// Analyze CPU usage
	cpuPattern, err := co.analyzeResourcePattern(ctx, projectID, "cpu", startDate)
	if err == nil {
		// Generate recommendations based on CPU pattern
		if cpuPattern.Utilization < 0.3 {
			co.createRecommendation(ctx, projectID, "rightsize",
				"Downsize CPU resources",
				fmt.Sprintf("CPU utilization is only %.1f%%. Consider reducing instance size.", cpuPattern.Utilization*100),
				cpuPattern.AvgUsage*0.5, // Estimated savings
				0.85,                    // Confidence
				"high",
				"Reduce instance type from current to smaller size",
				map[string]interface{}{
					"current_utilization": cpuPattern.Utilization,
					"avg_usage":           cpuPattern.AvgUsage,
				})
		}
	}

	// Analyze memory usage
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

	// Check for idle resources
	co.detectIdleResources(ctx, projectID, startDate)

	// Check for spot instance opportunities
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

// analyzeResourcePattern analyzes a specific resource type
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
	hourlyUsage := make(map[int][]float64) // Hour of day -> values

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

	// Calculate statistics
	pattern := &UsagePattern{
		ResourceType: resourceType,
		AvgUsage:     average(values),
		PeakUsage:    max(values),
		MinUsage:     min(values),
	}

	// Calculate utilization (assuming 100 is max)
	pattern.Utilization = pattern.AvgUsage / 100.0

	// Find peak hours
	peakHours := []int{}
	avgByHour := make(map[int]float64)
	for hour, vals := range hourlyUsage {
		avgByHour[hour] = average(vals)
	}
	overallAvg := pattern.AvgUsage
	for hour, avg := range avgByHour {
		if avg > overallAvg*1.2 { // 20% above average
			peakHours = append(peakHours, hour)
		}
	}
	pattern.PeakHours = peakHours

	return pattern, nil
}

// detectIdleResources finds resources with very low usage
func (co *CostOptimizer) detectIdleResources(ctx context.Context, projectID string, startDate time.Time) {
	// Check for resources with < 5% utilization
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
			float64(count)*50, // $50 per resource
			0.90,
			"high",
			"Review and terminate unused resources",
			map[string]interface{}{
				"idle_count": count,
			})
	}
}

// detectSpotOpportunities finds workloads suitable for spot instances
func (co *CostOptimizer) detectSpotOpportunities(ctx context.Context, projectID string, startDate time.Time) {
	// Check for non-production environments
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
			float64(envCount)*100*0.9, // 90% savings
			0.75,
			"high",
			"Migrate development and staging to spot instances",
			map[string]interface{}{
				"environment_count":         envCount,
				"potential_savings_percent": 90,
			})
	}
}

// ForecastCosts predicts future costs
func (co *CostOptimizer) ForecastCosts(ctx context.Context, projectID string, period string) (*CostForecast, error) {
	// Get historical cost data
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

	// Simple linear regression for trend
	avgCost := average(dailyCosts)
	trend := calculateTrend(dailyCosts)

	// Forecast
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

	// Calculate bounds (±20%)
	lowerBound := forecast * 0.8
	upperBound := forecast * 1.2

	// Determine trend direction
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

// GetRecommendations retrieves cost optimization recommendations
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

// createRecommendation creates a new recommendation
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

// Helper functions
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

	// Simple linear regression slope
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

	return slope / avgY // Normalized slope
}

// ApplyRecommendation marks a recommendation as applied
func (co *CostOptimizer) ApplyRecommendation(ctx context.Context, recommendationID string) error {
	_, err := co.db.ExecContext(ctx, `
		UPDATE cost_recommendations
		SET status = 'applied'
		WHERE id = $1
	`, recommendationID)
	return err
}

// DismissRecommendation marks a recommendation as dismissed
func (co *CostOptimizer) DismissRecommendation(ctx context.Context, recommendationID string) error {
	_, err := co.db.ExecContext(ctx, `
		UPDATE cost_recommendations
		SET status = 'dismissed'
		WHERE id = $1
	`, recommendationID)
	return err
}

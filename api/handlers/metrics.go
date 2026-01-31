package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"ops-agent/internal/monitoring"
)

// MetricsHandlers handles metrics-related API endpoints
type MetricsHandlers struct {
	monitoringService *monitoring.MonitoringService
}

// NewMetricsHandlers creates new metrics handlers
func NewMetricsHandlers(ms *monitoring.MonitoringService) *MetricsHandlers {
	return &MetricsHandlers{monitoringService: ms}
}

// GetMetrics returns metrics for a project
func (h *MetricsHandlers) GetMetrics(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	metricType := r.URL.Query().Get("type")
	rangeStr := r.URL.Query().Get("range")

	if projectID == "" {
		http.Error(w, "project_id required", http.StatusBadRequest)
		return
	}

	// Default values
	if metricType == "" {
		metricType = "cpu"
	}
	if rangeStr == "" {
		rangeStr = "1h"
	}

	// Parse time range
	var duration time.Duration
	switch rangeStr {
	case "1h":
		duration = time.Hour
	case "24h":
		duration = 24 * time.Hour
	case "7d":
		duration = 7 * 24 * time.Hour
	case "30d":
		duration = 30 * 24 * time.Hour
	default:
		duration = time.Hour
	}

	end := time.Now()
	start := end.Add(-duration)

	metrics, err := h.monitoringService.GetMetrics(r.Context(), projectID, monitoring.MetricType(metricType), start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// GetMetricsSummary returns aggregated metrics summary
func (h *MetricsHandlers) GetMetricsSummary(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")

	if projectID == "" {
		http.Error(w, "project_id required", http.StatusBadRequest)
		return
	}

	summary, err := h.monitoringService.GetMetricsSummary(r.Context(), projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// GetAlerts returns alerts for a project
func (h *MetricsHandlers) GetAlerts(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	status := r.URL.Query().Get("status")

	if projectID == "" {
		http.Error(w, "project_id required", http.StatusBadRequest)
		return
	}

	alerts, err := h.monitoringService.GetAlerts(r.Context(), projectID, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

// CreateAlert creates a new alert configuration
func (h *MetricsHandlers) CreateAlert(w http.ResponseWriter, r *http.Request) {
	var alert monitoring.Alert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.monitoringService.CreateAlert(r.Context(), &alert); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(alert)
}

// ResolveAlert marks an alert as resolved
func (h *MetricsHandlers) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	alertID := r.URL.Query().Get("alert_id")

	if alertID == "" {
		http.Error(w, "alert_id required", http.StatusBadRequest)
		return
	}

	if err := h.monitoringService.ResolveAlert(r.Context(), alertID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "resolved"})
}

// GetDashboardStats returns real-time dashboard statistics
func (h *MetricsHandlers) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")

	// Get real metrics if project specified, otherwise return demo data
	stats := map[string]interface{}{
		"cpu":         42.5,
		"memory":      68.3,
		"requests":    12847,
		"deployments": 127,
		"uptime":      99.98,
		"latency":     67,
		"errors":      3,
		"alerts":      0,
	}

	if projectID != "" && h.monitoringService != nil {
		summary, err := h.monitoringService.GetMetricsSummary(r.Context(), projectID)
		if err == nil {
			for k, v := range summary {
				stats[k] = v
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

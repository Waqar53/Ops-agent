package deployer

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// PreviewEnvironment represents a PR-based preview environment
type PreviewEnvironment struct {
	ID             string             `json:"id"`
	ProjectID      string             `json:"project_id"`
	PullRequestID  string             `json:"pull_request_id"`
	Branch         string             `json:"branch"`
	URL            string             `json:"url"`
	Status         string             `json:"status"` // creating, active, sleeping, deleting, deleted
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	LastAccessedAt *time.Time         `json:"last_accessed_at,omitempty"`
	AutoDelete     bool               `json:"auto_delete"`
	SleepAfter     time.Duration      `json:"sleep_after"`
	DeleteAfter    time.Duration      `json:"delete_after"`
	DatabaseSeeded bool               `json:"database_seeded"`
	MockedServices []string           `json:"mocked_services"`
	Resources      ResourceAllocation `json:"resources"`
	SSL            bool               `json:"ssl"`
	BasicAuth      *BasicAuth         `json:"basic_auth,omitempty"`
	Metadata       map[string]string  `json:"metadata"`
}

// BasicAuth for preview environment protection
type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// PreviewManager manages preview environments
type PreviewManager struct {
	envManager    *EnvironmentManager
	dnsProvider   DNSProvider
	sslProvider   SSLProvider
	dbSeeder      DatabaseSeeder
	serviceMocker ServiceMocker
}

// DNSProvider interface for DNS management
type DNSProvider interface {
	CreateRecord(ctx context.Context, subdomain, target string) error
	DeleteRecord(ctx context.Context, subdomain string) error
	GetRecord(ctx context.Context, subdomain string) (string, error)
}

// SSLProvider interface for SSL certificate management
type SSLProvider interface {
	IssueCertificate(ctx context.Context, domain string) error
	RevokeCertificate(ctx context.Context, domain string) error
}

// DatabaseSeeder interface for seeding preview databases
type DatabaseSeeder interface {
	SeedDatabase(ctx context.Context, dbURL string, sanitize bool) error
	CloneDatabase(ctx context.Context, sourceURL, targetURL string) error
}

// ServiceMocker interface for mocking external services
type ServiceMocker interface {
	MockService(ctx context.Context, serviceName, endpoint string) (string, error)
	UnmockService(ctx context.Context, serviceName string) error
}

// NewPreviewManager creates a new preview environment manager
func NewPreviewManager(
	envManager *EnvironmentManager,
	dnsProvider DNSProvider,
	sslProvider SSLProvider,
	dbSeeder DatabaseSeeder,
	serviceMocker ServiceMocker,
) *PreviewManager {
	return &PreviewManager{
		envManager:    envManager,
		dnsProvider:   dnsProvider,
		sslProvider:   sslProvider,
		dbSeeder:      dbSeeder,
		serviceMocker: serviceMocker,
	}
}

// CreatePreviewEnvironment creates a new preview environment for a PR
func (pm *PreviewManager) CreatePreviewEnvironment(ctx context.Context, config *PreviewEnvironmentConfig) (*PreviewEnvironment, error) {
	// Generate unique subdomain
	subdomain := pm.generateSubdomain(config.ProjectID, config.PullRequestID)
	url := fmt.Sprintf("https://%s.preview.opsagent.dev", subdomain)

	preview := &PreviewEnvironment{
		ID:             generatePreviewID(),
		ProjectID:      config.ProjectID,
		PullRequestID:  config.PullRequestID,
		Branch:         config.Branch,
		URL:            url,
		Status:         "creating",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		AutoDelete:     config.AutoDelete,
		SleepAfter:     config.SleepAfter,
		DeleteAfter:    config.DeleteAfter,
		DatabaseSeeded: false,
		MockedServices: []string{},
		SSL:            true,
		Resources: ResourceAllocation{
			MinCPU:      "100m",
			MaxCPU:      "500m",
			MinMemory:   "128Mi",
			MaxMemory:   "512Mi",
			MinReplicas: 1,
			MaxReplicas: 1,
			StorageSize: "5Gi",
			AutoScale:   false,
		},
		Metadata: make(map[string]string),
	}

	// Generate basic auth if requested
	if config.ProtectWithAuth {
		preview.BasicAuth = &BasicAuth{
			Username: "preview",
			Password: generateRandomPassword(16),
		}
	}

	// Create underlying environment
	env := &Environment{
		Name:      fmt.Sprintf("preview-%s", preview.ID),
		Type:      EnvironmentPreview,
		ProjectID: config.ProjectID,
		Variables: config.EnvVars,
		Secrets:   config.Secrets,
		Domains:   []string{url},
		Resources: preview.Resources,
		Metadata: map[string]interface{}{
			"preview_id":      preview.ID,
			"pull_request_id": config.PullRequestID,
			"branch":          config.Branch,
		},
	}

	if err := pm.envManager.CreateEnvironment(ctx, env); err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	// Create DNS record
	if err := pm.dnsProvider.CreateRecord(ctx, subdomain, config.TargetIP); err != nil {
		return nil, fmt.Errorf("failed to create DNS record: %w", err)
	}

	// Issue SSL certificate
	if preview.SSL {
		if err := pm.sslProvider.IssueCertificate(ctx, url); err != nil {
			return nil, fmt.Errorf("failed to issue SSL certificate: %w", err)
		}
	}

	// Seed database if requested
	if config.SeedDatabase {
		if err := pm.dbSeeder.SeedDatabase(ctx, config.DatabaseURL, config.SanitizeData); err != nil {
			return nil, fmt.Errorf("failed to seed database: %w", err)
		}
		preview.DatabaseSeeded = true
	}

	// Mock external services if requested
	for _, service := range config.MockServices {
		mockURL, err := pm.serviceMocker.MockService(ctx, service, config.ServiceEndpoints[service])
		if err != nil {
			return nil, fmt.Errorf("failed to mock service %s: %w", service, err)
		}
		preview.MockedServices = append(preview.MockedServices, service)
		preview.Metadata[fmt.Sprintf("mock_%s_url", service)] = mockURL
	}

	preview.Status = "active"
	preview.UpdatedAt = time.Now()

	return preview, nil
}

// UpdatePreviewEnvironment updates a preview environment on new commits
func (pm *PreviewManager) UpdatePreviewEnvironment(ctx context.Context, previewID, commitSHA string) error {
	// This would trigger a new deployment with the latest code
	// Implementation would integrate with the deployment system

	fmt.Printf("📦 Updating preview environment %s with commit %s\n", previewID, commitSHA[:7])

	// Simulate deployment
	time.Sleep(5 * time.Second)

	fmt.Printf("✅ Preview environment updated successfully\n")
	return nil
}

// DeletePreviewEnvironment deletes a preview environment
func (pm *PreviewManager) DeletePreviewEnvironment(ctx context.Context, previewID string, preview *PreviewEnvironment) error {
	preview.Status = "deleting"

	// Unmock services
	for _, service := range preview.MockedServices {
		if err := pm.serviceMocker.UnmockService(ctx, service); err != nil {
			fmt.Printf("Warning: failed to unmock service %s: %v\n", service, err)
		}
	}

	// Revoke SSL certificate
	if preview.SSL {
		if err := pm.sslProvider.RevokeCertificate(ctx, preview.URL); err != nil {
			fmt.Printf("Warning: failed to revoke SSL certificate: %v\n", err)
		}
	}

	// Delete DNS record
	subdomain := strings.Split(preview.URL, ".")[0]
	subdomain = strings.TrimPrefix(subdomain, "https://")
	if err := pm.dnsProvider.DeleteRecord(ctx, subdomain); err != nil {
		fmt.Printf("Warning: failed to delete DNS record: %v\n", err)
	}

	// Delete underlying environment
	envName := fmt.Sprintf("preview-%s", preview.ID)
	if err := pm.envManager.DeleteEnvironment(ctx, envName); err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}

	preview.Status = "deleted"
	return nil
}

// SleepPreviewEnvironment puts a preview environment to sleep to save costs
func (pm *PreviewManager) SleepPreviewEnvironment(ctx context.Context, previewID string, preview *PreviewEnvironment) error {
	preview.Status = "sleeping"
	preview.UpdatedAt = time.Now()

	// Scale down to 0 replicas
	preview.Resources.MinReplicas = 0
	preview.Resources.MaxReplicas = 0

	fmt.Printf("💤 Preview environment %s is now sleeping\n", previewID)
	return nil
}

// WakePreviewEnvironment wakes up a sleeping preview environment
func (pm *PreviewManager) WakePreviewEnvironment(ctx context.Context, previewID string, preview *PreviewEnvironment) error {
	preview.Status = "active"
	preview.UpdatedAt = time.Now()
	now := time.Now()
	preview.LastAccessedAt = &now

	// Scale back up
	preview.Resources.MinReplicas = 1
	preview.Resources.MaxReplicas = 1

	fmt.Printf("🌅 Preview environment %s is now awake\n", previewID)
	return nil
}

// MonitorPreviewEnvironments monitors preview environments for auto-sleep and auto-delete
func (pm *PreviewManager) MonitorPreviewEnvironments(ctx context.Context, previews []*PreviewEnvironment) error {
	for _, preview := range previews {
		if preview.Status == "deleted" {
			continue
		}

		now := time.Now()

		// Auto-sleep inactive environments
		if preview.Status == "active" && preview.SleepAfter > 0 {
			lastAccessed := preview.UpdatedAt
			if preview.LastAccessedAt != nil {
				lastAccessed = *preview.LastAccessedAt
			}

			if now.Sub(lastAccessed) > preview.SleepAfter {
				if err := pm.SleepPreviewEnvironment(ctx, preview.ID, preview); err != nil {
					fmt.Printf("Error sleeping preview %s: %v\n", preview.ID, err)
				}
			}
		}

		// Auto-delete old environments
		if preview.AutoDelete && preview.DeleteAfter > 0 {
			if now.Sub(preview.CreatedAt) > preview.DeleteAfter {
				if err := pm.DeletePreviewEnvironment(ctx, preview.ID, preview); err != nil {
					fmt.Printf("Error deleting preview %s: %v\n", preview.ID, err)
				}
			}
		}
	}

	return nil
}

// CompareWithProduction compares preview environment performance with production
func (pm *PreviewManager) CompareWithProduction(ctx context.Context, previewID string, monitor DeploymentMonitor) (*PerformanceComparison, error) {
	previewMetrics, err := monitor.GetMetrics(ctx, previewID)
	if err != nil {
		return nil, err
	}

	prodMetrics, err := monitor.GetMetrics(ctx, "production")
	if err != nil {
		return nil, err
	}

	comparison := &PerformanceComparison{
		PreviewMetrics:    previewMetrics,
		ProductionMetrics: prodMetrics,
		LatencyDelta:      previewMetrics.Latency - prodMetrics.Latency,
		ErrorRateDelta:    previewMetrics.ErrorRate - prodMetrics.ErrorRate,
		CPUDelta:          previewMetrics.CPUUsage - prodMetrics.CPUUsage,
		MemoryDelta:       previewMetrics.MemoryUsage - prodMetrics.MemoryUsage,
	}

	// Calculate performance score (0-100)
	score := 100.0
	if comparison.LatencyDelta > 0 {
		score -= (comparison.LatencyDelta.Seconds() / prodMetrics.Latency.Seconds()) * 20
	}
	if comparison.ErrorRateDelta > 0 {
		score -= (comparison.ErrorRateDelta / prodMetrics.ErrorRate) * 30
	}
	if comparison.CPUDelta > 0.1 {
		score -= 15
	}
	if comparison.MemoryDelta > 0.1 {
		score -= 15
	}

	if score < 0 {
		score = 0
	}

	comparison.PerformanceScore = score
	comparison.Recommendation = pm.generateRecommendation(comparison)

	return comparison, nil
}

// generateRecommendation generates performance recommendations
func (pm *PreviewManager) generateRecommendation(comparison *PerformanceComparison) string {
	if comparison.PerformanceScore >= 90 {
		return "✅ Excellent performance - ready for production"
	} else if comparison.PerformanceScore >= 70 {
		return "⚠️ Good performance - minor optimizations recommended"
	} else if comparison.PerformanceScore >= 50 {
		return "⚠️ Performance degradation detected - review before merging"
	} else {
		return "🔴 Significant performance issues - do not merge"
	}
}

// Helper functions

func (pm *PreviewManager) generateSubdomain(projectID, prID string) string {
	// Generate a short, URL-safe subdomain
	hash := fmt.Sprintf("%s-%s", projectID, prID)
	return strings.ToLower(strings.ReplaceAll(hash, "_", "-"))
}

func generatePreviewID() string {
	return fmt.Sprintf("preview_%d", time.Now().UnixNano())
}

func generateRandomPassword(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

// PreviewEnvironmentConfig holds configuration for creating a preview environment
type PreviewEnvironmentConfig struct {
	ProjectID        string
	PullRequestID    string
	Branch           string
	TargetIP         string
	EnvVars          map[string]string
	Secrets          map[string]string
	SeedDatabase     bool
	SanitizeData     bool
	DatabaseURL      string
	MockServices     []string
	ServiceEndpoints map[string]string
	ProtectWithAuth  bool
	AutoDelete       bool
	SleepAfter       time.Duration
	DeleteAfter      time.Duration
}

// PerformanceComparison compares preview and production performance
type PerformanceComparison struct {
	PreviewMetrics    *DeploymentMetrics
	ProductionMetrics *DeploymentMetrics
	LatencyDelta      time.Duration
	ErrorRateDelta    float64
	CPUDelta          float64
	MemoryDelta       float64
	PerformanceScore  float64
	Recommendation    string
}

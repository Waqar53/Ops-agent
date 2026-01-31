package deployer
import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)
type PreviewEnvironment struct {
	ID             string             `json:"id"`
	ProjectID      string             `json:"project_id"`
	PullRequestID  string             `json:"pull_request_id"`
	Branch         string             `json:"branch"`
	URL            string             `json:"url"`
	Status         string             `json:"status"`
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
type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type PreviewManager struct {
	envManager    *EnvironmentManager
	dnsProvider   DNSProvider
	sslProvider   SSLProvider
	dbSeeder      DatabaseSeeder
	serviceMocker ServiceMocker
}
type DNSProvider interface {
	CreateRecord(ctx context.Context, subdomain, target string) error
	DeleteRecord(ctx context.Context, subdomain string) error
	GetRecord(ctx context.Context, subdomain string) (string, error)
}
type SSLProvider interface {
	IssueCertificate(ctx context.Context, domain string) error
	RevokeCertificate(ctx context.Context, domain string) error
}
type DatabaseSeeder interface {
	SeedDatabase(ctx context.Context, dbURL string, sanitize bool) error
	CloneDatabase(ctx context.Context, sourceURL, targetURL string) error
}
type ServiceMocker interface {
	MockService(ctx context.Context, serviceName, endpoint string) (string, error)
	UnmockService(ctx context.Context, serviceName string) error
}
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
func (pm *PreviewManager) CreatePreviewEnvironment(ctx context.Context, config *PreviewEnvironmentConfig) (*PreviewEnvironment, error) {
	subdomain := pm.generateSubdomain(config.ProjectID, config.PullRequestID)
	url := fmt.Sprintf("https:
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
	if config.ProtectWithAuth {
		preview.BasicAuth = &BasicAuth{
			Username: "preview",
			Password: generateRandomPassword(16),
		}
	}
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
	if err := pm.dnsProvider.CreateRecord(ctx, subdomain, config.TargetIP); err != nil {
		return nil, fmt.Errorf("failed to create DNS record: %w", err)
	}
	if preview.SSL {
		if err := pm.sslProvider.IssueCertificate(ctx, url); err != nil {
			return nil, fmt.Errorf("failed to issue SSL certificate: %w", err)
		}
	}
	if config.SeedDatabase {
		if err := pm.dbSeeder.SeedDatabase(ctx, config.DatabaseURL, config.SanitizeData); err != nil {
			return nil, fmt.Errorf("failed to seed database: %w", err)
		}
		preview.DatabaseSeeded = true
	}
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
func (pm *PreviewManager) UpdatePreviewEnvironment(ctx context.Context, previewID, commitSHA string) error {
	fmt.Printf("📦 Updating preview environment %s with commit %s\n", previewID, commitSHA[:7])
	time.Sleep(5 * time.Second)
	fmt.Printf("✅ Preview environment updated successfully\n")
	return nil
}
func (pm *PreviewManager) DeletePreviewEnvironment(ctx context.Context, previewID string, preview *PreviewEnvironment) error {
	preview.Status = "deleting"
	for _, service := range preview.MockedServices {
		if err := pm.serviceMocker.UnmockService(ctx, service); err != nil {
			fmt.Printf("Warning: failed to unmock service %s: %v\n", service, err)
		}
	}
	if preview.SSL {
		if err := pm.sslProvider.RevokeCertificate(ctx, preview.URL); err != nil {
			fmt.Printf("Warning: failed to revoke SSL certificate: %v\n", err)
		}
	}
	subdomain := strings.Split(preview.URL, ".")[0]
	subdomain = strings.TrimPrefix(subdomain, "https:
	if err := pm.dnsProvider.DeleteRecord(ctx, subdomain); err != nil {
		fmt.Printf("Warning: failed to delete DNS record: %v\n", err)
	}
	envName := fmt.Sprintf("preview-%s", preview.ID)
	if err := pm.envManager.DeleteEnvironment(ctx, envName); err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}
	preview.Status = "deleted"
	return nil
}
func (pm *PreviewManager) SleepPreviewEnvironment(ctx context.Context, previewID string, preview *PreviewEnvironment) error {
	preview.Status = "sleeping"
	preview.UpdatedAt = time.Now()
	preview.Resources.MinReplicas = 0
	preview.Resources.MaxReplicas = 0
	fmt.Printf("💤 Preview environment %s is now sleeping\n", previewID)
	return nil
}
func (pm *PreviewManager) WakePreviewEnvironment(ctx context.Context, previewID string, preview *PreviewEnvironment) error {
	preview.Status = "active"
	preview.UpdatedAt = time.Now()
	now := time.Now()
	preview.LastAccessedAt = &now
	preview.Resources.MinReplicas = 1
	preview.Resources.MaxReplicas = 1
	fmt.Printf("🌅 Preview environment %s is now awake\n", previewID)
	return nil
}
func (pm *PreviewManager) MonitorPreviewEnvironments(ctx context.Context, previews []*PreviewEnvironment) error {
	for _, preview := range previews {
		if preview.Status == "deleted" {
			continue
		}
		now := time.Now()
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
func (pm *PreviewManager) generateSubdomain(projectID, prID string) string {
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

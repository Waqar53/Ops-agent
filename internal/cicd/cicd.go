package cicd

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// BuildStatus represents the status of a build
type BuildStatus string

const (
	BuildPending  BuildStatus = "pending"
	BuildRunning  BuildStatus = "running"
	BuildSuccess  BuildStatus = "success"
	BuildFailed   BuildStatus = "failed"
	BuildCanceled BuildStatus = "canceled"
)

// Build represents a CI/CD build
type Build struct {
	ID          string                 `json:"id"`
	ProjectID   string                 `json:"project_id"`
	GitCommit   string                 `json:"git_commit"`
	GitBranch   string                 `json:"git_branch"`
	GitAuthor   string                 `json:"git_author"`
	GitMessage  string                 `json:"git_message"`
	Status      BuildStatus            `json:"status"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Duration    *int                   `json:"duration,omitempty"` // seconds
	LogURL      string                 `json:"log_url,omitempty"`
	ArtifactURL string                 `json:"artifact_url,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// PreviewEnvironment represents a preview deployment for a PR
type PreviewEnvironment struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"project_id"`
	PullRequestID string     `json:"pull_request_id"`
	GitBranch     string     `json:"git_branch"`
	URL           string     `json:"url"`
	Status        string     `json:"status"` // creating, ready, destroying, destroyed
	CreatedAt     time.Time  `json:"created_at"`
	DestroyedAt   *time.Time `json:"destroyed_at,omitempty"`
}

// CICDService handles CI/CD operations
type CICDService struct {
	db *sql.DB
}

// NewCICDService creates a new CI/CD service
func NewCICDService(db *sql.DB) *CICDService {
	return &CICDService{db: db}
}

// CreateBuild creates a new build
func (cs *CICDService) CreateBuild(ctx context.Context, build *Build) error {
	metadataJSON, _ := json.Marshal(build.Metadata)

	return cs.db.QueryRowContext(ctx, `
		INSERT INTO builds (project_id, git_commit, git_branch, git_author, git_message, status, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`, build.ProjectID, build.GitCommit, build.GitBranch, build.GitAuthor,
		build.GitMessage, build.Status, metadataJSON).
		Scan(&build.ID, &build.CreatedAt)
}

// StartBuild starts a build process
func (cs *CICDService) StartBuild(ctx context.Context, buildID string) error {
	now := time.Now()
	_, err := cs.db.ExecContext(ctx, `
		UPDATE builds
		SET status = $1, started_at = $2
		WHERE id = $3
	`, BuildRunning, now, buildID)

	if err != nil {
		return err
	}

	// Start build process asynchronously
	go cs.executeBuild(buildID)

	return nil
}

// executeBuild runs the actual build process
func (cs *CICDService) executeBuild(buildID string) {
	ctx := context.Background()

	// Get build details
	var build Build
	var metadataJSON []byte
	err := cs.db.QueryRowContext(ctx, `
		SELECT project_id, git_commit, git_branch, metadata
		FROM builds WHERE id = $1
	`, buildID).Scan(&build.ProjectID, &build.GitCommit, &build.GitBranch, &metadataJSON)

	if err != nil {
		cs.failBuild(buildID, "Failed to get build details")
		return
	}

	json.Unmarshal(metadataJSON, &build.Metadata)

	// Clone repository
	repoURL := build.Metadata["git_repo"].(string)
	workDir := fmt.Sprintf("/tmp/builds/%s", buildID)

	cmd := exec.Command("git", "clone", "--depth", "1", "--branch", build.GitBranch, repoURL, workDir)
	if err := cmd.Run(); err != nil {
		cs.failBuild(buildID, "Failed to clone repository")
		return
	}

	// Build Docker image
	imageName := fmt.Sprintf("%s:%s", build.ProjectID, build.GitCommit[:7])
	cmd = exec.Command("docker", "build", "-t", imageName, workDir)
	output, err := cmd.CombinedOutput()

	if err != nil {
		cs.failBuild(buildID, fmt.Sprintf("Build failed: %s", string(output)))
		return
	}

	// Push to registry
	cmd = exec.Command("docker", "push", imageName)
	if err := cmd.Run(); err != nil {
		cs.failBuild(buildID, "Failed to push image")
		return
	}

	// Mark build as successful
	cs.completeBuild(buildID, BuildSuccess, imageName)
}

// completeBuild marks a build as complete
func (cs *CICDService) completeBuild(buildID string, status BuildStatus, artifactURL string) {
	now := time.Now()
	cs.db.Exec(`
		UPDATE builds
		SET status = $1, completed_at = $2, artifact_url = $3
		WHERE id = $4
	`, status, now, artifactURL, buildID)
}

// failBuild marks a build as failed
func (cs *CICDService) failBuild(buildID string, reason string) {
	now := time.Now()
	metadata := map[string]interface{}{"error": reason}
	metadataJSON, _ := json.Marshal(metadata)

	cs.db.Exec(`
		UPDATE builds
		SET status = $1, completed_at = $2, metadata = $3
		WHERE id = $4
	`, BuildFailed, now, metadataJSON, buildID)
}

// GetBuilds retrieves builds for a project
func (cs *CICDService) GetBuilds(ctx context.Context, projectID string, limit int) ([]Build, error) {
	rows, err := cs.db.QueryContext(ctx, `
		SELECT id, project_id, git_commit, git_branch, git_author, git_message, 
		       status, started_at, completed_at, log_url, artifact_url, metadata, created_at
		FROM builds
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var builds []Build
	for rows.Next() {
		var b Build
		var startedAt, completedAt sql.NullTime
		var logURL, artifactURL sql.NullString
		var metadataJSON []byte

		err := rows.Scan(&b.ID, &b.ProjectID, &b.GitCommit, &b.GitBranch, &b.GitAuthor,
			&b.GitMessage, &b.Status, &startedAt, &completedAt, &logURL, &artifactURL,
			&metadataJSON, &b.CreatedAt)
		if err != nil {
			continue
		}

		if startedAt.Valid {
			b.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			b.CompletedAt = &completedAt.Time
		}
		if logURL.Valid {
			b.LogURL = logURL.String
		}
		if artifactURL.Valid {
			b.ArtifactURL = artifactURL.String
		}
		json.Unmarshal(metadataJSON, &b.Metadata)

		builds = append(builds, b)
	}

	return builds, nil
}

// CreatePreviewEnvironment creates a preview environment for a PR
func (cs *CICDService) CreatePreviewEnvironment(ctx context.Context, projectID, prID, branch string) (*PreviewEnvironment, error) {
	preview := &PreviewEnvironment{
		ProjectID:     projectID,
		PullRequestID: prID,
		GitBranch:     branch,
		URL:           fmt.Sprintf("https://pr-%s.preview.opsagent.dev", prID),
		Status:        "creating",
	}

	err := cs.db.QueryRowContext(ctx, `
		INSERT INTO preview_environments (project_id, pull_request_id, git_branch, url, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, preview.ProjectID, preview.PullRequestID, preview.GitBranch, preview.URL, preview.Status).
		Scan(&preview.ID, &preview.CreatedAt)

	if err != nil {
		return nil, err
	}

	// Deploy preview environment asynchronously
	go cs.deployPreview(preview.ID)

	return preview, nil
}

// deployPreview deploys a preview environment
func (cs *CICDService) deployPreview(previewID string) {
	// TODO: Implement actual preview deployment
	// 1. Build Docker image
	// 2. Deploy to Kubernetes/ECS
	// 3. Configure DNS
	// 4. Update status to "ready"

	time.Sleep(30 * time.Second) // Simulate deployment

	cs.db.Exec(`
		UPDATE preview_environments
		SET status = 'ready'
		WHERE id = $1
	`, previewID)
}

// DestroyPreviewEnvironment destroys a preview environment
func (cs *CICDService) DestroyPreviewEnvironment(ctx context.Context, previewID string) error {
	now := time.Now()
	_, err := cs.db.ExecContext(ctx, `
		UPDATE preview_environments
		SET status = 'destroyed', destroyed_at = $1
		WHERE id = $2
	`, now, previewID)

	// Clean up resources asynchronously
	go cs.cleanupPreview(previewID)

	return err
}

// cleanupPreview cleans up preview environment resources
func (cs *CICDService) cleanupPreview(previewID string) {
	// TODO: Implement actual cleanup
	// 1. Delete Kubernetes/ECS resources
	// 2. Remove DNS records
	// 3. Delete Docker images
}

// HandleWebhook processes GitHub/GitLab webhooks
func (cs *CICDService) HandleWebhook(ctx context.Context, provider string, payload map[string]interface{}) error {
	switch provider {
	case "github":
		return cs.handleGitHubWebhook(ctx, payload)
	case "gitlab":
		return cs.handleGitLabWebhook(ctx, payload)
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}
}

// handleGitHubWebhook processes GitHub webhooks
func (cs *CICDService) handleGitHubWebhook(ctx context.Context, payload map[string]interface{}) error {
	// Parse webhook payload
	eventType := payload["event"].(string)

	switch eventType {
	case "push":
		// Trigger build
		return cs.handlePushEvent(ctx, payload)
	case "pull_request":
		// Create/destroy preview environment
		return cs.handlePullRequestEvent(ctx, payload)
	}

	return nil
}

// handleGitLabWebhook processes GitLab webhooks
func (cs *CICDService) handleGitLabWebhook(ctx context.Context, payload map[string]interface{}) error {
	// Similar to GitHub
	return nil
}

// handlePushEvent handles git push events
func (cs *CICDService) handlePushEvent(ctx context.Context, payload map[string]interface{}) error {
	// Extract commit info and trigger build
	return nil
}

// handlePullRequestEvent handles PR events
func (cs *CICDService) handlePullRequestEvent(ctx context.Context, payload map[string]interface{}) error {
	action := payload["action"].(string)

	if action == "opened" || action == "synchronize" {
		// Create/update preview environment
	} else if action == "closed" {
		// Destroy preview environment
	}

	return nil
}

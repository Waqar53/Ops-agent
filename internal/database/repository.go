package database
import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
)
type ProjectRepository struct {
	db *DB
}
func NewProjectRepository(db *DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}
func (r *ProjectRepository) Create(ctx context.Context, project *Project) error {
	if project.ID == "" {
		project.ID = uuid.New().String()
	}
	metadataJSON, err := json.Marshal(project.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	query := `
		INSERT INTO projects (id, name, slug, description, language, framework, git_repo, git_branch, status, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at
	`
	err = r.db.QueryRowContext(ctx, query,
		project.ID, project.Name, project.Slug, project.Description,
		project.Language, project.Framework, project.GitRepo, project.GitBranch,
		project.Status, metadataJSON,
	).Scan(&project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	return nil
}
func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*Project, error) {
	query := `
		SELECT id, name, slug, description, language, framework, git_repo, git_branch,
		       status, created_at, updated_at, metadata
		FROM projects
		WHERE id = $1
	`
	var project Project
	var metadataJSON []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&project.ID, &project.Name, &project.Slug, &project.Description,
		&project.Language, &project.Framework, &project.GitRepo, &project.GitBranch,
		&project.Status, &project.CreatedAt, &project.UpdatedAt, &metadataJSON,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	if err := json.Unmarshal(metadataJSON, &project.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	return &project, nil
}
func (r *ProjectRepository) List(ctx context.Context) ([]*Project, error) {
	query := `
		SELECT id, name, slug, description, language, framework, git_repo, git_branch,
		       status, created_at, updated_at, metadata
		FROM projects
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()
	var projects []*Project
	for rows.Next() {
		var project Project
		var metadataJSON []byte
		err := rows.Scan(
			&project.ID, &project.Name, &project.Slug, &project.Description,
			&project.Language, &project.Framework, &project.GitRepo, &project.GitBranch,
			&project.Status, &project.CreatedAt, &project.UpdatedAt, &metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		if err := json.Unmarshal(metadataJSON, &project.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		projects = append(projects, &project)
	}
	return projects, nil
}
func (r *ProjectRepository) Update(ctx context.Context, project *Project) error {
	metadataJSON, err := json.Marshal(project.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	query := `
		UPDATE projects
		SET name = $2, description = $3, language = $4, framework = $5,
		    git_repo = $6, git_branch = $7, status = $8, metadata = $9
		WHERE id = $1
		RETURNING updated_at
	`
	err = r.db.QueryRowContext(ctx, query,
		project.ID, project.Name, project.Description, project.Language,
		project.Framework, project.GitRepo, project.GitBranch, project.Status, metadataJSON,
	).Scan(&project.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}
func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM projects WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("project not found")
	}
	return nil
}
type DeploymentRepository struct {
	db *DB
}
func NewDeploymentRepository(db *DB) *DeploymentRepository {
	return &DeploymentRepository{db: db}
}
func (r *DeploymentRepository) Create(ctx context.Context, deployment *Deployment) error {
	if deployment.ID == "" {
		deployment.ID = uuid.New().String()
	}
	metadataJSON, err := json.Marshal(deployment.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	query := `
		INSERT INTO deployments (id, project_id, environment_id, version, git_commit, git_branch,
		                        strategy, status, deployed_by, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING deployed_at
	`
	err = r.db.QueryRowContext(ctx, query,
		deployment.ID, deployment.ProjectID, deployment.EnvironmentID, deployment.Version,
		deployment.GitCommit, deployment.GitBranch, deployment.Strategy, deployment.Status,
		deployment.DeployedBy, metadataJSON,
	).Scan(&deployment.DeployedAt)
	if err != nil {
		return fmt.Errorf("failed to create deployment: %w", err)
	}
	return nil
}
func (r *DeploymentRepository) GetByProjectID(ctx context.Context, projectID string) ([]*Deployment, error) {
	query := `
		SELECT id, project_id, environment_id, version, git_commit, git_branch, strategy,
		       status, deployed_by, deployed_at, completed_at, duration_seconds, metadata
		FROM deployments
		WHERE project_id = $1
		ORDER BY deployed_at DESC
		LIMIT 50
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployments: %w", err)
	}
	defer rows.Close()
	var deployments []*Deployment
	for rows.Next() {
		var deployment Deployment
		var metadataJSON []byte
		err := rows.Scan(
			&deployment.ID, &deployment.ProjectID, &deployment.EnvironmentID, &deployment.Version,
			&deployment.GitCommit, &deployment.GitBranch, &deployment.Strategy, &deployment.Status,
			&deployment.DeployedBy, &deployment.DeployedAt, &deployment.CompletedAt,
			&deployment.DurationSeconds, &metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deployment: %w", err)
		}
		if err := json.Unmarshal(metadataJSON, &deployment.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		deployments = append(deployments, &deployment)
	}
	return deployments, nil
}
func (r *DeploymentRepository) UpdateStatus(ctx context.Context, id, status string) error {
	query := `
		UPDATE deployments
		SET status = $2, completed_at = CASE WHEN $2 IN ('success', 'failed', 'rolled_back') THEN NOW() ELSE completed_at END,
		    duration_seconds = CASE WHEN $2 IN ('success', 'failed', 'rolled_back') THEN EXTRACT(EPOCH FROM (NOW() - deployed_at))::INTEGER ELSE duration_seconds END
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}
	return nil
}
type EnvironmentRepository struct {
	db *DB
}
func NewEnvironmentRepository(db *DB) *EnvironmentRepository {
	return &EnvironmentRepository{db: db}
}
func (r *EnvironmentRepository) GetByProjectID(ctx context.Context, projectID string) ([]*Environment, error) {
	query := `
		SELECT id, project_id, name, type, status, url, config, created_at, updated_at
		FROM environments
		WHERE project_id = $1
		ORDER BY 
			CASE type
				WHEN 'production' THEN 1
				WHEN 'staging' THEN 2
				WHEN 'development' THEN 3
				ELSE 4
			END
	`
	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get environments: %w", err)
	}
	defer rows.Close()
	var environments []*Environment
	for rows.Next() {
		var env Environment
		var configJSON []byte
		err := rows.Scan(
			&env.ID, &env.ProjectID, &env.Name, &env.Type, &env.Status,
			&env.URL, &configJSON, &env.CreatedAt, &env.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan environment: %w", err)
		}
		if err := json.Unmarshal(configJSON, &env.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
		environments = append(environments, &env)
	}
	return environments, nil
}
func (r *EnvironmentRepository) Create(ctx context.Context, env *Environment) error {
	if env.ID == "" {
		env.ID = uuid.New().String()
	}
	configJSON, err := json.Marshal(env.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	query := `
		INSERT INTO environments (id, project_id, name, type, status, url, config)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`
	err = r.db.QueryRowContext(ctx, query,
		env.ID, env.ProjectID, env.Name, env.Type, env.Status, env.URL, configJSON,
	).Scan(&env.CreatedAt, &env.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create environment: %w", err)
	}
	return nil
}

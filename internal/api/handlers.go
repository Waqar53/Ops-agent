package api
import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/opsagent/opsagent/internal/config"
	"github.com/opsagent/opsagent/internal/database"
	"golang.org/x/crypto/bcrypt"
)
type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type AuthResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	User         *User  `json:"user"`
}
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
type Project struct {
	ID              string            `json:"id"`
	OrganizationID  string            `json:"organization_id"`
	Name            string            `json:"name"`
	Slug            string            `json:"slug"`
	Description     string            `json:"description,omitempty"`
	RepositoryURL   string            `json:"repository_url,omitempty"`
	Language        string            `json:"language,omitempty"`
	Framework       string            `json:"framework,omitempty"`
	Config          map[string]any    `json:"config,omitempty"`
	Status          string            `json:"status"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}
type Environment struct {
	ID            string         `json:"id"`
	ProjectID     string         `json:"project_id"`
	Name          string         `json:"name"`
	CloudProvider string         `json:"cloud_provider"`
	Region        string         `json:"region"`
	Status        string         `json:"status"`
	Endpoint      string         `json:"endpoint,omitempty"`
	Config        map[string]any `json:"config,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}
type Deployment struct {
	ID            string         `json:"id"`
	ProjectID     string         `json:"project_id"`
	EnvironmentID string         `json:"environment_id"`
	TriggeredBy   string         `json:"triggered_by,omitempty"`
	GitRef        string         `json:"git_ref,omitempty"`
	GitSHA        string         `json:"git_sha,omitempty"`
	ImageTag      string         `json:"image_tag,omitempty"`
	Strategy      string         `json:"strategy"`
	Status        string         `json:"status"`
	StartedAt     *time.Time     `json:"started_at,omitempty"`
	CompletedAt   *time.Time     `json:"completed_at,omitempty"`
	ErrorMessage  string         `json:"error_message,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}
type CreateProjectRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	RepositoryURL string `json:"repository_url,omitempty"`
}
type DeployRequest struct {
	Environment string `json:"environment"`
	GitRef      string `json:"git_ref,omitempty"`
	Strategy    string `json:"strategy,omitempty"`
}
func handleSignup(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SignupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Email == "" || req.Password == "" {
			writeError(w, http.StatusBadRequest, "email and password required")
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to process password")
			return
		}
		userID := uuid.New().String()
		_, err = db.Exec(`
			INSERT INTO users (id, email, password_hash, name)
			VALUES ($1, $2, $3, $4)
		`, userID, req.Email, string(hash), req.Name)
		if err != nil {
			writeError(w, http.StatusConflict, "email already exists")
			return
		}
		orgID := uuid.New().String()
		orgSlug := generateSlug(req.Name)
		_, err = db.Exec(`
			INSERT INTO organizations (id, name, slug)
			VALUES ($1, $2, $3)
		`, orgID, req.Name+"'s Org", orgSlug)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create organization")
			return
		}
		db.Exec(`
			INSERT INTO organization_members (organization_id, user_id, role)
			VALUES ($1, $2, 'owner')
		`, orgID, userID)
		writeJSON(w, http.StatusCreated, map[string]string{
			"message": "account created successfully",
			"user_id": userID,
		})
	}
}
func handleLogin(db *database.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		var user User
		var passwordHash string
		err := db.QueryRow(`
			SELECT id, email, name, avatar_url, password_hash, created_at
			FROM users WHERE email = $1
		`, req.Email).Scan(&user.ID, &user.Email, &user.Name, &user.AvatarURL, &passwordHash, &user.CreatedAt)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		var orgID string
		db.QueryRow(`
			SELECT organization_id FROM organization_members WHERE user_id = $1 LIMIT 1
		`, user.ID).Scan(&orgID)
		expiresAt := time.Now().Add(cfg.Auth.JWTExpiration)
		claims := Claims{
			UserID:         user.ID,
			OrganizationID: orgID,
			Email:          user.Email,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expiresAt),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(cfg.Auth.JWTSecret))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to generate token")
			return
		}
		writeJSON(w, http.StatusOK, AuthResponse{
			Token:     tokenString,
			ExpiresAt: expiresAt.Unix(),
			User:      &user,
		})
	}
}
func handleRefresh(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}
func handleGetUser(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := getUserID(r)
		var user User
		err := db.QueryRow(`
			SELECT id, email, name, avatar_url, created_at
			FROM users WHERE id = $1
		`, userID).Scan(&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.CreatedAt)
		if err != nil {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeJSON(w, http.StatusOK, user)
	}
}
func handleUpdateUser(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}
func handleListOrganizations(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []Organization{})
	}
}
func handleCreateOrganization(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}
func handleGetOrganization(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}
func handleUpdateOrganization(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}
func handleListProjects(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgID := getOrgID(r)
		rows, err := db.Query(`
			SELECT id, organization_id, name, slug, description, repository_url, 
			       language, framework, status, created_at, updated_at
			FROM projects WHERE organization_id = $1
			ORDER BY updated_at DESC
		`, orgID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to fetch projects")
			return
		}
		defer rows.Close()
		var projects []Project
		for rows.Next() {
			var p Project
			rows.Scan(&p.ID, &p.OrganizationID, &p.Name, &p.Slug, &p.Description,
				&p.RepositoryURL, &p.Language, &p.Framework, &p.Status, &p.CreatedAt, &p.UpdatedAt)
			projects = append(projects, p)
		}
		writeJSON(w, http.StatusOK, projects)
	}
}
func handleCreateProject(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		orgID := getOrgID(r)
		projectID := uuid.New().String()
		slug := generateSlug(req.Name)
		_, err := db.Exec(`
			INSERT INTO projects (id, organization_id, name, slug, description, repository_url)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, projectID, orgID, req.Name, slug, req.Description, req.RepositoryURL)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create project")
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{
			"id":   projectID,
			"slug": slug,
		})
	}
}
func handleGetProject(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "projectId")
		orgID := getOrgID(r)
		var p Project
		err := db.QueryRow(`
			SELECT id, organization_id, name, slug, description, repository_url,
			       language, framework, status, created_at, updated_at
			FROM projects WHERE id = $1 AND organization_id = $2
		`, projectID, orgID).Scan(&p.ID, &p.OrganizationID, &p.Name, &p.Slug, &p.Description,
			&p.RepositoryURL, &p.Language, &p.Framework, &p.Status, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			writeError(w, http.StatusNotFound, "project not found")
			return
		}
		writeJSON(w, http.StatusOK, p)
	}
}
func handleUpdateProject(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}
func handleDeleteProject(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "projectId")
		orgID := getOrgID(r)
		result, err := db.Exec(`
			DELETE FROM projects WHERE id = $1 AND organization_id = $2
		`, projectID, orgID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to delete project")
			return
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			writeError(w, http.StatusNotFound, "project not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
func handleAnalyzeProject(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"language":  "Node.js",
			"framework": "Express",
			"services":  []string{"postgresql"},
			"resources": map[string]string{
				"cpu":    "500m",
				"memory": "512Mi",
			},
		})
	}
}
func handleDeploy(db *database.DB, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req DeployRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		projectID := chi.URLParam(r, "projectId")
		userID := getUserID(r)
		deploymentID := uuid.New().String()
		_, err := db.Exec(`
			INSERT INTO deployments (id, project_id, environment_id, triggered_by, strategy, status, started_at)
			SELECT $1, $2, e.id, $3, $4, 'running', NOW()
			FROM environments e
			WHERE e.project_id = $2 AND e.name = $5
		`, deploymentID, projectID, userID, req.Strategy, req.Environment)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to start deployment")
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]string{
			"deployment_id": deploymentID,
			"status":        "running",
		})
	}
}
func handleListDeployments(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "projectId")
		rows, err := db.Query(`
			SELECT id, project_id, environment_id, triggered_by, git_ref, git_sha,
			       image_tag, strategy, status, started_at, completed_at, created_at
			FROM deployments WHERE project_id = $1
			ORDER BY created_at DESC
			LIMIT 50
		`, projectID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to fetch deployments")
			return
		}
		defer rows.Close()
		var deployments []Deployment
		for rows.Next() {
			var d Deployment
			rows.Scan(&d.ID, &d.ProjectID, &d.EnvironmentID, &d.TriggeredBy, &d.GitRef, &d.GitSHA,
				&d.ImageTag, &d.Strategy, &d.Status, &d.StartedAt, &d.CompletedAt, &d.CreatedAt)
			deployments = append(deployments, d)
		}
		writeJSON(w, http.StatusOK, deployments)
	}
}
func handleGetDeployment(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		deploymentID := chi.URLParam(r, "deploymentId")
		var d Deployment
		err := db.QueryRow(`
			SELECT id, project_id, environment_id, triggered_by, git_ref, git_sha,
			       image_tag, strategy, status, started_at, completed_at, error_message, created_at
			FROM deployments WHERE id = $1
		`, deploymentID).Scan(&d.ID, &d.ProjectID, &d.EnvironmentID, &d.TriggeredBy, &d.GitRef, &d.GitSHA,
			&d.ImageTag, &d.Strategy, &d.Status, &d.StartedAt, &d.CompletedAt, &d.ErrorMessage, &d.CreatedAt)
		if err != nil {
			writeError(w, http.StatusNotFound, "deployment not found")
			return
		}
		writeJSON(w, http.StatusOK, d)
	}
}
func handleRollback(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status": "rolled back",
		})
	}
}
func handleListEnvironments(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "projectId")
		rows, err := db.Query(`
			SELECT id, project_id, name, cloud_provider, region, status, endpoint, created_at, updated_at
			FROM environments WHERE project_id = $1
		`, projectID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to fetch environments")
			return
		}
		defer rows.Close()
		var environments []Environment
		for rows.Next() {
			var e Environment
			rows.Scan(&e.ID, &e.ProjectID, &e.Name, &e.CloudProvider, &e.Region, &e.Status, &e.Endpoint, &e.CreatedAt, &e.UpdatedAt)
			environments = append(environments, e)
		}
		writeJSON(w, http.StatusOK, environments)
	}
}
func handleCreateEnvironment(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}
func handleGetEnvironment(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}
func handleDeleteEnvironment(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}
func handleListSecrets(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []map[string]string{})
	}
}
func handleCreateSecret(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}
func handleDeleteSecret(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}
func handleGetLogs(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, []map[string]any{})
	}
}
func handleGetMetrics(db *database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"cpu_usage":    15.5,
			"memory_usage": 234.0,
			"requests":     1250,
			"error_rate":   0.01,
		})
	}
}
type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Plan      string    `json:"plan"`
	CreatedAt time.Time `json:"created_at"`
}
func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "'", "")
	return slug + "-" + uuid.New().String()[:8]
}

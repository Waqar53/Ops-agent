package api
import (
	"log/slog"
	"net/http"
	"time"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/opsagent/opsagent/internal/config"
	"github.com/opsagent/opsagent/internal/database"
)
func NewRouter(cfg *config.Config, db *database.DB, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http:
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy","version":"0.1.0"}`))
	})
	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Post("/auth/signup", handleSignup(db))
			r.Post("/auth/login", handleLogin(db, cfg))
			r.Post("/auth/refresh", handleRefresh(cfg))
		})
		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware(cfg))
			r.Get("/user", handleGetUser(db))
			r.Patch("/user", handleUpdateUser(db))
			r.Get("/organizations", handleListOrganizations(db))
			r.Post("/organizations", handleCreateOrganization(db))
			r.Get("/organizations/{orgId}", handleGetOrganization(db))
			r.Patch("/organizations/{orgId}", handleUpdateOrganization(db))
			r.Get("/projects", handleListProjects(db))
			r.Post("/projects", handleCreateProject(db))
			r.Get("/projects/{projectId}", handleGetProject(db))
			r.Patch("/projects/{projectId}", handleUpdateProject(db))
			r.Delete("/projects/{projectId}", handleDeleteProject(db))
			r.Post("/projects/{projectId}/analyze", handleAnalyzeProject(db))
			r.Post("/projects/{projectId}/deploy", handleDeploy(db, cfg))
			r.Get("/projects/{projectId}/deployments", handleListDeployments(db))
			r.Get("/projects/{projectId}/deployments/{deploymentId}", handleGetDeployment(db))
			r.Post("/projects/{projectId}/deployments/{deploymentId}/rollback", handleRollback(db))
			r.Get("/projects/{projectId}/environments", handleListEnvironments(db))
			r.Post("/projects/{projectId}/environments", handleCreateEnvironment(db))
			r.Get("/projects/{projectId}/environments/{envName}", handleGetEnvironment(db))
			r.Delete("/projects/{projectId}/environments/{envName}", handleDeleteEnvironment(db))
			r.Get("/projects/{projectId}/environments/{envName}/secrets", handleListSecrets(db))
			r.Post("/projects/{projectId}/environments/{envName}/secrets", handleCreateSecret(db))
			r.Delete("/projects/{projectId}/environments/{envName}/secrets/{key}", handleDeleteSecret(db))
			r.Get("/projects/{projectId}/logs", handleGetLogs(db))
			r.Get("/projects/{projectId}/metrics", handleGetMetrics(db))
		})
	})
	return r
}

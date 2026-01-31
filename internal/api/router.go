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

// NewRouter creates the main API router
func NewRouter(cfg *config.Config, db *database.DB, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://*.opsagent.dev"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy","version":"0.1.0"}`))
	})

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Group(func(r chi.Router) {
			r.Post("/auth/signup", handleSignup(db))
			r.Post("/auth/login", handleLogin(db, cfg))
			r.Post("/auth/refresh", handleRefresh(cfg))
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware(cfg))

			// User
			r.Get("/user", handleGetUser(db))
			r.Patch("/user", handleUpdateUser(db))

			// Organizations
			r.Get("/organizations", handleListOrganizations(db))
			r.Post("/organizations", handleCreateOrganization(db))
			r.Get("/organizations/{orgId}", handleGetOrganization(db))
			r.Patch("/organizations/{orgId}", handleUpdateOrganization(db))

			// Projects
			r.Get("/projects", handleListProjects(db))
			r.Post("/projects", handleCreateProject(db))
			r.Get("/projects/{projectId}", handleGetProject(db))
			r.Patch("/projects/{projectId}", handleUpdateProject(db))
			r.Delete("/projects/{projectId}", handleDeleteProject(db))

			// Analysis
			r.Post("/projects/{projectId}/analyze", handleAnalyzeProject(db))

			// Deployments
			r.Post("/projects/{projectId}/deploy", handleDeploy(db, cfg))
			r.Get("/projects/{projectId}/deployments", handleListDeployments(db))
			r.Get("/projects/{projectId}/deployments/{deploymentId}", handleGetDeployment(db))
			r.Post("/projects/{projectId}/deployments/{deploymentId}/rollback", handleRollback(db))

			// Environments
			r.Get("/projects/{projectId}/environments", handleListEnvironments(db))
			r.Post("/projects/{projectId}/environments", handleCreateEnvironment(db))
			r.Get("/projects/{projectId}/environments/{envName}", handleGetEnvironment(db))
			r.Delete("/projects/{projectId}/environments/{envName}", handleDeleteEnvironment(db))

			// Secrets
			r.Get("/projects/{projectId}/environments/{envName}/secrets", handleListSecrets(db))
			r.Post("/projects/{projectId}/environments/{envName}/secrets", handleCreateSecret(db))
			r.Delete("/projects/{projectId}/environments/{envName}/secrets/{key}", handleDeleteSecret(db))

			// Logs
			r.Get("/projects/{projectId}/logs", handleGetLogs(db))

			// Metrics
			r.Get("/projects/{projectId}/metrics", handleGetMetrics(db))
		})
	})

	return r
}

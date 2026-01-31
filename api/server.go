package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/opsagent/opsagent/api/handlers"
	"github.com/opsagent/opsagent/api/middleware"
	"github.com/opsagent/opsagent/internal/auth"
	"github.com/opsagent/opsagent/internal/database"
)

var (
	projectRepo     *database.ProjectRepository
	deploymentRepo  *database.DeploymentRepository
	environmentRepo *database.EnvironmentRepository
	upgrader        = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func main() {
	// Connect to database
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	projectRepo = database.NewProjectRepository(db)
	deploymentRepo = database.NewDeploymentRepository(db)
	environmentRepo = database.NewEnvironmentRepository(db)

	// Initialize auth service
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-in-production" // Default for development
	}
	authService := auth.NewAuthService(db.DB, jwtSecret) // Use embedded *sql.DB
	authHandlers := handlers.NewAuthHandlers(authService)

	// Setup router
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(corsMiddleware)

	// Auth routes (public)
	api.HandleFunc("/auth/register", authHandlers.Register).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/login", authHandlers.Login).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/logout", authHandlers.Logout).Methods("POST", "OPTIONS")

	// Protected routes
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware(authService))

	// Auth (protected)
	protected.HandleFunc("/auth/me", authHandlers.GetMe).Methods("GET", "OPTIONS")
	protected.HandleFunc("/auth/api-keys", authHandlers.CreateAPIKey).Methods("POST", "OPTIONS")

	// Projects
	api.HandleFunc("/projects", listProjects).Methods("GET", "OPTIONS")
	api.HandleFunc("/projects", createProject).Methods("POST", "OPTIONS")
	api.HandleFunc("/projects/{id}", getProject).Methods("GET", "OPTIONS")
	api.HandleFunc("/projects/{id}", updateProject).Methods("PUT", "OPTIONS")
	api.HandleFunc("/projects/{id}", deleteProject).Methods("DELETE", "OPTIONS")

	// Deployments (protected)
	protected.HandleFunc("/projects/{id}/deployments", getDeployments).Methods("GET", "OPTIONS")
	protected.HandleFunc("/projects/{id}/deploy", deployProject).Methods("POST", "OPTIONS")
	protected.HandleFunc("/deployments/{id}/rollback", rollbackDeployment).Methods("POST", "OPTIONS")

	// Environments (protected)
	protected.HandleFunc("/projects/{id}/environments", getEnvironments).Methods("GET", "OPTIONS")

	// Metrics (protected)
	protected.HandleFunc("/projects/{id}/metrics/{metric}", getMetrics).Methods("GET", "OPTIONS")

	// Logs (protected)
	protected.HandleFunc("/projects/{id}/logs", getLogs).Methods("GET", "OPTIONS")

	// Cost (protected)
	protected.HandleFunc("/projects/{id}/cost", getCost).Methods("GET", "OPTIONS")
	protected.HandleFunc("/projects/{id}/cost/forecast", getCostForecast).Methods("GET", "OPTIONS")

	// WebSocket (protected)
	protected.HandleFunc("/ws", handleWebSocket)

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	}).Methods("GET")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("🚀 OpsAgent API Server listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Project handlers
func listProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := projectRepo.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func createProject(w http.ResponseWriter, r *http.Request) {
	var project database.Project
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := projectRepo.Create(r.Context(), &project); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

func getProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	project, err := projectRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func updateProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var project database.Project
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	project.ID = id
	if err := projectRepo.Update(r.Context(), &project); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func deleteProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := projectRepo.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Deployment handlers
func getDeployments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	deployments, err := deploymentRepo.GetByProjectID(r.Context(), projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployments)
}

func deployProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	var req struct {
		Environment string `json:"environment"`
		Strategy    string `json:"strategy"`
		Branch      string `json:"branch"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create deployment record
	deployment := &database.Deployment{
		ProjectID:  projectID,
		Version:    fmt.Sprintf("v1.0.%d", time.Now().Unix()),
		GitBranch:  req.Branch,
		Strategy:   req.Strategy,
		Status:     "pending",
		DeployedBy: "user@example.com",
		Metadata:   make(map[string]interface{}),
	}

	if err := deploymentRepo.Create(r.Context(), deployment); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: Trigger actual deployment process

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(deployment)
}

func rollbackDeployment(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement rollback logic
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "rolled_back"})
}

// Environment handlers
func getEnvironments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectID := vars["id"]

	environments, err := environmentRepo.GetByProjectID(r.Context(), projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(environments)
}

// Metrics handlers
func getMetrics(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement metrics retrieval from InfluxDB
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]map[string]interface{}{})
}

// Logs handlers
func getLogs(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement logs retrieval from Elasticsearch
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]map[string]interface{}{})
}

// Cost handlers
func getCost(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement cost retrieval
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"current":  1247,
		"forecast": 1580,
	})
}

func getCostForecast(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"forecast": 1580,
		"trend":    "stable",
	})
}

// WebSocket handler
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Println("WebSocket client connected")

	// Handle WebSocket messages
	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// Echo message back (for now)
		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}
}

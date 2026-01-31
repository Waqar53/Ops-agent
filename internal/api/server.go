package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Server represents the API server
type Server struct {
	router   *mux.Router
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
}

// NewServer creates a new API server
func NewServer() *Server {
	s := &Server{
		router:  mux.NewRouter(),
		clients: make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Configure CORS properly in production
			},
		},
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Projects
	s.router.HandleFunc("/api/v1/projects", s.listProjects).Methods("GET")
	s.router.HandleFunc("/api/v1/projects", s.createProject).Methods("POST")
	s.router.HandleFunc("/api/v1/projects/{id}", s.getProject).Methods("GET")
	s.router.HandleFunc("/api/v1/projects/{id}", s.updateProject).Methods("PUT")
	s.router.HandleFunc("/api/v1/projects/{id}", s.deleteProject).Methods("DELETE")

	// Deployments
	s.router.HandleFunc("/api/v1/projects/{id}/deployments", s.listDeployments).Methods("GET")
	s.router.HandleFunc("/api/v1/projects/{id}/deploy", s.deploy).Methods("POST")
	s.router.HandleFunc("/api/v1/deployments/{id}/rollback", s.rollback).Methods("POST")

	// Environments
	s.router.HandleFunc("/api/v1/projects/{id}/environments", s.listEnvironments).Methods("GET")
	s.router.HandleFunc("/api/v1/environments", s.createEnvironment).Methods("POST")

	// Monitoring
	s.router.HandleFunc("/api/v1/projects/{id}/metrics", s.getMetrics).Methods("GET")
	s.router.HandleFunc("/api/v1/projects/{id}/logs", s.getLogs).Methods("GET")

	// Cost
	s.router.HandleFunc("/api/v1/projects/{id}/cost", s.getCostReport).Methods("GET")

	// WebSocket for real-time updates
	s.router.HandleFunc("/api/v1/ws", s.handleWebSocket)
}

// Project handlers
func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	projects := []map[string]interface{}{
		{"id": "proj_1", "name": "my-app", "status": "active"},
		{"id": "proj_2", "name": "api-service", "status": "active"},
	}
	json.NewEncoder(w).Encode(projects)
}

func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	project := map[string]interface{}{
		"id":         "proj_new",
		"name":       req["name"],
		"status":     "creating",
		"created_at": time.Now(),
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

func (s *Server) getProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := map[string]interface{}{
		"id":     vars["id"],
		"name":   "my-app",
		"status": "active",
	}
	json.NewEncoder(w).Encode(project)
}

func (s *Server) updateProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	project := map[string]interface{}{
		"id":         vars["id"],
		"name":       req["name"],
		"updated_at": time.Now(),
	}
	json.NewEncoder(w).Encode(project)
}

func (s *Server) deleteProject(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// Deployment handlers
func (s *Server) listDeployments(w http.ResponseWriter, r *http.Request) {
	deployments := []map[string]interface{}{
		{"id": "deploy_1", "version": "v1.2.3", "status": "success", "deployed_at": time.Now()},
		{"id": "deploy_2", "version": "v1.2.2", "status": "success", "deployed_at": time.Now().Add(-24 * time.Hour)},
	}
	json.NewEncoder(w).Encode(deployments)
}

func (s *Server) deploy(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	deployment := map[string]interface{}{
		"id":          "deploy_new",
		"version":     req["version"],
		"strategy":    req["strategy"],
		"status":      "deploying",
		"deployed_at": time.Now(),
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(deployment)
}

func (s *Server) rollback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	result := map[string]interface{}{
		"id":             vars["id"],
		"status":         "rolled_back",
		"rolled_back_at": time.Now(),
	}
	json.NewEncoder(w).Encode(result)
}

// Environment handlers
func (s *Server) listEnvironments(w http.ResponseWriter, r *http.Request) {
	environments := []map[string]interface{}{
		{"id": "env_1", "name": "production", "type": "production"},
		{"id": "env_2", "name": "staging", "type": "staging"},
	}
	json.NewEncoder(w).Encode(environments)
}

func (s *Server) createEnvironment(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	json.NewDecoder(r.Body).Decode(&req)

	env := map[string]interface{}{
		"id":         "env_new",
		"name":       req["name"],
		"type":       req["type"],
		"created_at": time.Now(),
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(env)
}

// Monitoring handlers
func (s *Server) getMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := map[string]interface{}{
		"cpu_usage":    45.2,
		"memory_usage": 62.8,
		"request_rate": 1250.0,
		"error_rate":   0.02,
		"latency_p99":  245.0,
	}
	json.NewEncoder(w).Encode(metrics)
}

func (s *Server) getLogs(w http.ResponseWriter, r *http.Request) {
	logs := []map[string]interface{}{
		{"timestamp": time.Now(), "level": "INFO", "message": "Request processed"},
		{"timestamp": time.Now(), "level": "ERROR", "message": "Database connection failed"},
	}
	json.NewEncoder(w).Encode(logs)
}

// Cost handlers
func (s *Server) getCostReport(w http.ResponseWriter, r *http.Request) {
	report := map[string]interface{}{
		"total_cost": 1250.50,
		"breakdown": map[string]float64{
			"compute":  500.00,
			"database": 350.00,
			"storage":  150.50,
			"network":  250.00,
		},
		"trend": "increasing",
	}
	json.NewEncoder(w).Encode(report)
}

// WebSocket handler
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	s.clients[conn] = true

	// Send real-time updates
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			update := map[string]interface{}{
				"type":      "metrics",
				"timestamp": time.Now(),
				"data": map[string]float64{
					"cpu":    45.2,
					"memory": 62.8,
				},
			}
			conn.WriteJSON(update)
		}
	}
}

// Start starts the API server
func (s *Server) Start(ctx context.Context, addr string) error {
	srv := &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	return srv.ListenAndServe()
}

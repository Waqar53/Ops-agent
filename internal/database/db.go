package database
import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"
	_ "github.com/lib/pq"
)
type DB struct {
	*sql.DB
}
var db *DB
func Connect() (*DB, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres:
	}
	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	db = &DB{sqlDB}
	return db, nil
}
func GetDB() *DB {
	return db
}
func (db *DB) Close() error {
	return db.DB.Close()
}
type Project struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Slug        string                 `json:"slug"`
	Description string                 `json:"description"`
	Language    string                 `json:"language"`
	Framework   string                 `json:"framework"`
	GitRepo     string                 `json:"gitRepo"`
	GitBranch   string                 `json:"gitBranch"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
	Metadata    map[string]interface{} `json:"metadata"`
}
type Environment struct {
	ID        string                 `json:"id"`
	ProjectID string                 `json:"projectId"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Status    string                 `json:"status"`
	URL       string                 `json:"url"`
	Config    map[string]interface{} `json:"config"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
}
type Deployment struct {
	ID              string                 `json:"id"`
	ProjectID       string                 `json:"projectId"`
	EnvironmentID   string                 `json:"environmentId"`
	Version         string                 `json:"version"`
	GitCommit       string                 `json:"gitCommit"`
	GitBranch       string                 `json:"gitBranch"`
	Strategy        string                 `json:"strategy"`
	Status          string                 `json:"status"`
	DeployedBy      string                 `json:"deployedBy"`
	DeployedAt      time.Time              `json:"deployedAt"`
	CompletedAt     *time.Time             `json:"completedAt,omitempty"`
	DurationSeconds *int                   `json:"durationSeconds,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
}
type Service struct {
	ID        string                 `json:"id"`
	ProjectID string                 `json:"projectId"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Version   string                 `json:"version"`
	Config    map[string]interface{} `json:"config"`
	Status    string                 `json:"status"`
	Endpoint  string                 `json:"endpoint"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
}
type Alert struct {
	ID            string                 `json:"id"`
	ProjectID     string                 `json:"projectId"`
	EnvironmentID *string                `json:"environmentId,omitempty"`
	AlertType     string                 `json:"alertType"`
	Severity      string                 `json:"severity"`
	Title         string                 `json:"title"`
	Message       string                 `json:"message"`
	Status        string                 `json:"status"`
	TriggeredAt   time.Time              `json:"triggeredAt"`
	ResolvedAt    *time.Time             `json:"resolvedAt,omitempty"`
	Metadata      map[string]interface{} `json:"metadata"`
}
type Cost struct {
	ID            string                 `json:"id"`
	ProjectID     string                 `json:"projectId"`
	EnvironmentID *string                `json:"environmentId,omitempty"`
	ResourceID    *string                `json:"resourceId,omitempty"`
	Date          time.Time              `json:"date"`
	Amount        float64                `json:"amount"`
	Currency      string                 `json:"currency"`
	Breakdown     map[string]interface{} `json:"breakdown"`
}

package config
import (
	"fmt"
	"os"
	"time"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Auth     AuthConfig     `yaml:"auth"`
	Cloud    CloudConfig    `yaml:"cloud"`
	Logging  LoggingConfig  `yaml:"logging"`
}
type ServerConfig struct {
	Port            int           `yaml:"port" envconfig:"PORT" default:"8080"`
	Host            string        `yaml:"host" envconfig:"HOST" default:"0.0.0.0"`
	ReadTimeout     time.Duration `yaml:"read_timeout" envconfig:"READ_TIMEOUT" default:"15s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" envconfig:"WRITE_TIMEOUT" default:"15s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" envconfig:"SHUTDOWN_TIMEOUT" default:"30s"`
}
type DatabaseConfig struct {
	Host         string `yaml:"host" envconfig:"DB_HOST" default:"localhost"`
	Port         int    `yaml:"port" envconfig:"DB_PORT" default:"5432"`
	Database     string `yaml:"database" envconfig:"DB_NAME" default:"opsagent"`
	User         string `yaml:"user" envconfig:"DB_USER" default:"opsagent"`
	Password     string `yaml:"password" envconfig:"DB_PASSWORD"`
	SSLMode      string `yaml:"ssl_mode" envconfig:"DB_SSL_MODE" default:"disable"`
	MaxOpenConns int    `yaml:"max_open_conns" envconfig:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns int    `yaml:"max_idle_conns" envconfig:"DB_MAX_IDLE_CONNS" default:"5"`
}
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Database, d.SSLMode,
	)
}
type RedisConfig struct {
	Host     string `yaml:"host" envconfig:"REDIS_HOST" default:"localhost"`
	Port     int    `yaml:"port" envconfig:"REDIS_PORT" default:"6379"`
	Password string `yaml:"password" envconfig:"REDIS_PASSWORD"`
	DB       int    `yaml:"db" envconfig:"REDIS_DB" default:"0"`
}
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}
type AuthConfig struct {
	JWTSecret          string        `yaml:"jwt_secret" envconfig:"JWT_SECRET"`
	JWTExpiration      time.Duration `yaml:"jwt_expiration" envconfig:"JWT_EXPIRATION" default:"24h"`
	RefreshExpiration  time.Duration `yaml:"refresh_expiration" envconfig:"REFRESH_EXPIRATION" default:"168h"`
	BcryptCost         int           `yaml:"bcrypt_cost" envconfig:"BCRYPT_COST" default:"12"`
	OAuthGitHubID      string        `yaml:"oauth_github_id" envconfig:"OAUTH_GITHUB_ID"`
	OAuthGitHubSecret  string        `yaml:"oauth_github_secret" envconfig:"OAUTH_GITHUB_SECRET"`
	OAuthGoogleID      string        `yaml:"oauth_google_id" envconfig:"OAUTH_GOOGLE_ID"`
	OAuthGoogleSecret  string        `yaml:"oauth_google_secret" envconfig:"OAUTH_GOOGLE_SECRET"`
}
type CloudConfig struct {
	DefaultProvider string              `yaml:"default_provider" envconfig:"CLOUD_DEFAULT_PROVIDER" default:"aws"`
	AWS             AWSConfig           `yaml:"aws"`
	GCP             GCPConfig           `yaml:"gcp"`
	Azure           AzureConfig         `yaml:"azure"`
	Terraform       TerraformConfig     `yaml:"terraform"`
}
type AWSConfig struct {
	Region          string `yaml:"region" envconfig:"AWS_REGION" default:"us-east-1"`
	AccessKeyID     string `yaml:"access_key_id" envconfig:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string `yaml:"secret_access_key" envconfig:"AWS_SECRET_ACCESS_KEY"`
}
type GCPConfig struct {
	Project         string `yaml:"project" envconfig:"GCP_PROJECT"`
	Region          string `yaml:"region" envconfig:"GCP_REGION" default:"us-central1"`
	CredentialsFile string `yaml:"credentials_file" envconfig:"GCP_CREDENTIALS_FILE"`
}
type AzureConfig struct {
	SubscriptionID string `yaml:"subscription_id" envconfig:"AZURE_SUBSCRIPTION_ID"`
	TenantID       string `yaml:"tenant_id" envconfig:"AZURE_TENANT_ID"`
	ClientID       string `yaml:"client_id" envconfig:"AZURE_CLIENT_ID"`
	ClientSecret   string `yaml:"client_secret" envconfig:"AZURE_CLIENT_SECRET"`
}
type TerraformConfig struct {
	BinaryPath     string `yaml:"binary_path" envconfig:"TERRAFORM_BINARY_PATH" default:"terraform"`
	StateBucket    string `yaml:"state_bucket" envconfig:"TERRAFORM_STATE_BUCKET"`
	WorkspacePath  string `yaml:"workspace_path" envconfig:"TERRAFORM_WORKSPACE_PATH" default:"/tmp/terraform"`
}
type LoggingConfig struct {
	Level  string `yaml:"level" envconfig:"LOG_LEVEL" default:"info"`
	Format string `yaml:"format" envconfig:"LOG_FORMAT" default:"json"`
}
func Load() (*Config, error) {
	cfg := &Config{}
	cfg.Server.Port = 8080
	cfg.Server.ReadTimeout = 15 * time.Second
	cfg.Server.WriteTimeout = 15 * time.Second
	cfg.Server.ShutdownTimeout = 30 * time.Second
	cfg.Database.Port = 5432
	cfg.Database.MaxOpenConns = 25
	cfg.Database.MaxIdleConns = 5
	cfg.Redis.Port = 6379
	cfg.Auth.JWTExpiration = 24 * time.Hour
	cfg.Auth.RefreshExpiration = 168 * time.Hour
	cfg.Auth.BcryptCost = 12
	configPaths := []string{
		"config.yml",
		"config.yaml",
		"/etc/opsagent/config.yml",
	}
	for _, path := range configPaths {
		if data, err := os.ReadFile(path); err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
			}
			break
		}
	}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("failed to process environment: %w", err)
	}
	if cfg.Auth.JWTSecret == "" {
		cfg.Auth.JWTSecret = "dev-secret-change-in-production"
	}
	return cfg, nil
}

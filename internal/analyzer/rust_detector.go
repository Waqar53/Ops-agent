package analyzer
import (
	"context"
	"os"
	"path/filepath"
	"strings"
)
type RustDetector struct{}
func NewRustDetector() *RustDetector {
	return &RustDetector{}
}
func (d *RustDetector) Detect(ctx context.Context, path string) (*DetectionResult, error) {
	cargoPath := filepath.Join(path, "Cargo.toml")
	if _, err := os.Stat(cargoPath); os.IsNotExist(err) {
		return nil, nil
	}
	content, err := os.ReadFile(cargoPath)
	if err != nil {
		return nil, err
	}
	return &DetectionResult{
		Language:   LanguageRust,
		Confidence: 0.95,
		EntryPoint: d.findEntryPoint(path, string(content)),
		Version:    d.extractRustVersion(string(content)),
	}, nil
}
func (d *RustDetector) findEntryPoint(path string, cargoContent string) string {
	if strings.Contains(cargoContent, "[[bin]]") {
		entries := []string{"src/main.rs", "src/bin/main.rs"}
		for _, entry := range entries {
			if _, err := os.Stat(filepath.Join(path, entry)); err == nil {
				return entry
			}
		}
	}
	if _, err := os.Stat(filepath.Join(path, "src/main.rs")); err == nil {
		return "src/main.rs"
	}
	return "src/lib.rs"
}
func (d *RustDetector) extractRustVersion(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(line, "rust-version") {
			parts := strings.Split(line, "=")
			if len(parts) == 2 {
				return strings.Trim(strings.TrimSpace(parts[1]), "\"")
			}
		}
	}
	return "1.75"
}
func (d *RustDetector) DetectFramework(ctx context.Context, path string) (Framework, float64, error) {
	cargoPath := filepath.Join(path, "Cargo.toml")
	content, err := os.ReadFile(cargoPath)
	if err != nil {
		return FrameworkUnknown, 0, err
	}
	contentStr := string(content)
	frameworks := []struct {
		pattern    string
		framework  Framework
		confidence float64
	}{
		{"actix-web", FrameworkActix, 0.98},
		{"rocket", FrameworkRocket, 0.98},
		{"axum", FrameworkAxum, 0.98},
		{"warp", FrameworkWarp, 0.95},
		{"tide", FrameworkTide, 0.95},
		{"poem", FrameworkPoem, 0.95},
		{"salvo", FrameworkSalvo, 0.95},
	}
	for _, fw := range frameworks {
		if strings.Contains(contentStr, fw.pattern) {
			return fw.framework, fw.confidence, nil
		}
	}
	return FrameworkUnknown, 0.6, nil
}
func (d *RustDetector) DetectServices(ctx context.Context, path string) ([]Service, error) {
	var services []Service
	cargoPath := filepath.Join(path, "Cargo.toml")
	content, err := os.ReadFile(cargoPath)
	if err != nil {
		return services, nil
	}
	contentStr := string(content)
	if strings.Contains(contentStr, "sqlx") || strings.Contains(contentStr, "tokio-postgres") {
		services = append(services, Service{
			Type:    "postgresql",
			Version: "15",
			Reason:  "PostgreSQL driver in Cargo.toml",
		})
	}
	if strings.Contains(contentStr, "mysql_async") || strings.Contains(contentStr, "mysql") {
		services = append(services, Service{
			Type:    "mysql",
			Version: "8",
			Reason:  "MySQL driver in Cargo.toml",
		})
	}
	if strings.Contains(contentStr, "redis") {
		services = append(services, Service{
			Type:    "redis",
			Version: "7",
			Reason:  "Redis client in Cargo.toml",
		})
	}
	if strings.Contains(contentStr, "mongodb") {
		services = append(services, Service{
			Type:    "mongodb",
			Version: "7",
			Reason:  "MongoDB driver in Cargo.toml",
		})
	}
	if strings.Contains(contentStr, "lapin") || strings.Contains(contentStr, "amqprs") {
		services = append(services, Service{
			Type:   "rabbitmq",
			Reason: "RabbitMQ client in Cargo.toml",
		})
	}
	if strings.Contains(contentStr, "rdkafka") {
		services = append(services, Service{
			Type:   "kafka",
			Reason: "Kafka client in Cargo.toml",
		})
	}
	if strings.Contains(contentStr, "aws-sdk-s3") || strings.Contains(contentStr, "rusoto_s3") {
		services = append(services, Service{
			Type:   "s3",
			Reason: "AWS S3 SDK in Cargo.toml",
		})
	}
	return services, nil
}
func (d *RustDetector) ScanSecurity(ctx context.Context, path string) ([]SecurityIssue, error) {
	var issues []SecurityIssue
	srcPath := filepath.Join(path, "src")
	if _, err := os.Stat(srcPath); err == nil {
		filepath.Walk(srcPath, func(filePath string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || !strings.HasSuffix(filePath, ".rs") {
				return nil
			}
			content, err := os.ReadFile(filePath)
			if err != nil {
				return nil
			}
			if strings.Contains(string(content), "unsafe {") {
				issues = append(issues, SecurityIssue{
					Severity:    "medium",
					Type:        "unsafe-code",
					Description: "Unsafe code block detected",
					File:        filePath,
					Suggestion:  "Review unsafe code blocks for memory safety",
				})
			}
			return nil
		})
	}
	gitignorePath := filepath.Join(path, ".gitignore")
	if gitignoreContent, err := os.ReadFile(gitignorePath); err == nil {
		if !strings.Contains(string(gitignoreContent), ".env") {
			issues = append(issues, SecurityIssue{
				Severity:    "high",
				Type:        "exposed-env",
				Description: ".env file may be committed to version control",
				File:        ".gitignore",
				Suggestion:  "Add .env to .gitignore",
			})
		}
	}
	return issues, nil
}
func (d *RustDetector) GetBuildConfig(ctx context.Context, path string, framework Framework) (*BuildConfig, error) {
	config := &BuildConfig{
		BuildCommand: "cargo build --release",
		StartCommand: "./target/release/app",
		Port:         8080,
		HealthCheck:  "/health",
		EnvVars:      map[string]string{"RUST_LOG": "info"},
		BaseImage:    "rust:1.75-alpine",
		MultiStage:   true,
	}
	switch framework {
	case FrameworkActix:
		config.Port = 8080
		config.StartCommand = "./target/release/app"
	case FrameworkRocket:
		config.Port = 8000
		config.EnvVars["ROCKET_ADDRESS"] = "0.0.0.0"
		config.EnvVars["ROCKET_PORT"] = "8000"
	case FrameworkAxum:
		config.Port = 3000
	case FrameworkWarp:
		config.Port = 3030
	}
	config.Dockerfile = d.generateDockerfile(config, framework)
	return config, nil
}
func (d *RustDetector) generateDockerfile(config *BuildConfig, framework Framework) string {
	dockerfile := `# Auto-generated by OpsAgent - Rust Multi-Stage Build
FROM rust:1.75-alpine AS builder
WORKDIR /app
# Install build dependencies
RUN apk add --no-cache musl-dev openssl-dev
# Copy manifests
COPY Cargo.toml Cargo.lock ./
# Build dependencies (cached layer)
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release
RUN rm -rf src
# Copy source code
COPY . .
# Build application
RUN cargo build --release
# Runtime stage
FROM alpine:latest AS runner
WORKDIR /app
# Install runtime dependencies
RUN apk add --no-cache ca-certificates libgcc
# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser
# Copy binary from builder
COPY --from=builder /app/target/release/app /app/app
# Set ownership
RUN chown -R appuser:appuser /app
USER appuser
EXPOSE ` + string(rune(config.Port)) + `
CMD ["./app"]
`
	return dockerfile
}

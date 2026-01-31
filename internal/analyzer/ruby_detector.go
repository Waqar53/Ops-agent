package analyzer
import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
)
type RubyDetector struct{}
func NewRubyDetector() *RubyDetector {
	return &RubyDetector{}
}
func (d *RubyDetector) Detect(ctx context.Context, path string) (*DetectionResult, error) {
	gemfilePath := filepath.Join(path, "Gemfile")
	if _, err := os.Stat(gemfilePath); os.IsNotExist(err) {
		return nil, nil
	}
	return &DetectionResult{
		Language:   LanguageRuby,
		Confidence: 0.9,
		EntryPoint: d.findEntryPoint(path),
		Version:    d.extractRubyVersion(path),
	}, nil
}
func (d *RubyDetector) findEntryPoint(path string) string {
	if _, err := os.Stat(filepath.Join(path, "config/application.rb")); err == nil {
		return "config.ru"
	}
	entries := []string{"config.ru", "app.rb", "server.rb", "main.rb"}
	for _, entry := range entries {
		if _, err := os.Stat(filepath.Join(path, entry)); err == nil {
			return entry
		}
	}
	return "config.ru"
}
func (d *RubyDetector) extractRubyVersion(path string) string {
	versionPath := filepath.Join(path, ".ruby-version")
	if content, err := os.ReadFile(versionPath); err == nil {
		return strings.TrimSpace(string(content))
	}
	gemfilePath := filepath.Join(path, "Gemfile")
	if content, err := os.ReadFile(gemfilePath); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "ruby") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					return strings.Trim(parts[1], "'\"")
				}
			}
		}
	}
	return "3.2"
}
func (d *RubyDetector) DetectFramework(ctx context.Context, path string) (Framework, float64, error) {
	gemfilePath := filepath.Join(path, "Gemfile")
	content, err := os.ReadFile(gemfilePath)
	if err != nil {
		return FrameworkUnknown, 0, err
	}
	contentStr := strings.ToLower(string(content))
	frameworks := []struct {
		pattern    string
		framework  Framework
		confidence float64
	}{
		{"gem 'rails'", FrameworkRails, 0.98},
		{"gem \"rails\"", FrameworkRails, 0.98},
		{"gem 'sinatra'", FrameworkSinatra, 0.95},
		{"gem \"sinatra\"", FrameworkSinatra, 0.95},
		{"gem 'hanami'", FrameworkHanami, 0.95},
		{"gem \"hanami\"", FrameworkHanami, 0.95},
		{"gem 'padrino'", FrameworkPadrino, 0.95},
		{"gem \"padrino\"", FrameworkPadrino, 0.95},
		{"gem 'grape'", FrameworkGrape, 0.90},
		{"gem \"grape\"", FrameworkGrape, 0.90},
	}
	for _, fw := range frameworks {
		if strings.Contains(contentStr, fw.pattern) {
			return fw.framework, fw.confidence, nil
		}
	}
	return FrameworkUnknown, 0.5, nil
}
func (d *RubyDetector) DetectServices(ctx context.Context, path string) ([]Service, error) {
	var services []Service
	gemfilePath := filepath.Join(path, "Gemfile")
	file, err := os.Open(gemfilePath)
	if err != nil {
		return services, nil
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.ToLower(scanner.Text())
		if strings.Contains(line, "pg") || strings.Contains(line, "postgresql") {
			services = append(services, Service{
				Type:    "postgresql",
				Version: "15",
				Reason:  "pg gem in Gemfile",
			})
		}
		if strings.Contains(line, "mysql2") {
			services = append(services, Service{
				Type:    "mysql",
				Version: "8",
				Reason:  "mysql2 gem in Gemfile",
			})
		}
		if strings.Contains(line, "redis") {
			services = append(services, Service{
				Type:    "redis",
				Version: "7",
				Reason:  "redis gem in Gemfile",
			})
		}
		if strings.Contains(line, "mongoid") || strings.Contains(line, "mongo") {
			services = append(services, Service{
				Type:    "mongodb",
				Version: "7",
				Reason:  "MongoDB gem in Gemfile",
			})
		}
		if strings.Contains(line, "sidekiq") {
			services = append(services, Service{
				Type:   "sidekiq-worker",
				Reason: "Sidekiq gem in Gemfile",
			})
		}
		if strings.Contains(line, "resque") {
			services = append(services, Service{
				Type:   "resque-worker",
				Reason: "Resque gem in Gemfile",
			})
		}
		if strings.Contains(line, "elasticsearch") {
			services = append(services, Service{
				Type:    "elasticsearch",
				Version: "8",
				Reason:  "Elasticsearch gem in Gemfile",
			})
		}
		if strings.Contains(line, "aws-sdk-s3") {
			services = append(services, Service{
				Type:   "s3",
				Reason: "AWS S3 SDK in Gemfile",
			})
		}
	}
	return services, nil
}
func (d *RubyDetector) ScanSecurity(ctx context.Context, path string) ([]SecurityIssue, error) {
	var issues []SecurityIssue
	configPath := filepath.Join(path, "config/secrets.yml")
	if _, err := os.Stat(configPath); err == nil {
		issues = append(issues, SecurityIssue{
			Severity:    "high",
			Type:        "secrets-file",
			Description: "secrets.yml file detected - ensure it's not committed",
			File:        "config/secrets.yml",
			Suggestion:  "Use environment variables or Rails credentials instead",
		})
	}
	dbConfigPath := filepath.Join(path, "config/database.yml")
	if content, err := os.ReadFile(dbConfigPath); err == nil {
		if strings.Contains(string(content), "password:") && !strings.Contains(string(content), "ENV") {
			issues = append(issues, SecurityIssue{
				Severity:    "critical",
				Type:        "hardcoded-credentials",
				Description: "Hardcoded database credentials in database.yml",
				File:        "config/database.yml",
				Suggestion:  "Use environment variables for database credentials",
			})
		}
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
	secretsPath := filepath.Join(path, "config/secrets.yml")
	if content, err := os.ReadFile(secretsPath); err == nil {
		if strings.Contains(string(content), "secret_key_base:") && !strings.Contains(string(content), "ENV") {
			issues = append(issues, SecurityIssue{
				Severity:    "critical",
				Type:        "hardcoded-secret-key",
				Description: "Hardcoded secret_key_base in secrets.yml",
				File:        "config/secrets.yml",
				Suggestion:  "Use environment variables for secret_key_base",
			})
		}
	}
	return issues, nil
}
func (d *RubyDetector) GetBuildConfig(ctx context.Context, path string, framework Framework) (*BuildConfig, error) {
	config := &BuildConfig{
		BuildCommand: "bundle install --without development test",
		Port:         3000,
		HealthCheck:  "/health",
		EnvVars: map[string]string{
			"RAILS_ENV":                "production",
			"RACK_ENV":                 "production",
			"RAILS_SERVE_STATIC_FILES": "true",
		},
		BaseImage:  "ruby:3.2-alpine",
		MultiStage: true,
	}
	switch framework {
	case FrameworkRails:
		config.BuildCommand = "bundle install --without development test && bundle exec rake assets:precompile"
		config.StartCommand = "bundle exec rails server -b 0.0.0.0 -p 3000"
		config.Port = 3000
		config.EnvVars["RAILS_LOG_TO_STDOUT"] = "true"
	case FrameworkSinatra:
		config.BuildCommand = "bundle install --without development test"
		config.StartCommand = "bundle exec rackup -o 0.0.0.0 -p 4567"
		config.Port = 4567
	case FrameworkHanami:
		config.StartCommand = "bundle exec hanami server"
		config.Port = 2300
	case FrameworkPadrino:
		config.StartCommand = "bundle exec padrino start -h 0.0.0.0 -p 3000"
		config.Port = 3000
	default:
		config.StartCommand = "bundle exec rackup -o 0.0.0.0 -p 9292"
		config.Port = 9292
	}
	config.Dockerfile = d.generateDockerfile(config, framework)
	return config, nil
}
func (d *RubyDetector) generateDockerfile(config *BuildConfig, framework Framework) string {
	dockerfile := `# Auto-generated by OpsAgent - Ruby Multi-Stage Build
FROM ruby:3.2-alpine AS builder
WORKDIR /app
# Install build dependencies
RUN apk add --no-cache build-base postgresql-dev nodejs yarn tzdata
# Copy Gemfile
COPY Gemfile Gemfile.lock ./
# Install gems
RUN bundle config set --local deployment 'true' && \
    bundle config set --local without 'development test' && \
    bundle install -j4
# Copy application code
COPY . .
# Precompile assets (Rails)
RUN if [ -f "bin/rails" ]; then bundle exec rake assets:precompile; fi
# Runtime stage
FROM ruby:3.2-alpine AS runner
WORKDIR /app
# Install runtime dependencies
RUN apk add --no-cache postgresql-client tzdata
# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser
# Copy gems and app from builder
COPY --from=builder /usr/local/bundle /usr/local/bundle
COPY --from=builder /app /app
# Set ownership
RUN chown -R appuser:appuser /app
USER appuser
EXPOSE ` + string(rune(config.Port)) + `
CMD ["` + config.StartCommand + `"]
`
	return dockerfile
}

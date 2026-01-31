package analyzer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// PHPDetector detects PHP projects
type PHPDetector struct{}

func NewPHPDetector() *PHPDetector {
	return &PHPDetector{}
}

func (d *PHPDetector) Detect(ctx context.Context, path string) (*DetectionResult, error) {
	composerPath := filepath.Join(path, "composer.json")
	if _, err := os.Stat(composerPath); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(composerPath)
	if err != nil {
		return nil, err
	}

	var composer struct {
		Require map[string]string `json:"require"`
		Config  struct {
			Platform struct {
				PHP string `json:"php"`
			} `json:"platform"`
		} `json:"config"`
	}
	json.Unmarshal(data, &composer)

	version := "8.2" // Default
	if phpVer, ok := composer.Require["php"]; ok {
		version = strings.TrimPrefix(phpVer, "^")
		version = strings.TrimPrefix(version, "~")
	}

	return &DetectionResult{
		Language:   LanguagePHP,
		Confidence: 0.9,
		EntryPoint: d.findEntryPoint(path),
		Version:    version,
	}, nil
}

func (d *PHPDetector) findEntryPoint(path string) string {
	// Laravel
	if _, err := os.Stat(filepath.Join(path, "artisan")); err == nil {
		return "public/index.php"
	}

	// Symfony
	if _, err := os.Stat(filepath.Join(path, "bin/console")); err == nil {
		return "public/index.php"
	}

	// Common entry points
	entries := []string{"public/index.php", "index.php", "app.php", "web/index.php"}
	for _, entry := range entries {
		if _, err := os.Stat(filepath.Join(path, entry)); err == nil {
			return entry
		}
	}

	return "index.php"
}

func (d *PHPDetector) DetectFramework(ctx context.Context, path string) (Framework, float64, error) {
	composerPath := filepath.Join(path, "composer.json")
	data, err := os.ReadFile(composerPath)
	if err != nil {
		return FrameworkUnknown, 0, err
	}

	var composer struct {
		Require map[string]string `json:"require"`
	}
	json.Unmarshal(data, &composer)

	frameworks := []struct {
		pkg        string
		framework  Framework
		confidence float64
	}{
		{"laravel/framework", FrameworkLaravel, 0.98},
		{"symfony/symfony", FrameworkSymfony, 0.98},
		{"symfony/framework-bundle", FrameworkSymfony, 0.98},
		{"codeigniter4/framework", FrameworkCodeIgniter, 0.95},
		{"slim/slim", FrameworkSlim, 0.95},
		{"laravel/lumen-framework", FrameworkLumen, 0.95},
		{"cakephp/cakephp", FrameworkCakePHP, 0.95},
		{"yiisoft/yii2", FrameworkYii, 0.95},
		{"laminas/laminas-mvc", FrameworkLaminas, 0.90},
	}

	for _, fw := range frameworks {
		if _, ok := composer.Require[fw.pkg]; ok {
			return fw.framework, fw.confidence, nil
		}
	}

	return FrameworkUnknown, 0.5, nil
}

func (d *PHPDetector) DetectServices(ctx context.Context, path string) ([]Service, error) {
	var services []Service
	composerPath := filepath.Join(path, "composer.json")

	data, err := os.ReadFile(composerPath)
	if err != nil {
		return services, nil
	}

	var composer struct {
		Require map[string]string `json:"require"`
	}
	json.Unmarshal(data, &composer)

	// Database detection
	dbPackages := map[string]Service{
		"doctrine/dbal":               {Type: "postgresql", Version: "15", Reason: "Doctrine DBAL in composer.json"},
		"illuminate/database":         {Type: "mysql", Version: "8", Reason: "Laravel database in composer.json"},
		"mongodb/mongodb":             {Type: "mongodb", Version: "7", Reason: "MongoDB driver in composer.json"},
		"predis/predis":               {Type: "redis", Version: "7", Reason: "Predis in composer.json"},
		"phpredis/phpredis":           {Type: "redis", Version: "7", Reason: "PhpRedis in composer.json"},
		"elasticsearch/elasticsearch": {Type: "elasticsearch", Version: "8", Reason: "Elasticsearch client in composer.json"},
	}

	for pkg, svc := range dbPackages {
		if _, ok := composer.Require[pkg]; ok {
			services = append(services, svc)
		}
	}

	// Queue systems
	if _, ok := composer.Require["php-amqplib/php-amqplib"]; ok {
		services = append(services, Service{
			Type:   "rabbitmq",
			Reason: "AMQP library in composer.json",
		})
	}

	// Object storage
	if _, ok := composer.Require["aws/aws-sdk-php"]; ok {
		services = append(services, Service{
			Type:   "s3",
			Reason: "AWS SDK in composer.json",
		})
	}

	return services, nil
}

func (d *PHPDetector) ScanSecurity(ctx context.Context, path string) ([]SecurityIssue, error) {
	var issues []SecurityIssue

	// Check for .env in .gitignore
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

	// Check for debug mode in Laravel
	envPath := filepath.Join(path, ".env")
	if content, err := os.ReadFile(envPath); err == nil {
		if strings.Contains(string(content), "APP_DEBUG=true") {
			issues = append(issues, SecurityIssue{
				Severity:    "critical",
				Type:        "debug-enabled",
				Description: "APP_DEBUG=true in .env file",
				File:        ".env",
				Suggestion:  "Set APP_DEBUG=false for production",
			})
		}
	}

	// Check for hardcoded database credentials
	configFiles := []string{"config/database.php", ".env", "config.php"}
	for _, configFile := range configFiles {
		configPath := filepath.Join(path, configFile)
		if content, err := os.ReadFile(configPath); err == nil {
			contentStr := string(content)
			if strings.Contains(contentStr, "password") &&
				(strings.Contains(contentStr, "root") || strings.Contains(contentStr, "admin")) &&
				!strings.Contains(contentStr, "env(") {
				issues = append(issues, SecurityIssue{
					Severity:    "critical",
					Type:        "hardcoded-credentials",
					Description: "Potential hardcoded database credentials",
					File:        configFile,
					Suggestion:  "Use environment variables for credentials",
				})
			}
		}
	}

	return issues, nil
}

func (d *PHPDetector) GetBuildConfig(ctx context.Context, path string, framework Framework) (*BuildConfig, error) {
	config := &BuildConfig{
		BuildCommand: "composer install --no-dev --optimize-autoloader",
		Port:         8000,
		HealthCheck:  "/health",
		EnvVars: map[string]string{
			"APP_ENV": "production",
		},
		BaseImage:  "php:8.2-fpm-alpine",
		MultiStage: true,
	}

	switch framework {
	case FrameworkLaravel:
		config.BuildCommand = "composer install --no-dev --optimize-autoloader && php artisan config:cache && php artisan route:cache && php artisan view:cache"
		config.StartCommand = "php artisan serve --host=0.0.0.0 --port=8000"
		config.Port = 8000
		config.EnvVars["APP_ENV"] = "production"
		config.EnvVars["APP_DEBUG"] = "false"
	case FrameworkSymfony:
		config.BuildCommand = "composer install --no-dev --optimize-autoloader && php bin/console cache:clear --env=prod"
		config.StartCommand = "php -S 0.0.0.0:8000 -t public"
		config.Port = 8000
		config.EnvVars["APP_ENV"] = "prod"
	case FrameworkCodeIgniter:
		config.StartCommand = "php spark serve --host=0.0.0.0 --port=8080"
		config.Port = 8080
	case FrameworkSlim:
		config.StartCommand = "php -S 0.0.0.0:8080 -t public"
		config.Port = 8080
	default:
		config.StartCommand = "php -S 0.0.0.0:8000"
		config.Port = 8000
	}

	config.Dockerfile = d.generateDockerfile(config, framework)

	return config, nil
}

func (d *PHPDetector) generateDockerfile(config *BuildConfig, framework Framework) string {
	dockerfile := `# Auto-generated by OpsAgent - PHP Multi-Stage Build
FROM php:8.2-fpm-alpine AS builder
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache \
    git \
    unzip \
    libzip-dev \
    postgresql-dev \
    && docker-php-ext-install pdo pdo_mysql pdo_pgsql zip

# Install Composer
COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

# Copy composer files
COPY composer.json composer.lock ./

# Install dependencies
RUN composer install --no-dev --optimize-autoloader --no-scripts

# Copy application code
COPY . .

# Run post-install scripts
RUN composer run-script post-install-cmd --no-interaction || true

# Runtime stage
FROM php:8.2-fpm-alpine AS runner
WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache \
    libzip \
    postgresql-libs \
    nginx \
    && docker-php-ext-install pdo pdo_mysql pdo_pgsql zip opcache

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Copy app from builder
COPY --from=builder /app /app

# Set ownership
RUN chown -R appuser:appuser /app

USER appuser

EXPOSE ` + string(rune(config.Port)) + `

CMD ["` + config.StartCommand + `"]
`
	return dockerfile
}

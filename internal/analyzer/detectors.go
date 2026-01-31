package analyzer

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// NodeDetector detects Node.js projects
type NodeDetector struct{}

func NewNodeDetector() *NodeDetector {
	return &NodeDetector{}
}

func (d *NodeDetector) Detect(ctx context.Context, path string) (*DetectionResult, error) {
	pkgPath := filepath.Join(path, "package.json")
	if _, err := os.Stat(pkgPath); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil, err
	}

	var pkg struct {
		Main    string            `json:"main"`
		Scripts map[string]string `json:"scripts"`
		Engines struct {
			Node string `json:"node"`
		} `json:"engines"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	result := &DetectionResult{
		Language:   LanguageNodeJS,
		Confidence: 0.9,
		EntryPoint: pkg.Main,
		Version:    pkg.Engines.Node,
	}

	// Check for common entry points
	if result.EntryPoint == "" {
		for _, entry := range []string{"server.js", "app.js", "index.js", "src/index.js", "dist/index.js"} {
			if _, err := os.Stat(filepath.Join(path, entry)); err == nil {
				result.EntryPoint = entry
				break
			}
		}
	}

	return result, nil
}

func (d *NodeDetector) DetectFramework(ctx context.Context, path string) (Framework, float64, error) {
	pkgPath := filepath.Join(path, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return FrameworkUnknown, 0, err
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return FrameworkUnknown, 0, err
	}

	allDeps := make(map[string]string)
	for k, v := range pkg.Dependencies {
		allDeps[k] = v
	}
	for k, v := range pkg.DevDependencies {
		allDeps[k] = v
	}

	// Framework detection with confidence
	frameworks := []struct {
		pkg        string
		framework  Framework
		confidence float64
	}{
		{"next", FrameworkNextJS, 0.98},
		{"@nestjs/core", FrameworkNestJS, 0.98},
		{"fastify", FrameworkFastify, 0.95},
		{"koa", FrameworkKoa, 0.95},
		{"@hapi/hapi", FrameworkHapi, 0.95},
		{"@sveltejs/kit", FrameworkSvelteKit, 0.98},
		{"@remix-run/react", FrameworkRemix, 0.98},
		{"astro", FrameworkAstro, 0.98},
		{"express", FrameworkExpress, 0.90},
	}

	for _, fw := range frameworks {
		if _, ok := allDeps[fw.pkg]; ok {
			return fw.framework, fw.confidence, nil
		}
	}

	return FrameworkUnknown, 0.5, nil
}

func (d *NodeDetector) DetectServices(ctx context.Context, path string) ([]Service, error) {
	pkgPath := filepath.Join(path, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil, err
	}

	var pkg struct {
		Dependencies map[string]string `json:"dependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	var services []Service

	// Database detection
	dbDetectors := []struct {
		packages []string
		service  string
		version  string
	}{
		{[]string{"pg", "postgres", "pg-pool", "pg-promise"}, "postgresql", "15"},
		{[]string{"mysql", "mysql2"}, "mysql", "8"},
		{[]string{"mongodb", "mongoose", "mongodb-memory-server"}, "mongodb", "7"},
		{[]string{"redis", "ioredis", "redis-om"}, "redis", "7"},
		{[]string{"@elastic/elasticsearch", "elasticsearch"}, "elasticsearch", "8"},
		{[]string{"kafkajs", "kafka-node"}, "kafka", "latest"},
		{[]string{"amqplib", "rabbitmq-client"}, "rabbitmq", "3"},
		{[]string{"@aws-sdk/client-s3", "aws-sdk"}, "s3", ""},
		{[]string{"@prisma/client"}, "prisma-orm", ""},
		{[]string{"typeorm"}, "typeorm", ""},
		{[]string{"sequelize"}, "sequelize-orm", ""},
	}

	for _, detector := range dbDetectors {
		for _, pkgName := range detector.packages {
			if version, ok := pkg.Dependencies[pkgName]; ok {
				services = append(services, Service{
					Type:     detector.service,
					Version:  detector.version,
					Reason:   pkgName + " package in package.json",
					Required: true,
				})
				_ = version
				break
			}
		}
	}

	// Check for .env files to detect more services
	envPath := filepath.Join(path, ".env.example")
	if _, err := os.Stat(envPath); err == nil {
		envContent, _ := os.ReadFile(envPath)
		envStr := string(envContent)
		
		if strings.Contains(envStr, "DATABASE_URL") && !hasService(services, "postgresql") {
			services = append(services, Service{
				Type:   "postgresql",
				Reason: "DATABASE_URL in .env.example",
			})
		}
		if strings.Contains(envStr, "REDIS") && !hasService(services, "redis") {
			services = append(services, Service{
				Type:   "redis",
				Reason: "REDIS config in .env.example",
			})
		}
		if strings.Contains(envStr, "STRIPE") {
			services = append(services, Service{
				Type:   "stripe-payments",
				Reason: "STRIPE config detected",
			})
		}
		if strings.Contains(envStr, "OPENAI") || strings.Contains(envStr, "ANTHROPIC") {
			services = append(services, Service{
				Type:   "ai-api",
				Reason: "AI API keys detected",
			})
		}
	}

	return services, nil
}

func (d *NodeDetector) ScanSecurity(ctx context.Context, path string) ([]SecurityIssue, error) {
	var issues []SecurityIssue

	// Check package.json for known vulnerable packages
	pkgPath := filepath.Join(path, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return issues, nil
	}

	var pkg struct {
		Dependencies map[string]string `json:"dependencies"`
	}
	json.Unmarshal(data, &pkg)

	// Check for hardcoded secrets in common files
	secretPatterns := []string{
		"password",
		"secret",
		"api_key",
		"apikey",
		"private_key",
		"token",
	}

	filesToCheck := []string{
		".env",
		"config.js",
		"config.json",
		"src/config.js",
	}

	for _, file := range filesToCheck {
		filePath := filepath.Join(path, file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		
		contentLower := strings.ToLower(string(content))
		for _, pattern := range secretPatterns {
			if strings.Contains(contentLower, pattern) && !strings.Contains(file, ".example") {
				issues = append(issues, SecurityIssue{
					Severity:    "high",
					Type:        "hardcoded-secret",
					Description: "Potential secret found in " + file,
					File:        file,
					Suggestion:  "Move sensitive values to environment variables",
				})
				break
			}
		}
	}

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

	// Check for outdated dependencies
	outdatedPackages := map[string]string{
		"express": "4.17.0", // example minimum safe version
	}
	for pkgName, minVersion := range outdatedPackages {
		if version, ok := pkg.Dependencies[pkgName]; ok {
			_ = version
			_ = minVersion
			// In production, compare versions properly
		}
	}

	return issues, nil
}

func (d *NodeDetector) GetBuildConfig(ctx context.Context, path string, framework Framework) (*BuildConfig, error) {
	config := &BuildConfig{
		BuildCommand: "npm run build",
		StartCommand: "npm start",
		Port:         3000,
		HealthCheck:  "/health",
		EnvVars:      map[string]string{"NODE_ENV": "production"},
		BaseImage:    "node:20-alpine",
		MultiStage:   true,
	}

	// Framework-specific configurations
	switch framework {
	case FrameworkNextJS:
		config.StartCommand = "npm run start"
		config.Port = 3000
		config.HealthCheck = "/"
		config.BaseImage = "node:20-alpine"
	case FrameworkNestJS:
		config.BuildCommand = "npm run build"
		config.StartCommand = "node dist/main"
		config.Port = 3000
	case FrameworkExpress:
		config.BuildCommand = ""
		config.StartCommand = "node server.js"
		config.Port = 3000
	case FrameworkFastify:
		config.Port = 3000
	}

	// Check for custom scripts in package.json
	pkgPath := filepath.Join(path, "package.json")
	if data, err := os.ReadFile(pkgPath); err == nil {
		var pkg struct {
			Scripts map[string]string `json:"scripts"`
		}
		if json.Unmarshal(data, &pkg) == nil {
			if _, ok := pkg.Scripts["build"]; !ok {
				config.BuildCommand = ""
			}
			if start, ok := pkg.Scripts["start"]; ok {
				config.StartCommand = "npm start"
				_ = start
			}
		}
	}

	// Generate Dockerfile
	config.Dockerfile = d.generateDockerfile(config, framework)

	return config, nil
}

func (d *NodeDetector) generateDockerfile(config *BuildConfig, framework Framework) string {
	dockerfile := `# Auto-generated by OpsAgent
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production

COPY --from=builder /app/node_modules ./node_modules
COPY . .

EXPOSE ` + string(rune(config.Port)) + `
CMD ["` + config.StartCommand + `"]
`
	return dockerfile
}

// PythonDetector detects Python projects
type PythonDetector struct{}

func NewPythonDetector() *PythonDetector {
	return &PythonDetector{}
}

func (d *PythonDetector) Detect(ctx context.Context, path string) (*DetectionResult, error) {
	// Check for requirements.txt, pyproject.toml, or setup.py
	indicators := []string{"requirements.txt", "pyproject.toml", "setup.py", "Pipfile"}
	
	for _, indicator := range indicators {
		if _, err := os.Stat(filepath.Join(path, indicator)); err == nil {
			return &DetectionResult{
				Language:   LanguagePython,
				Confidence: 0.9,
				EntryPoint: d.findEntryPoint(path),
			}, nil
		}
	}
	
	return nil, nil
}

func (d *PythonDetector) findEntryPoint(path string) string {
	entries := []string{"main.py", "app.py", "run.py", "server.py", "manage.py", "wsgi.py"}
	for _, entry := range entries {
		if _, err := os.Stat(filepath.Join(path, entry)); err == nil {
			return entry
		}
	}
	return "app.py"
}

func (d *PythonDetector) DetectFramework(ctx context.Context, path string) (Framework, float64, error) {
	reqPath := filepath.Join(path, "requirements.txt")
	pyprojectPath := filepath.Join(path, "pyproject.toml")
	
	var content string
	if data, err := os.ReadFile(reqPath); err == nil {
		content = string(data)
	} else if data, err := os.ReadFile(pyprojectPath); err == nil {
		content = string(data)
	}
	
	contentLower := strings.ToLower(content)
	
	frameworks := []struct {
		pattern    string
		framework  Framework
		confidence float64
	}{
		{"fastapi", FrameworkFastAPI, 0.98},
		{"django", FrameworkDjango, 0.98},
		{"flask", FrameworkFlask, 0.95},
		{"starlette", FrameworkStarlette, 0.90},
		{"sanic", FrameworkSanic, 0.90},
		{"aiohttp", FrameworkAiohttp, 0.90},
	}
	
	for _, fw := range frameworks {
		if strings.Contains(contentLower, fw.pattern) {
			return fw.framework, fw.confidence, nil
		}
	}
	
	return FrameworkUnknown, 0.5, nil
}

func (d *PythonDetector) DetectServices(ctx context.Context, path string) ([]Service, error) {
	var services []Service
	reqPath := filepath.Join(path, "requirements.txt")
	
	file, err := os.Open(reqPath)
	if err != nil {
		return services, nil
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.ToLower(scanner.Text())
		
		if strings.Contains(line, "psycopg") || strings.Contains(line, "asyncpg") {
			services = append(services, Service{Type: "postgresql", Version: "15", Reason: "psycopg in requirements.txt"})
		}
		if strings.Contains(line, "pymysql") || strings.Contains(line, "mysqlclient") {
			services = append(services, Service{Type: "mysql", Version: "8", Reason: "mysql driver in requirements.txt"})
		}
		if strings.Contains(line, "redis") {
			services = append(services, Service{Type: "redis", Version: "7", Reason: "redis in requirements.txt"})
		}
		if strings.Contains(line, "pymongo") {
			services = append(services, Service{Type: "mongodb", Version: "7", Reason: "pymongo in requirements.txt"})
		}
		if strings.Contains(line, "celery") {
			services = append(services, Service{Type: "celery-worker", Reason: "celery in requirements.txt"})
		}
		if strings.Contains(line, "boto3") {
			services = append(services, Service{Type: "aws-s3", Reason: "boto3 in requirements.txt"})
		}
	}
	
	return services, nil
}

func (d *PythonDetector) ScanSecurity(ctx context.Context, path string) ([]SecurityIssue, error) {
	var issues []SecurityIssue
	
	// Check for DEBUG=True in Django settings
	settingsPath := filepath.Join(path, "settings.py")
	if content, err := os.ReadFile(settingsPath); err == nil {
		if strings.Contains(string(content), "DEBUG = True") {
			issues = append(issues, SecurityIssue{
				Severity:    "critical",
				Type:        "debug-enabled",
				Description: "DEBUG=True in production settings",
				File:        "settings.py",
				Suggestion:  "Set DEBUG=False for production",
			})
		}
	}
	
	return issues, nil
}

func (d *PythonDetector) GetBuildConfig(ctx context.Context, path string, framework Framework) (*BuildConfig, error) {
	config := &BuildConfig{
		BuildCommand: "pip install -r requirements.txt",
		Port:         8000,
		HealthCheck:  "/health",
		EnvVars:      map[string]string{"PYTHONUNBUFFERED": "1"},
		BaseImage:    "python:3.11-slim",
	}
	
	switch framework {
	case FrameworkFastAPI:
		config.StartCommand = "uvicorn main:app --host 0.0.0.0 --port 8000"
	case FrameworkDjango:
		config.StartCommand = "gunicorn app.wsgi:application --bind 0.0.0.0:8000"
		config.Port = 8000
	case FrameworkFlask:
		config.StartCommand = "gunicorn app:app --bind 0.0.0.0:5000"
		config.Port = 5000
	}
	
	return config, nil
}

// GoDetector detects Go projects
type GoDetector struct{}

func NewGoDetector() *GoDetector {
	return &GoDetector{}
}

func (d *GoDetector) Detect(ctx context.Context, path string) (*DetectionResult, error) {
	modPath := filepath.Join(path, "go.mod")
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		return nil, nil
	}
	
	return &DetectionResult{
		Language:   LanguageGo,
		Confidence: 0.95,
		EntryPoint: d.findEntryPoint(path),
	}, nil
}

func (d *GoDetector) findEntryPoint(path string) string {
	entries := []string{"main.go", "cmd/main.go", "cmd/server/main.go", "cmd/api/main.go"}
	for _, entry := range entries {
		if _, err := os.Stat(filepath.Join(path, entry)); err == nil {
			return entry
		}
	}
	return "main.go"
}

func (d *GoDetector) DetectFramework(ctx context.Context, path string) (Framework, float64, error) {
	modPath := filepath.Join(path, "go.mod")
	content, err := os.ReadFile(modPath)
	if err != nil {
		return FrameworkUnknown, 0, err
	}
	
	contentStr := string(content)
	
	frameworks := []struct {
		pattern    string
		framework  Framework
		confidence float64
	}{
		{"github.com/gin-gonic/gin", FrameworkGin, 0.98},
		{"github.com/labstack/echo", FrameworkEcho, 0.98},
		{"github.com/gofiber/fiber", FrameworkFiber, 0.98},
		{"github.com/go-chi/chi", FrameworkChi, 0.95},
		{"github.com/gorilla/mux", FrameworkMux, 0.95},
	}
	
	for _, fw := range frameworks {
		if strings.Contains(contentStr, fw.pattern) {
			return fw.framework, fw.confidence, nil
		}
	}
	
	return FrameworkUnknown, 0.6, nil
}

func (d *GoDetector) DetectServices(ctx context.Context, path string) ([]Service, error) {
	var services []Service
	modPath := filepath.Join(path, "go.mod")
	
	content, err := os.ReadFile(modPath)
	if err != nil {
		return services, nil
	}
	
	contentStr := string(content)
	
	if strings.Contains(contentStr, "github.com/lib/pq") || strings.Contains(contentStr, "github.com/jackc/pgx") {
		services = append(services, Service{Type: "postgresql", Version: "15", Reason: "PostgreSQL driver in go.mod"})
	}
	if strings.Contains(contentStr, "github.com/go-redis/redis") || strings.Contains(contentStr, "github.com/redis/go-redis") {
		services = append(services, Service{Type: "redis", Version: "7", Reason: "Redis driver in go.mod"})
	}
	if strings.Contains(contentStr, "go.mongodb.org/mongo-driver") {
		services = append(services, Service{Type: "mongodb", Version: "7", Reason: "MongoDB driver in go.mod"})
	}
	
	return services, nil
}

func (d *GoDetector) ScanSecurity(ctx context.Context, path string) ([]SecurityIssue, error) {
	return []SecurityIssue{}, nil
}

func (d *GoDetector) GetBuildConfig(ctx context.Context, path string, framework Framework) (*BuildConfig, error) {
	return &BuildConfig{
		BuildCommand: "go build -o app .",
		StartCommand: "./app",
		Port:         8080,
		HealthCheck:  "/health",
		BaseImage:    "golang:1.21-alpine",
		MultiStage:   true,
	}, nil
}

// Helper function
func hasService(services []Service, svcType string) bool {
	for _, s := range services {
		if s.Type == svcType {
			return true
		}
	}
	return false
}

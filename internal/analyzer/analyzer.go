package analyzer

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Language represents detected programming language
type Language int

const (
	LanguageUnknown Language = iota
	LanguageNodeJS
	LanguagePython
	LanguageGo
	LanguageRust
	LanguageJava
	LanguageRuby
	LanguagePHP
	LanguageDotNet
)

func (l Language) String() string {
	return [...]string{"Unknown", "Node.js", "Python", "Go", "Rust", "Java", "Ruby", "PHP", ".NET"}[l]
}

// Framework represents detected framework
type Framework int

const (
	FrameworkUnknown Framework = iota
	// Node.js
	FrameworkExpress
	FrameworkNextJS
	FrameworkNestJS
	FrameworkFastify
	FrameworkKoa
	FrameworkHapi
	FrameworkSvelteKit
	FrameworkRemix
	FrameworkAstro
	// Python
	FrameworkFastAPI
	FrameworkDjango
	FrameworkFlask
	FrameworkStarlette
	FrameworkSanic
	FrameworkAiohttp
	// Go
	FrameworkGin
	FrameworkEcho
	FrameworkFiber
	FrameworkChi
	FrameworkMux
	// Rust
	FrameworkActix
	FrameworkRocket
	FrameworkAxum
	FrameworkWarp
	FrameworkTide
	FrameworkPoem
	FrameworkSalvo
	// Java
	FrameworkSpringBoot
	FrameworkQuarkus
	FrameworkMicronaut
	FrameworkPlay
	FrameworkDropwizard
	// Ruby
	FrameworkRails
	FrameworkSinatra
	FrameworkHanami
	FrameworkPadrino
	FrameworkGrape
	// PHP
	FrameworkLaravel
	FrameworkSymfony
	FrameworkCodeIgniter
	FrameworkSlim
	FrameworkLumen
	FrameworkCakePHP
	FrameworkYii
	FrameworkLaminas
	// .NET
	FrameworkASPNETCore
	FrameworkBlazor
	FrameworkNancy
)

func (f Framework) String() string {
	names := map[Framework]string{
		FrameworkUnknown:     "Unknown",
		FrameworkExpress:     "Express",
		FrameworkNextJS:      "Next.js",
		FrameworkNestJS:      "NestJS",
		FrameworkFastify:     "Fastify",
		FrameworkKoa:         "Koa",
		FrameworkHapi:        "Hapi",
		FrameworkSvelteKit:   "SvelteKit",
		FrameworkRemix:       "Remix",
		FrameworkAstro:       "Astro",
		FrameworkFastAPI:     "FastAPI",
		FrameworkDjango:      "Django",
		FrameworkFlask:       "Flask",
		FrameworkStarlette:   "Starlette",
		FrameworkSanic:       "Sanic",
		FrameworkAiohttp:     "aiohttp",
		FrameworkGin:         "Gin",
		FrameworkEcho:        "Echo",
		FrameworkFiber:       "Fiber",
		FrameworkChi:         "Chi",
		FrameworkMux:         "Gorilla Mux",
		FrameworkActix:       "Actix Web",
		FrameworkRocket:      "Rocket",
		FrameworkAxum:        "Axum",
		FrameworkWarp:        "Warp",
		FrameworkTide:        "Tide",
		FrameworkPoem:        "Poem",
		FrameworkSalvo:       "Salvo",
		FrameworkSpringBoot:  "Spring Boot",
		FrameworkQuarkus:     "Quarkus",
		FrameworkMicronaut:   "Micronaut",
		FrameworkPlay:        "Play Framework",
		FrameworkDropwizard:  "Dropwizard",
		FrameworkRails:       "Ruby on Rails",
		FrameworkSinatra:     "Sinatra",
		FrameworkHanami:      "Hanami",
		FrameworkPadrino:     "Padrino",
		FrameworkGrape:       "Grape",
		FrameworkLaravel:     "Laravel",
		FrameworkSymfony:     "Symfony",
		FrameworkCodeIgniter: "CodeIgniter",
		FrameworkSlim:        "Slim",
		FrameworkLumen:       "Lumen",
		FrameworkCakePHP:     "CakePHP",
		FrameworkYii:         "Yii",
		FrameworkLaminas:     "Laminas",
		FrameworkASPNETCore:  "ASP.NET Core",
		FrameworkBlazor:      "Blazor",
		FrameworkNancy:       "Nancy",
	}
	if name, ok := names[f]; ok {
		return name
	}
	return "Unknown"
}

// Service represents a detected external service
type Service struct {
	Type     string `json:"type"`
	Version  string `json:"version,omitempty"`
	Reason   string `json:"reason"`
	Required bool   `json:"required"`
	Config   string `json:"config,omitempty"`
}

// SecurityIssue represents a detected security concern
type SecurityIssue struct {
	Severity    string `json:"severity"` // critical, high, medium, low
	Type        string `json:"type"`
	Description string `json:"description"`
	File        string `json:"file,omitempty"`
	Line        int    `json:"line,omitempty"`
	Suggestion  string `json:"suggestion"`
}

// Dependency represents a package dependency
type Dependency struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	Latest          string `json:"latest,omitempty"`
	Vulnerabilities int    `json:"vulnerabilities"`
	License         string `json:"license,omitempty"`
	Deprecated      bool   `json:"deprecated"`
	DevOnly         bool   `json:"dev_only"`
}

// Resources represents resource estimates
type Resources struct {
	MinCPU      string  `json:"min_cpu"`
	MaxCPU      string  `json:"max_cpu"`
	MinMemory   string  `json:"min_memory"`
	MaxMemory   string  `json:"max_memory"`
	Storage     string  `json:"storage"`
	EstCost     float64 `json:"est_cost"`
	Replicas    int     `json:"replicas"`
	AutoScale   bool    `json:"auto_scale"`
	GPURequired bool    `json:"gpu_required"`
}

// BuildConfig represents build configuration
type BuildConfig struct {
	Dockerfile   string            `json:"dockerfile,omitempty"`
	BuildCommand string            `json:"build_command"`
	StartCommand string            `json:"start_command"`
	Port         int               `json:"port"`
	HealthCheck  string            `json:"health_check"`
	EnvVars      map[string]string `json:"env_vars"`
	BuildArgs    map[string]string `json:"build_args,omitempty"`
	BaseImage    string            `json:"base_image"`
	MultiStage   bool              `json:"multi_stage"`
}

// MonitoringConfig represents auto-configured monitoring
type MonitoringConfig struct {
	MetricsEnabled bool     `json:"metrics_enabled"`
	LoggingEnabled bool     `json:"logging_enabled"`
	TracingEnabled bool     `json:"tracing_enabled"`
	AlertRules     []string `json:"alert_rules"`
	DashboardType  string   `json:"dashboard_type"`
	RetentionDays  int      `json:"retention_days"`
}

// Analysis represents complete project analysis
type Analysis struct {
	ProjectPath  string           `json:"project_path"`
	ProjectName  string           `json:"project_name"`
	Language     Language         `json:"language"`
	Framework    Framework        `json:"framework"`
	EntryPoint   string           `json:"entry_point"`
	Confidence   float64          `json:"confidence"`
	Services     []Service        `json:"services"`
	Dependencies []Dependency     `json:"dependencies"`
	Security     []SecurityIssue  `json:"security_issues"`
	Resources    Resources        `json:"resources"`
	Build        BuildConfig      `json:"build"`
	Monitoring   MonitoringConfig `json:"monitoring"`
	Suggestions  []string         `json:"suggestions"`
}

// Analyzer performs intelligent code analysis
type Analyzer struct {
	detectors []LanguageDetector
}

// LanguageDetector interface for language-specific detection
type LanguageDetector interface {
	Detect(ctx context.Context, path string) (*DetectionResult, error)
	DetectFramework(ctx context.Context, path string) (Framework, float64, error)
	DetectServices(ctx context.Context, path string) ([]Service, error)
	ScanSecurity(ctx context.Context, path string) ([]SecurityIssue, error)
	GetBuildConfig(ctx context.Context, path string, framework Framework) (*BuildConfig, error)
}

// DetectionResult from a language detector
type DetectionResult struct {
	Language   Language
	Confidence float64
	EntryPoint string
	Version    string
}

// New creates a new analyzer with all detectors
func New() *Analyzer {
	return &Analyzer{
		detectors: []LanguageDetector{
			NewNodeDetector(),
			NewPythonDetector(),
			NewGoDetector(),
			NewRustDetector(),
			NewRubyDetector(),
			NewPHPDetector(),
			// TODO: Add Java and .NET detectors
		},
	}
}

// Analyze performs complete analysis of a project
func (a *Analyzer) Analyze(ctx context.Context, projectPath string) (*Analysis, error) {
	analysis := &Analysis{
		ProjectPath:  projectPath,
		ProjectName:  filepath.Base(projectPath),
		Confidence:   0,
		Services:     []Service{},
		Dependencies: []Dependency{},
		Security:     []SecurityIssue{},
		Suggestions:  []string{},
	}

	// Detect language and framework
	var bestDetector LanguageDetector

	for _, detector := range a.detectors {
		result, err := detector.Detect(ctx, projectPath)
		if err != nil {
			continue
		}
		if result != nil && result.Confidence > analysis.Confidence {
			bestDetector = detector
			analysis.Language = result.Language
			analysis.Confidence = result.Confidence
			analysis.EntryPoint = result.EntryPoint
		}
	}

	if bestDetector != nil {
		// Detect framework
		framework, frameworkConf, _ := bestDetector.DetectFramework(ctx, projectPath)
		analysis.Framework = framework
		analysis.Confidence = (analysis.Confidence + frameworkConf) / 2

		// Detect services
		services, _ := bestDetector.DetectServices(ctx, projectPath)
		analysis.Services = services

		// Security scan
		issues, _ := bestDetector.ScanSecurity(ctx, projectPath)
		analysis.Security = issues

		// Get build config
		buildConfig, _ := bestDetector.GetBuildConfig(ctx, projectPath, framework)
		if buildConfig != nil {
			analysis.Build = *buildConfig
		}
	}

	// Parse dependencies
	analysis.Dependencies = a.parseDependencies(projectPath, analysis.Language)

	// Estimate resources
	analysis.Resources = a.estimateResources(analysis)

	// Configure monitoring
	analysis.Monitoring = a.configureMonitoring(analysis)

	// Generate suggestions
	analysis.Suggestions = a.generateSuggestions(analysis)

	return analysis, nil
}

func (a *Analyzer) parseDependencies(path string, lang Language) []Dependency {
	deps := []Dependency{}

	switch lang {
	case LanguageNodeJS:
		pkgPath := filepath.Join(path, "package.json")
		data, err := os.ReadFile(pkgPath)
		if err != nil {
			return deps
		}
		var pkg struct {
			Dependencies    map[string]string `json:"dependencies"`
			DevDependencies map[string]string `json:"devDependencies"`
		}
		if err := json.Unmarshal(data, &pkg); err != nil {
			return deps
		}
		for name, version := range pkg.Dependencies {
			deps = append(deps, Dependency{
				Name:    name,
				Version: version,
				DevOnly: false,
			})
		}
		for name, version := range pkg.DevDependencies {
			deps = append(deps, Dependency{
				Name:    name,
				Version: version,
				DevOnly: true,
			})
		}
	case LanguagePython:
		reqPath := filepath.Join(path, "requirements.txt")
		file, err := os.Open(reqPath)
		if err != nil {
			return deps
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := regexp.MustCompile(`[=<>!]+`).Split(line, 2)
			name := parts[0]
			version := ""
			if len(parts) > 1 {
				version = parts[1]
			}
			deps = append(deps, Dependency{
				Name:    name,
				Version: version,
			})
		}
	case LanguageGo:
		modPath := filepath.Join(path, "go.mod")
		data, err := os.ReadFile(modPath)
		if err != nil {
			return deps
		}
		lines := strings.Split(string(data), "\n")
		inRequire := false
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "require (") {
				inRequire = true
				continue
			}
			if line == ")" {
				inRequire = false
				continue
			}
			if inRequire || strings.HasPrefix(line, "require ") {
				parts := strings.Fields(strings.TrimPrefix(line, "require "))
				if len(parts) >= 2 {
					deps = append(deps, Dependency{
						Name:    parts[0],
						Version: parts[1],
					})
				}
			}
		}
	}

	return deps
}

func (a *Analyzer) estimateResources(analysis *Analysis) Resources {
	resources := Resources{
		MinCPU:    "250m",
		MaxCPU:    "1000m",
		MinMemory: "256Mi",
		MaxMemory: "1Gi",
		Storage:   "10Gi",
		EstCost:   25.0,
		Replicas:  1,
		AutoScale: true,
	}

	// Adjust based on framework
	switch analysis.Framework {
	case FrameworkNextJS, FrameworkNuxt, FrameworkRemix:
		resources.MinMemory = "512Mi"
		resources.MaxMemory = "2Gi"
		resources.EstCost = 45.0
	case FrameworkDjango, FrameworkRails:
		resources.MinCPU = "500m"
		resources.MaxCPU = "2000m"
		resources.MinMemory = "512Mi"
		resources.MaxMemory = "2Gi"
		resources.EstCost = 65.0
	case FrameworkSpringBoot:
		resources.MinMemory = "1Gi"
		resources.MaxMemory = "4Gi"
		resources.EstCost = 95.0
	}

	// Adjust based on services
	for _, svc := range analysis.Services {
		switch svc.Type {
		case "postgresql", "mysql":
			resources.EstCost += 25.0
			resources.Storage = "50Gi"
		case "mongodb":
			resources.EstCost += 35.0
			resources.Storage = "100Gi"
		case "redis":
			resources.EstCost += 15.0
		case "elasticsearch":
			resources.EstCost += 50.0
			resources.MaxMemory = "4Gi"
		}
	}

	// Adjust based on dependency count
	depCount := len(analysis.Dependencies)
	if depCount > 50 {
		resources.MinMemory = "512Mi"
		resources.MaxMemory = "2Gi"
	}
	if depCount > 100 {
		resources.MinMemory = "1Gi"
		resources.MaxMemory = "4Gi"
		resources.EstCost += 20.0
	}

	return resources
}

func (a *Analyzer) configureMonitoring(analysis *Analysis) MonitoringConfig {
	config := MonitoringConfig{
		MetricsEnabled: true,
		LoggingEnabled: true,
		TracingEnabled: false,
		AlertRules: []string{
			"cpu_usage > 80%",
			"memory_usage > 85%",
			"error_rate > 1%",
			"latency_p99 > 500ms",
		},
		DashboardType: "standard",
		RetentionDays: 30,
	}

	// Enable tracing for microservice architectures
	if analysis.Framework == FrameworkNestJS || analysis.Framework == FrameworkFastAPI ||
		analysis.Framework == FrameworkGin || analysis.Framework == FrameworkSpringBoot {
		config.TracingEnabled = true
	}

	// Add framework-specific alerts
	switch analysis.Framework {
	case FrameworkNextJS, FrameworkRemix:
		config.AlertRules = append(config.AlertRules, "ssr_render_time > 200ms")
	case FrameworkExpress, FrameworkFastify:
		config.AlertRules = append(config.AlertRules, "request_queue_size > 100")
	}

	return config
}

func (a *Analyzer) generateSuggestions(analysis *Analysis) []string {
	suggestions := []string{}

	// Security suggestions
	if len(analysis.Security) > 0 {
		criticalCount := 0
		for _, issue := range analysis.Security {
			if issue.Severity == "critical" || issue.Severity == "high" {
				criticalCount++
			}
		}
		if criticalCount > 0 {
			suggestions = append(suggestions,
				fmt.Sprintf("🔒 %d security issues found - review before deploying", criticalCount))
		}
	}

	// Dependency suggestions
	outdatedCount := 0
	for _, dep := range analysis.Dependencies {
		if dep.Vulnerabilities > 0 {
			outdatedCount++
		}
	}
	if outdatedCount > 0 {
		suggestions = append(suggestions,
			fmt.Sprintf("📦 %d dependencies have known vulnerabilities", outdatedCount))
	}

	// Performance suggestions
	if analysis.Resources.EstCost > 100 {
		suggestions = append(suggestions,
			"💡 Consider using spot instances to reduce costs by up to 70%")
	}

	// Framework-specific suggestions
	switch analysis.Framework {
	case FrameworkNextJS:
		suggestions = append(suggestions,
			"⚡ Enable ISR for better performance on dynamic pages")
	case FrameworkExpress:
		suggestions = append(suggestions,
			"⚡ Add compression middleware for smaller response sizes")
	case FrameworkDjango:
		suggestions = append(suggestions,
			"⚡ Enable database connection pooling for better performance")
	}

	// Missing configurations
	if analysis.Build.HealthCheck == "" {
		suggestions = append(suggestions,
			"❤️ Add a health check endpoint for better reliability monitoring")
	}

	return suggestions
}

// Nuxt framework constant (was missing)
const FrameworkNuxt Framework = 100

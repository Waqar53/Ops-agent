package cicd

import (
	"context"
	"fmt"
	"time"
)

// Pipeline represents a CI/CD pipeline
type Pipeline struct {
	ID          string
	ProjectID   string
	Name        string
	Stages      []Stage
	Triggers    []Trigger
	Environment map[string]string
	Secrets     map[string]string
	Artifacts   []Artifact
	CreatedAt   time.Time
}

// Stage represents a pipeline stage
type Stage struct {
	Name      string
	Jobs      []Job
	Parallel  bool
	Condition string
}

// Job represents a job within a stage
type Job struct {
	Name        string
	Image       string
	Script      []string
	Environment map[string]string
	Artifacts   []string
	Cache       *CacheConfig
	Timeout     time.Duration
	Retry       int
}

// CacheConfig for job caching
type CacheConfig struct {
	Paths []string
	Key   string
}

// Trigger defines when a pipeline runs
type Trigger struct {
	Type   string // push, pull_request, schedule, manual
	Branch string
	Cron   string
}

// Artifact represents a build artifact
type Artifact struct {
	Name string
	Path string
	Type string // docker, binary, archive
}

// PipelineExecutor executes CI/CD pipelines
type PipelineExecutor struct {
	containerRunner ContainerRunner
	testRunner      TestRunner
	securityScanner SecurityScanner
	artifactStore   ArtifactStore
}

// ContainerRunner runs containers
type ContainerRunner interface {
	Run(ctx context.Context, image string, cmd []string, env map[string]string) error
}

// TestRunner runs tests
type TestRunner interface {
	RunUnitTests(ctx context.Context, path string) (*TestResult, error)
	RunIntegrationTests(ctx context.Context, path string) (*TestResult, error)
	RunE2ETests(ctx context.Context, path string) (*TestResult, error)
}

// SecurityScanner scans for vulnerabilities
type SecurityScanner interface {
	ScanCode(ctx context.Context, path string) (*ScanResult, error)
	ScanDependencies(ctx context.Context, path string) (*ScanResult, error)
	ScanContainer(ctx context.Context, image string) (*ScanResult, error)
}

// ArtifactStore stores build artifacts
type ArtifactStore interface {
	Upload(ctx context.Context, artifact *Artifact) error
	Download(ctx context.Context, name string) (*Artifact, error)
}

// TestResult holds test results
type TestResult struct {
	Passed   int
	Failed   int
	Skipped  int
	Duration time.Duration
	Coverage float64
}

// ScanResult holds security scan results
type ScanResult struct {
	Critical int
	High     int
	Medium   int
	Low      int
	Issues   []SecurityIssue
}

// SecurityIssue represents a security issue
type SecurityIssue struct {
	Severity    string
	Type        string
	Description string
	File        string
	Line        int
	Fix         string
}

// NewPipelineExecutor creates a new pipeline executor
func NewPipelineExecutor(
	runner ContainerRunner,
	tester TestRunner,
	scanner SecurityScanner,
	store ArtifactStore,
) *PipelineExecutor {
	return &PipelineExecutor{
		containerRunner: runner,
		testRunner:      tester,
		securityScanner: scanner,
		artifactStore:   store,
	}
}

// Execute executes a pipeline
func (pe *PipelineExecutor) Execute(ctx context.Context, pipeline *Pipeline) (*PipelineResult, error) {
	result := &PipelineResult{
		PipelineID: pipeline.ID,
		StartTime:  time.Now(),
		Stages:     []StageResult{},
	}

	for _, stage := range pipeline.Stages {
		stageResult := pe.executeStage(ctx, &stage, pipeline)
		result.Stages = append(result.Stages, stageResult)

		if stageResult.Status == "failed" {
			result.Status = "failed"
			result.EndTime = time.Now()
			return result, fmt.Errorf("stage %s failed", stage.Name)
		}
	}

	result.Status = "success"
	result.EndTime = time.Now()
	return result, nil
}

func (pe *PipelineExecutor) executeStage(ctx context.Context, stage *Stage, pipeline *Pipeline) StageResult {
	result := StageResult{
		Name:      stage.Name,
		StartTime: time.Now(),
		Jobs:      []JobResult{},
	}

	if stage.Parallel {
		// Execute jobs in parallel
		jobResults := make(chan JobResult, len(stage.Jobs))
		for _, job := range stage.Jobs {
			go func(j Job) {
				jobResults <- pe.executeJob(ctx, &j, pipeline)
			}(job)
		}

		for range stage.Jobs {
			jobResult := <-jobResults
			result.Jobs = append(result.Jobs, jobResult)
			if jobResult.Status == "failed" {
				result.Status = "failed"
			}
		}
	} else {
		// Execute jobs sequentially
		for _, job := range stage.Jobs {
			jobResult := pe.executeJob(ctx, &job, pipeline)
			result.Jobs = append(result.Jobs, jobResult)
			if jobResult.Status == "failed" {
				result.Status = "failed"
				break
			}
		}
	}

	if result.Status == "" {
		result.Status = "success"
	}

	result.EndTime = time.Now()
	return result
}

func (pe *PipelineExecutor) executeJob(ctx context.Context, job *Job, pipeline *Pipeline) JobResult {
	result := JobResult{
		Name:      job.Name,
		StartTime: time.Now(),
	}

	// Merge environment variables
	env := make(map[string]string)
	for k, v := range pipeline.Environment {
		env[k] = v
	}
	for k, v := range job.Environment {
		env[k] = v
	}

	// Execute job script
	for _, cmd := range job.Script {
		if err := pe.containerRunner.Run(ctx, job.Image, []string{"sh", "-c", cmd}, env); err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			result.EndTime = time.Now()
			return result
		}
	}

	result.Status = "success"
	result.EndTime = time.Now()
	return result
}

// GeneratePipeline generates a pipeline based on project analysis
func (pe *PipelineExecutor) GeneratePipeline(language, framework string) *Pipeline {
	pipeline := &Pipeline{
		ID:   fmt.Sprintf("pipeline_%d", time.Now().UnixNano()),
		Name: "Auto-generated Pipeline",
		Stages: []Stage{
			{
				Name: "Build",
				Jobs: []Job{
					{
						Name:  "build",
						Image: pe.getBuildImage(language),
						Script: []string{
							pe.getBuildCommand(language, framework),
						},
						Timeout: 10 * time.Minute,
					},
				},
			},
			{
				Name: "Test",
				Jobs: []Job{
					{
						Name:  "unit-tests",
						Image: pe.getBuildImage(language),
						Script: []string{
							pe.getTestCommand(language, framework),
						},
						Timeout: 15 * time.Minute,
					},
				},
			},
			{
				Name: "Security",
				Jobs: []Job{
					{
						Name:  "security-scan",
						Image: "aquasec/trivy:latest",
						Script: []string{
							"trivy fs --severity HIGH,CRITICAL .",
						},
						Timeout: 5 * time.Minute,
					},
				},
			},
			{
				Name: "Deploy",
				Jobs: []Job{
					{
						Name:  "deploy",
						Image: "ops-agent/deployer:latest",
						Script: []string{
							"ops deploy --environment production",
						},
						Timeout: 20 * time.Minute,
					},
				},
			},
		},
		Triggers: []Trigger{
			{Type: "push", Branch: "main"},
			{Type: "pull_request"},
		},
	}

	return pipeline
}

func (pe *PipelineExecutor) getBuildImage(language string) string {
	images := map[string]string{
		"nodejs": "node:18-alpine",
		"python": "python:3.11-alpine",
		"go":     "golang:1.21-alpine",
		"rust":   "rust:1.75-alpine",
		"ruby":   "ruby:3.2-alpine",
		"php":    "php:8.2-alpine",
	}
	if img, ok := images[language]; ok {
		return img
	}
	return "alpine:latest"
}

func (pe *PipelineExecutor) getBuildCommand(language, framework string) string {
	commands := map[string]string{
		"nodejs": "npm ci && npm run build",
		"python": "pip install -r requirements.txt",
		"go":     "go build -o app .",
		"rust":   "cargo build --release",
		"ruby":   "bundle install",
		"php":    "composer install --no-dev",
	}
	if cmd, ok := commands[language]; ok {
		return cmd
	}
	return "echo 'No build command defined'"
}

func (pe *PipelineExecutor) getTestCommand(language, framework string) string {
	commands := map[string]string{
		"nodejs": "npm test",
		"python": "pytest",
		"go":     "go test ./...",
		"rust":   "cargo test",
		"ruby":   "bundle exec rspec",
		"php":    "vendor/bin/phpunit",
	}
	if cmd, ok := commands[language]; ok {
		return cmd
	}
	return "echo 'No test command defined'"
}

// PipelineResult holds pipeline execution result
type PipelineResult struct {
	PipelineID string
	Status     string
	StartTime  time.Time
	EndTime    time.Time
	Stages     []StageResult
}

// StageResult holds stage execution result
type StageResult struct {
	Name      string
	Status    string
	StartTime time.Time
	EndTime   time.Time
	Jobs      []JobResult
}

// JobResult holds job execution result
type JobResult struct {
	Name      string
	Status    string
	StartTime time.Time
	EndTime   time.Time
	Error     string
}

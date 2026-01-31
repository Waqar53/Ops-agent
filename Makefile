.PHONY: all build test run clean docker dev

# Build variables
VERSION ?= 0.1.0
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"

# Go settings
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOCLEAN := $(GOCMD) clean
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Binary names
BIN_DIR := bin
API_BINARY := $(BIN_DIR)/opsagent
CLI_BINARY := $(BIN_DIR)/ops

# Default target
all: build

## Build commands

build: build-api build-cli ## Build all binaries

build-api: ## Build API server
	@echo "Building API server..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(API_BINARY) ./cmd/opsagent

build-cli: ## Build CLI tool
	@echo "Building CLI..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(CLI_BINARY) ./cmd/opsctl

## Development commands

dev: ## Start development environment
	@echo "Starting development environment..."
	docker-compose up -d postgres redis
	@echo "Waiting for services..."
	@sleep 3
	@echo "Database: postgresql://opsagent:opsagent_dev_password@localhost:5432/opsagent"
	@echo "Redis: localhost:6379"

dev-full: ## Start full development environment
	@echo "Starting full development environment..."
	docker-compose --profile full up -d
	@echo "Services started!"

dev-down: ## Stop development environment
	docker-compose --profile full down

run-api: build-api ## Run API server
	@echo "Running API server..."
	DB_HOST=localhost DB_PASSWORD=opsagent_dev_password $(API_BINARY)

run-web: ## Run web dashboard
	@echo "Running web dashboard..."
	cd web && npm run dev

## Test commands

test: ## Run all tests
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

lint: ## Run linters
	@echo "Running golangci-lint..."
	golangci-lint run ./...

## Database commands

db-migrate: ## Run database migrations
	@echo "Running migrations..."
	$(GOBUILD) -o /tmp/migrate ./cmd/migrate
	DB_HOST=localhost DB_PASSWORD=opsagent_dev_password /tmp/migrate up

db-reset: ## Reset database
	@echo "Resetting database..."
	docker-compose exec postgres psql -U opsagent -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	$(MAKE) db-migrate

## Docker commands

docker-build: ## Build Docker images
	@echo "Building Docker images..."
	docker build -t opsagent/api:$(VERSION) -f docker/Dockerfile.api .
	docker build -t opsagent/cli:$(VERSION) -f docker/Dockerfile.cli .

docker-push: ## Push Docker images
	docker push opsagent/api:$(VERSION)
	docker push opsagent/cli:$(VERSION)

## Clean commands

clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BIN_DIR)
	rm -f coverage.out coverage.html

## Install commands

install: build-cli ## Install CLI locally
	@echo "Installing ops CLI..."
	cp $(CLI_BINARY) /usr/local/bin/ops
	@echo "Installed to /usr/local/bin/ops"

## Dependency commands

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

deps-update: ## Update dependencies
	$(GOGET) -u ./...
	$(GOMOD) tidy

## Setup commands

setup: ## Initial project setup
	@echo "Setting up project..."
	$(MAKE) deps
	@if [ -d "web" ]; then cd web && npm install; fi
	$(MAKE) dev
	@echo "Setup complete!"

## Generate commands

generate: ## Run code generation
	$(GOCMD) generate ./...

proto: ## Generate protobuf files
	@echo "Generating protobuf..."
	protoc --go_out=. --go-grpc_out=. proto/*.proto

## Help

help: ## Show this help
	@echo "OpsAgent - DevOps on Autopilot"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

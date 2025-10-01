# Policy Engine Makefile
# This Makefile provides common commands for building, testing, and running the policy engine

# Variables
APP_NAME := policy-engine
VERSION := 1.0.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD)
GO_VERSION := $(shell go version | cut -d' ' -f3)

# Build variables
BUILD_DIR := build
BINARY_NAME := $(APP_NAME)
MAIN_PACKAGE := ./cmd/main.go
DOCKER_IMAGE := $(APP_NAME):$(VERSION)
DOCKER_LATEST := $(APP_NAME):latest

# Go build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT) -X main.goVersion=$(GO_VERSION)"
BUILD_FLAGS := -a -installsuffix cgo -o

# Test variables
TEST_TIMEOUT := 30m
TEST_COVERAGE := coverage.out
TEST_PROFILE := profile.out

# Docker variables
DOCKER_REGISTRY := 
DOCKER_NAMESPACE := kcloud-opt

# Default target
.DEFAULT_GOAL := help

# Help target
.PHONY: help
help: ## Show this help message
	@echo "Policy Engine Makefile"
	@echo "====================="
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development targets
.PHONY: dev
dev: ## Run the application in development mode
	@echo "Starting policy engine in development mode..."
	@go run $(MAIN_PACKAGE) --config config/config.yaml --debug

.PHONY: dev-docker
dev-docker: ## Run the application in Docker development mode
	@echo "Starting policy engine in Docker development mode..."
	@docker-compose -f docker-compose.dev.yml up --build

# Build targets
.PHONY: build
build: clean ## Build the application binary
	@echo "Building $(APP_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) $(BUILD_FLAGS) $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: build-all
build-all: clean ## Build binaries for all platforms
	@echo "Building $(APP_NAME) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@echo "Building for Linux AMD64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) $(BUILD_FLAGS) $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	@echo "Building for Linux ARM64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) $(BUILD_FLAGS) $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)
	@echo "Building for macOS AMD64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) $(BUILD_FLAGS) $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	@echo "Building for macOS ARM64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) $(BUILD_FLAGS) $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	@echo "Building for Windows AMD64..."
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) $(BUILD_FLAGS) $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
	@echo "Build completed for all platforms"

.PHONY: install
install: build ## Install the binary to GOPATH/bin
	@echo "Installing $(APP_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

# Test targets
.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	@go test -v -timeout $(TEST_TIMEOUT) ./...

.PHONY: test-race
test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	@go test -v -race -timeout $(TEST_TIMEOUT) ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -timeout $(TEST_TIMEOUT) -coverprofile=$(TEST_COVERAGE) ./...
	@go tool cover -html=$(TEST_COVERAGE) -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-benchmark
test-benchmark: ## Run benchmark tests
	@echo "Running benchmark tests..."
	@go test -v -bench=. -benchmem ./...

.PHONY: test-profile
test-profile: ## Run tests with profiling
	@echo "Running tests with profiling..."
	@go test -v -timeout $(TEST_TIMEOUT) -cpuprofile=$(TEST_PROFILE) ./...

# Code quality targets
.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting Go code..."
	@go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@golangci-lint run

.PHONY: lint-fix
lint-fix: ## Run golangci-lint with auto-fix
	@echo "Running golangci-lint with auto-fix..."
	@golangci-lint run --fix

.PHONY: check
check: fmt vet lint ## Run all code quality checks

# Dependency targets
.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

.PHONY: deps-verify
deps-verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	@go mod verify

.PHONY: tidy
tidy: ## Tidy up go.mod and go.sum
	@echo "Tidying up go.mod and go.sum..."
	@go mod tidy

# Docker targets
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image $(DOCKER_IMAGE)..."
	@docker build -t $(DOCKER_IMAGE) -t $(DOCKER_LATEST) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

.PHONY: docker-run
docker-run: docker-build ## Run Docker container
	@echo "Running Docker container..."
	@docker run --rm -p 8005:8005 $(DOCKER_IMAGE)

.PHONY: docker-push
docker-push: docker-build ## Push Docker image to registry
	@echo "Pushing Docker image to registry..."
	@docker push $(DOCKER_IMAGE)
	@docker push $(DOCKER_LATEST)

# Kubernetes targets
.PHONY: k8s-deploy
k8s-deploy: ## Deploy to Kubernetes
	@echo "Deploying to Kubernetes..."
	@kubectl apply -f k8s/

.PHONY: k8s-delete
k8s-delete: ## Delete from Kubernetes
	@echo "Deleting from Kubernetes..."
	@kubectl delete -f k8s/

.PHONY: k8s-logs
k8s-logs: ## Show Kubernetes logs
	@echo "Showing Kubernetes logs..."
	@kubectl logs -f deployment/$(APP_NAME)

# Utility targets
.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(TEST_COVERAGE)
	@rm -f $(TEST_PROFILE)
	@rm -f coverage.html
	@echo "Clean completed"

.PHONY: clean-deps
clean-deps: ## Clean dependency cache
	@echo "Cleaning dependency cache..."
	@go clean -modcache

.PHONY: clean-all
clean-all: clean clean-deps ## Clean everything
	@echo "Cleaning everything..."
	@docker system prune -f

# Release targets
.PHONY: release
release: test check build-all ## Create a release build
	@echo "Creating release $(VERSION)..."
	@mkdir -p $(BUILD_DIR)/release
	@cp $(BUILD_DIR)/$(BINARY_NAME)-* $(BUILD_DIR)/release/
	@cp README.md $(BUILD_DIR)/release/
	@cp config/config.yaml $(BUILD_DIR)/release/
	@cp -r examples/ $(BUILD_DIR)/release/
	@cd $(BUILD_DIR) && tar -czf $(APP_NAME)-$(VERSION).tar.gz release/
	@echo "Release created: $(BUILD_DIR)/$(APP_NAME)-$(VERSION).tar.gz"

# Development workflow targets
.PHONY: dev-setup
dev-setup: deps ## Setup development environment
	@echo "Setting up development environment..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development environment setup completed"

.PHONY: pre-commit
pre-commit: check test ## Run pre-commit checks
	@echo "Running pre-commit checks..."
	@echo "Pre-commit checks completed"

.PHONY: ci
ci: deps-verify check test-coverage build ## Run CI pipeline
	@echo "Running CI pipeline..."
	@echo "CI pipeline completed"

# Documentation targets
.PHONY: docs
docs: ## Generate documentation
	@echo "Generating documentation..."
	@go doc -all ./... > docs/api.md
	@echo "Documentation generated: docs/api.md"

# Monitoring targets
.PHONY: metrics
metrics: ## Show application metrics
	@echo "Application metrics available at: http://localhost:8005/metrics"

.PHONY: health
health: ## Check application health
	@echo "Checking application health..."
	@curl -f http://localhost:8005/health || echo "Health check failed"

# Policy management targets
.PHONY: load-policies
load-policies: ## Load example policies
	@echo "Loading example policies..."
	@curl -X POST -H "Content-Type: application/json" -d @examples/policies/cost-optimization-policy.yaml http://localhost:8005/api/v1/policies || echo "Failed to load cost optimization policy"
	@curl -X POST -H "Content-Type: application/json" -d @examples/policies/automation-rule.yaml http://localhost:8005/api/v1/automation/rules || echo "Failed to load automation rule"
	@echo "Example policies loaded"

.PHONY: list-policies
list-policies: ## List all policies
	@echo "Listing all policies..."
	@curl -s http://localhost:8005/api/v1/policies | jq '.' || echo "Failed to list policies"

# Version information
.PHONY: version
version: ## Show version information
	@echo "Application: $(APP_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Go Version: $(GO_VERSION)"

# Caddy Admin UI Makefile
# Optimized for production builds and development

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=caddy-admin-ui
BINARY_UNIX=$(BINARY_NAME)_unix

# Build flags
LDFLAGS=-s -w -X main.compileTime=$(shell date +%s) -X main.gitCommit=$(shell git rev-parse --short HEAD)
BUILD_FLAGS=-ldflags="$(LDFLAGS)" -buildmode=pie -trimpath

# Production build flags (optimized)
PROD_LDFLAGS=-s -w -X main.compileTime=$(shell date +%s) -X main.gitCommit=$(shell git rev-parse --short HEAD)
PROD_FLAGS=-ldflags="$(PROD_LDFLAGS)" -buildmode=pie -trimpath -tags=release

# Coverage
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Frontend directory
WEB_DIR=web
BUILD_DIR=build

# Version
VERSION ?= $(shell git describe --tags --always --dirty)
PLATFORMS ?= linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: all build clean test coverage benchmark deps frontend docker release help

# Default target
all: deps frontend build

# Install dependencies
deps:
	@echo "Installing Go dependencies..."
	$(GOMOD) download
	$(GOMOD) verify
	@echo "Installing frontend dependencies..."
	cd $(WEB_DIR) && npm install

# Build frontend
frontend:
	@echo "Building frontend..."
	cd $(WEB_DIR) && npm run build
	@echo "Frontend built successfully"

# Build the binary
build: frontend
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_NAME) -v .
	@echo "Build complete: $(BINARY_NAME)"

# Production build (optimized)
prod: deps frontend
	@echo "Building production binary..."
	$(GOBUILD) $(PROD_FLAGS) -o $(BINARY_NAME)-$(VERSION) -v .
	@echo "Production build complete: $(BINARY_NAME)-$(VERSION)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -rf $(BUILD_DIR)
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
coverage: test
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=$(COVERAGE_FILE) ./...
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem -run=^$$ ./...

# Run specific benchmarks
benchmark-websocket:
	@echo "Running WebSocket benchmarks..."
	$(GOTEST) -bench=BenchmarkWebSocket -benchmem -run=^$$ ./...

# Run race condition tests
race:
	@echo "Running tests with race detection..."
	$(GOTEST) -race -v ./...

# Run memory sanitizer
msan:
	@echo "Running tests with memory sanitizer..."
	$(GOTEST) -msan -v ./...

# Lint code
lint:
	@echo "Running linter..."
	golangci-lint run
	@echo "Linting complete"

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	cd $(WEB_DIR) && npm run lint
	@echo "Code formatted"

# Security scan
security:
	@echo "Running security scan..."
	gosec ./...
	cd $(WEB_DIR) && npm audit
	@echo "Security scan complete"

# Build for multiple platforms
release:
	@echo "Building release for multiple platforms..."
	@$(foreach platform,$(PLATFORMS), \
		echo "Building for $(platform)..."; \
		GOOS=$(word 1,$(subst /, ,$(platform))) GOARCH=$(word 2,$(subst /, ,$(platform))) \
		$(GOBUILD) $(PROD_FLAGS) -o $(BINARY_NAME)-$(VERSION)-$(platform) .; \
	)
	@echo "Release builds complete"

# Development server (with frontend)
dev:
	@echo "Starting development server..."
	@echo "Starting frontend..."
	cd $(WEB_DIR) && npm run dev &
	@echo "Starting backend..."
	$(GOCMD) run .

# Development with hot reload
dev-watch:
	@echo "Starting development with hot reload..."
	air -c .air.toml

# Build Docker image
docker:
	@echo "Building Docker image..."
	docker build -t caddy-admin-ui:$(VERSION) .
	docker tag caddy-admin-ui:$(VERSION) caddy-admin-ui:latest
	@echo "Docker image built"

# Docker multi-platform build
docker-buildx:
	@echo "Building Docker image for multiple platforms..."
	docker buildx build --platform linux/amd64,linux/arm64 -t caddy-admin-ui:$(VERSION) --push .
	@echo "Docker multi-platform build complete"

# Install tools
install-tools:
	@echo "Installing development tools..."
	$(GOGET) -u github.com/cosmtrek/air
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GOGET) -u github.com/securecodewarrior/gosec/v2/cmd/gosec
	@echo "Tools installed"

# Generate docs
docs:
	@echo "Generating documentation..."
	godoc -http=:6060

# Profile CPU
profile-cpu:
	@echo "Starting CPU profiling..."
	$(GOTEST) -cpuprofile=cpu.prof -bench=. .
	$(GOCMD) tool pprof cpu.prof

# Profile Memory
profile-mem:
	@echo "Starting memory profiling..."
	$(GOTEST) -memprofile=mem.prof -bench=. .
	$(GOCMD) tool pprof mem.prof

# Update dependencies
update-deps:
	@echo "Updating dependencies..."
	$(GOMOD) get -u ./...
	$(GOMOD) tidy
	cd $(WEB_DIR) && npm update
	@echo "Dependencies updated"

# Verify dependencies
verify-deps:
	@echo "Verifying dependencies..."
	$(GOMOD) verify
	@echo "Dependencies verified"

# Size analysis
size:
	@echo "Analyzing binary size..."
	$(GOBUILD) $(PROD_FLAGS) -o $(BINARY_NAME)-size .
	ls -lh $(BINARY_NAME)-size
	file $(BINARY_NAME)-size

# Performance test
perf:
	@echo "Running performance tests..."
	$(GOTEST) -bench=BenchmarkConcurrentConnections -benchtime=10s ./...

# Load test (requires hey)
load-test:
	@echo "Running load test..."
	hey -n 1000 -c 10 http://localhost:8080/

# Integration tests
integration:
	@echo "Running integration tests..."
	$(GOTEST) -tags=integration -v ./...

# End-to-end tests
e2e:
	@echo "Running E2E tests..."
	cd $(WEB_DIR) && npm run test:e2e

# Pre-commit checks
pre-commit: fmt lint security test
	@echo "Pre-commit checks complete"

# CI/CD pipeline
ci: deps fmt lint security test coverage
	@echo "CI pipeline complete"

# Quick build for development
quick: build

# Full production build
full: deps frontend lint security test coverage prod
	@echo "Full production build complete"

# Help target
help:
	@echo "Available targets:"
	@echo "  all          - Install deps, build frontend, and build binary"
	@echo "  build        - Build the binary with default flags"
	@echo "  prod         - Build optimized production binary"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  coverage     - Run tests with coverage report"
	@echo "  benchmark    - Run all benchmarks"
	@echo "  deps         - Install Go and npm dependencies"
	@echo "  frontend     - Build frontend only"
	@echo "  dev          - Start development servers"
	@echo "  dev-watch    - Start with hot reload (requires air)"
	@echo "  docker       - Build Docker image"
	@echo "  release      - Build for multiple platforms"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"
	@echo "  security     - Run security scan"
	@echo "  race         - Run tests with race detection"
	@echo "  help         - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make quick          - Quick development build"
	@echo "  make prod           - Production build"
	@echo "  make benchmark      - Run performance benchmarks"
	@echo "  make docker         - Build Docker image"
	@echo "  make pre-commit     - Run all pre-commit checks"
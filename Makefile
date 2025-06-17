# SmartProxy Makefile

# Variables
BINARY_NAME := smartproxy
MAIN_FILES := main.go config.go
BUILD_DIR := build
DOCKER_IMAGE := smartproxy
DOCKER_TAG := latest

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Build flags
LDFLAGS := -s -w
CGO_ENABLED := 0

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILES)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
.PHONY: build-all
build-all: build-linux build-darwin build-windows

.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_FILES)
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_FILES)

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_FILES)
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_FILES)

.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_FILES)

# Run the proxy
.PHONY: run
run: build
	@echo "Starting $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@echo "Clean complete"

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@$(GOMOD) tidy

# Update dependencies
.PHONY: update-deps
update-deps:
	@echo "Updating dependencies..."
	@$(GOGET) -u ./...
	@$(GOMOD) tidy

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@./test_proxy.sh
	@./test_https.sh
	@./test_http_js.sh
	@./test_js_direct.sh

# Generate CA certificate
.PHONY: ca-cert
ca-cert:
	@echo "Generating CA certificate..."
	@./generate_ca.sh

# Docker targets
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	@docker run -it --rm -p 8888:8888 -v $(PWD)/config.yaml:/app/config.yaml:ro $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-push
docker-push:
	@echo "Pushing Docker image..."
	@docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

# Development helpers
.PHONY: dev
dev:
	@echo "Running in development mode with live reload..."
	@pkill -f $(BINARY_NAME) 2>/dev/null || true
	@$(GOBUILD) -o $(BINARY_NAME) $(MAIN_FILES) && ./$(BINARY_NAME)

.PHONY: kill
kill:
	@echo "Killing running proxy..."
	@pkill -f $(BINARY_NAME) || true

# Linting and formatting
.PHONY: lint
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@gofmt -s -w .

# Help
.PHONY: help
help:
	@echo "SmartProxy Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all          - Build the binary (default)"
	@echo "  build        - Build the binary for current platform"
	@echo "  build-all    - Build for all platforms (Linux, macOS, Windows)"
	@echo "  run          - Build and run the proxy"
	@echo "  clean        - Remove build artifacts"
	@echo "  deps         - Download dependencies"
	@echo "  update-deps  - Update dependencies"
	@echo "  test         - Run all tests"
	@echo "  ca-cert      - Generate CA certificate for HTTPS MITM"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  dev          - Run in development mode"
	@echo "  kill         - Kill running proxy"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"
	@echo "  help         - Show this help message"
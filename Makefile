# SmartProxy Makefile

# Variables
BINARY_NAME := smartproxy
MAIN_PATH := ./cmd/smartproxy
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
	@CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
.PHONY: build-all
build-all: build-linux build-darwin build-windows

.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

# Run the proxy
.PHONY: run
run: build
	@echo "Starting $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

# Run the proxy in debug mode
.PHONY: debug
debug: build
	@echo "Starting $(BINARY_NAME) in debug mode..."
	@cp configs/config.example.yaml configs/config.yaml 2>/dev/null || true
	@sed -i '' 's/level: info/level: debug/' configs/config.yaml 2>/dev/null || sed -i 's/level: info/level: debug/' configs/config.yaml 2>/dev/null || true
	@SMARTPROXY_CONFIG=configs/config.yaml ./$(BUILD_DIR)/$(BINARY_NAME)

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
	@./scripts/test/test_proxy.sh
	@./scripts/test/test_https.sh
	@./scripts/test/test_http_js.sh
	@./scripts/test/test_js_direct.sh

# Generate CA certificate
.PHONY: ca-cert
ca-cert:
	@echo "Generating CA certificate..."
	@./scripts/setup/generate_ca.sh

# Docker targets
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	@docker build -f docker/Dockerfile -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	@docker run -it --rm -p 8888:8888  \
		-v $(PWD)/configs:/app/configs:ro \
		-v $(PWD)/certs:/app/certs:ro \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-push
docker-push:
	@echo "Pushing Docker image..."
	@docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

# RouterOS 7 Docker targets
.PHONY: docker-build-routeros
docker-build-routeros:
	@echo "Building RouterOS 7 Docker image (ARM64)..."
	@docker build --platform=linux/arm64 -f docker/Dockerfile.routeros -t $(DOCKER_IMAGE):routeros .

.PHONY: docker-run-routeros
docker-run-routeros:
	@echo "Running RouterOS 7 Docker container..."
	@docker run -it --rm --platform=linux/arm64 -p 8888:8888 \
		-v $(PWD)/configs:/app/configs:ro \
		-v $(PWD)/certs:/app/certs:ro \
		$(DOCKER_IMAGE):routeros

.PHONY: docker-compose-routeros
docker-compose-routeros:
	@echo "Starting RouterOS 7 services with docker-compose..."
	@cd docker && docker-compose -f docker-compose.routeros.yml up -d

.PHONY: docker-compose-routeros-down
docker-compose-routeros-down:
	@echo "Stopping RouterOS 7 services..."
	@cd docker && docker-compose -f docker-compose.routeros.yml down

.PHONY: docker-setup-routeros
docker-setup-routeros:
	@echo "Setting up RouterOS 7 deployment directories..."
	@mkdir -p /routeros/{configs,certs,logs}
	@cp configs/config.example.yaml /routeros/configs/config.yaml
	@cp configs/ad_domains.yaml /routeros/configs/ad_domains.yaml
	@chmod 644 /routeros/configs/*.yaml
	@echo "RouterOS directories created at /routeros/"
	@echo "Edit /routeros/configs/config.yaml before running"

# Test RouterOS container locally
.PHONY: docker-test-routeros
docker-test-routeros: docker-build-routeros
	@echo "Testing RouterOS container locally..."
	@mkdir -p ./test-routeros/{configs,certs,logs}
	@cp configs/config.example.yaml ./test-routeros/configs/config.yaml
	@cp configs/ad_domains.yaml ./test-routeros/configs/ad_domains.yaml || true
	@docker run -it --rm --platform=linux/arm64 -p 8888:8888 \
		--privileged \
		--ulimit nofile=65536:65536 \
		--ulimit nproc=32768:32768 \
		-v $(PWD)/test-routeros/configs:/app/configs \
		-v $(PWD)/test-routeros/certs:/app/certs \
		$(DOCKER_IMAGE):routeros

# Development helpers
.PHONY: dev
dev:
	@echo "Running in development mode with live reload..."
	@pkill -f $(BINARY_NAME) 2>/dev/null || true
	@$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH) && ./$(BINARY_NAME)

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
	@echo "  debug        - Build and run the proxy in debug mode"
	@echo "  clean        - Remove build artifacts"
	@echo "  deps         - Download dependencies"
	@echo "  update-deps  - Update dependencies"
	@echo "  test         - Run all tests"
	@echo "  ca-cert      - Generate CA certificate for HTTPS MITM"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  docker-build-routeros     - Build RouterOS 7 Docker image (ARM64)"
	@echo "  docker-run-routeros       - Run RouterOS 7 Docker container"
	@echo "  docker-compose-routeros   - Start RouterOS 7 with docker-compose"
	@echo "  docker-setup-routeros     - Setup RouterOS 7 deployment directories"
	@echo "  dev          - Run in development mode"
	@echo "  kill         - Kill running proxy"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"
	@echo "  help         - Show this help message"
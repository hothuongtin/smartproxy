# SmartProxy Development Guide

## Development Environment

### System Requirements
- Go 1.21 or higher
- Make
- Git
- Docker (optional)
- golangci-lint (for linting)

### Installing Go
```bash
# macOS
brew install go

# Linux
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Windows
# Download installer from https://go.dev/dl/
```

### Clone Repository
```bash
git clone https://github.com/hothuongtin/smartproxy.git
cd smartproxy
```

### Install Dependencies
```bash
make deps
```

## Project Structure

```
smartproxy/
├── main.go              # Entry point and main logic
├── config.go            # Configuration handling
├── configs/
│   ├── config.yaml          # Main configuration
│   ├── config.example.yaml  # Example configuration
│   └── ad_domains.yaml      # Ad blocking domains
├── certs/               # CA certificates (generated)
├── build/               # Build output
├── docker/              # Docker files
│   ├── Dockerfile           # Production image
│   ├── Dockerfile.scratch   # Minimal image
│   └── docker-compose.yml   # Docker Compose config
├── scripts/             # Utility scripts
│   ├── test/               # Test scripts
│   └── generate_ca.sh      # CA certificate generator
├── docs/                # Documentation
├── Makefile            # Build automation
├── go.mod              # Go module definition
└── go.sum              # Dependency checksums
```

## Development Workflow

### Basic Development Cycle
```bash
# 1. Make code changes
vim main.go

# 2. Build and test
make build

# 3. Run with hot reload
make dev

# 4. Test functionality
make test
```

### Live Reload with Air
```bash
# Install air
go install github.com/air-verse/air@latest

# Create .air.toml
air init

# Run with live reload
air
```

### Debugging

#### With VSCode
```json
// .vscode/launch.json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug SmartProxy",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}",
      "args": [],
      "env": {
        "SMARTPROXY_CONFIG": "configs/config.yaml"
      }
    }
  ]
}
```

#### With Delve
```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug
dlv debug --headless --listen=:2345 --api-version=2
```

## Build Instructions

### Local Development
```bash
# Quick build and run
make dev

# Build binary only
make build

# Run existing binary
make run

# Clean and rebuild
make clean build
```

### Production Build
```bash
# Build optimized binary
make build-prod

# Build for all platforms
make build-all

# Build specific platform
GOOS=linux GOARCH=amd64 make build
```

### Docker Build
```bash
# Build Docker image
make docker-build

# Run Docker container
make docker-run

# Build minimal scratch image
docker build -f docker/Dockerfile.scratch -t smartproxy:scratch .
```

## Testing

### Unit Tests
```go
// config_test.go
func TestLoadConfig(t *testing.T) {
    config, err := LoadConfig("testdata/config.yaml")
    if err != nil {
        t.Fatalf("Failed to load config: %v", err)
    }
    
    if config.Server.HTTPPort != 8888 {
        t.Errorf("Expected port 8888, got %d", config.Server.HTTPPort)
    }
}
```

### Integration Tests
```bash
# Run all tests
make test

# Run specific test script
./scripts/test/test_proxy.sh
./scripts/test/test_https.sh
./scripts/test/test_auth.sh
```

### Benchmark Tests
```go
func BenchmarkIsStaticFile(b *testing.B) {
    urls := []string{
        "http://example.com/script.js",
        "http://example.com/image.png",
        "http://example.com/page.html",
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        isStaticFile(urls[i%len(urls)])
    }
}
```

### Load Testing
```bash
# Using hey
go install github.com/rakyll/hey@latest
hey -n 10000 -c 100 -x http://localhost:8888 http://httpbin.org/get

# Using wrk
wrk -t12 -c400 -d30s --latency http://localhost:8888

# Using Apache Bench
ab -n 10000 -c 100 -X localhost:8888 http://httpbin.org/get
```

## Code Conventions

### Formatting
```bash
# Format code
make fmt

# Or manually
gofmt -s -w .
go fmt ./...
```

### Linting
```bash
# Run linter
make lint

# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

### Naming Conventions
- Functions: `camelCase` for private, `PascalCase` for exported
- Constants: `UPPER_SNAKE_CASE` or `PascalCase`
- Files: `snake_case.go`
- Packages: lowercase, no underscores

### Comments
```go
// Package main implements a high-performance HTTP/HTTPS proxy
package main

// Config represents the application configuration
type Config struct {
    // Server contains server-specific settings
    Server ServerConfig `yaml:"server"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
    // Implementation
}
```

## Architecture Overview

### Core Components

1. **main.go** - Application entry point and core proxy logic
   - HTTP/HTTPS request handling
   - Routing decisions
   - Transport management

2. **config.go** - Configuration management
   - YAML parsing
   - Validation
   - Default values

3. **Transport Layer** - Connection handling
   - Direct transport for static/CDN
   - HTTP proxy transport
   - SOCKS5 proxy transport

### Request Flow

```
Client Request
    ↓
Authentication Check
    ↓
Parse Upstream from Auth
    ↓
Routing Decision
    ├─ Ad Domain? → Block (204)
    ├─ Static File? → Direct
    ├─ CDN Domain? → Direct
    └─ Other → Upstream Proxy
```

## Contributing Guidelines

### Fork and Clone
```bash
# Fork on GitHub
# Clone your fork
git clone https://github.com/yourusername/smartproxy.git
cd smartproxy
git remote add upstream https://github.com/hothuongtin/smartproxy.git
```

### Create Branch
```bash
git checkout -b feature/my-feature
```

### Commit Conventions
```bash
# Format: <type>(<scope>): <subject>
git commit -m "feat(proxy): add WebSocket support"
git commit -m "fix(config): handle missing upstream proxy"
git commit -m "docs(readme): update installation instructions"
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance tasks

### Pull Request Process
1. Update documentation
2. Add tests for new features
3. Ensure all tests pass
4. Update CHANGELOG.md
5. Submit PR with clear description

## Debugging and Profiling

### Enable Debug Logging
```yaml
# config.yaml
logging:
  level: debug
```

### Performance Profiling
```go
import _ "net/http/pprof"

// In main()
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// Usage
go tool pprof http://localhost:6060/debug/pprof/heap
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

### Memory Analysis
```bash
# Run with memory profiling
go run -memprofile=mem.prof main.go

# Analyze
go tool pprof mem.prof
```

## Common Development Tasks

### Adding New Features

1. **Plan the feature**
   - Define requirements
   - Consider configuration needs
   - Plan tests

2. **Implement incrementally**
   - Start with core functionality
   - Add configuration support
   - Add logging and metrics
   - Write tests

3. **Document thoroughly**
   - Update code comments
   - Update configuration examples
   - Update user documentation

### Updating Dependencies
```bash
# Update all dependencies
make update-deps

# Update specific dependency
go get -u github.com/package/name

# Verify and tidy
go mod verify
go mod tidy
```

### Release Process
```bash
# Tag version
git tag -a v1.2.3 -m "Release version 1.2.3"

# Build release binaries
make release

# Push tag
git push origin v1.2.3
```

## Troubleshooting Development Issues

### Build Errors
```bash
# Clean build cache
go clean -cache
go clean -modcache

# Rebuild
make clean
make build
```

### Import Issues
```bash
# Verify module
go mod verify

# Download missing dependencies
go mod download

# Update go.sum
go mod tidy
```

### Port Conflicts
```bash
# Find process using port
lsof -i :8888  # macOS/Linux
netstat -ano | findstr :8888  # Windows

# Kill existing proxy
make kill
```

### Test Failures
```bash
# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v -run TestFunctionName

# Debug test
dlv test
```

## Performance Optimization Tips

1. **Profile First** - Don't optimize without data
2. **Minimize Allocations** - Reuse objects where possible
3. **Concurrent Safe** - Use sync primitives correctly
4. **Buffer Pools** - Reuse buffers for I/O operations
5. **Connection Pooling** - Reuse HTTP connections

## Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)
- [Awesome Go](https://awesome-go.com/)
- [Go Performance Tips](https://github.com/golang/go/wiki/Performance)
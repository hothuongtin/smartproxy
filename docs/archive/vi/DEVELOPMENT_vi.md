# Hướng dẫn phát triển SmartProxy

## Môi trường phát triển

### Yêu cầu hệ thống
- Go 1.21 hoặc mới hơn
- Make
- Git
- Docker (tùy chọn)
- golangci-lint (cho linting)

### Cài đặt Go
```bash
# macOS
brew install go

# Linux
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Windows
# Download installer từ https://go.dev/dl/
```

### Clone repository
```bash
git clone https://github.com/hothuongtin/smartproxy.git
cd smartproxy
```

### Cài đặt dependencies
```bash
make deps
```

## Cấu trúc project

```
smartproxy/
├── main.go              # Entry point và logic chính
├── config.go            # Xử lý cấu hình
├── config.yaml          # File cấu hình chính
├── config.example.yaml  # Cấu hình mẫu
├── ad_domains.yaml      # Danh sách domain quảng cáo
├── Makefile            # Build automation
├── go.mod              # Go module definition
├── go.sum              # Dependency checksums
├── Dockerfile          # Production Docker image
├── Dockerfile.scratch  # Minimal Docker image
├── docker-compose.yml  # Docker Compose config
├── .gitignore         # Git ignore rules
├── .dockerignore      # Docker ignore rules
├── test_*.sh          # Test scripts
├── generate_ca.sh     # CA certificate generator
└── docs/              # Documentation
    ├── README_vi.md
    ├── FAQ_vi.md
    └── ...
```

## Workflow phát triển

### 1. Development cycle cơ bản
```bash
# 1. Thay đổi code
vim main.go

# 2. Build và test
make build

# 3. Run với hot reload
make dev

# 4. Test chức năng
make test
```

### 2. Live reload với air
```bash
# Cài đặt air
go install github.com/air-verse/air@latest

# Tạo .air.toml
air init

# Run với live reload
air
```

### 3. Debugging

#### Với VSCode
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
        "SMARTPROXY_CONFIG": "config.yaml"
      }
    }
  ]
}
```

#### Với Delve
```bash
# Cài đặt delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug
dlv debug --headless --listen=:2345 --api-version=2
```

## Code conventions

### 1. Formatting
```bash
# Format code
make fmt

# Hoặc
gofmt -s -w .
```

### 2. Linting
```bash
# Run linter
make lint

# Cài đặt golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

### 3. Naming conventions
- Functions: `camelCase`
- Exported functions: `PascalCase`
- Constants: `UPPER_SNAKE_CASE` hoặc `PascalCase`
- Files: `snake_case.go`

### 4. Comments
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

## Testing

### 1. Unit tests
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

### 2. Integration tests
```bash
# Run all tests
make test

# Run specific test
./test_https.sh
```

### 3. Benchmark tests
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

### 4. Load testing
```bash
# Sử dụng hey
go install github.com/rakyll/hey@latest
hey -n 10000 -c 100 -x http://localhost:8888 http://httpbin.org/get

# Sử dụng wrk
wrk -t12 -c400 -d30s --latency http://localhost:8888
```

## Logging và Monitoring

### 1. Structured logging
```go
logger.Info("Request processed",
    "method", r.Method,
    "url", r.URL.String(),
    "status", resp.StatusCode,
    "duration", time.Since(start),
)
```

### 2. Debug logging
```go
if logger.Enabled(nil, slog.LevelDebug) {
    logger.Debug("Checking static file",
        "url", urlStr,
        "isStatic", result,
    )
}
```

### 3. Performance profiling
```go
import _ "net/http/pprof"

// Trong main()
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// Sử dụng
go tool pprof http://localhost:6060/debug/pprof/heap
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

## Contributing

### 1. Fork và clone
```bash
# Fork trên GitHub
# Clone fork của bạn
git clone https://github.com/hothuongtin/smartproxy.git
cd smartproxy
git remote add upstream https://github.com/originalowner/smartproxy.git
```

### 2. Tạo branch
```bash
git checkout -b feature/my-feature
```

### 3. Commit conventions
```bash
# Format: <type>(<scope>): <subject>
git commit -m "feat(proxy): add WebSocket support"
git commit -m "fix(config): handle missing upstream proxy"
git commit -m "docs(readme): update installation instructions"
```

Types:
- `feat`: Tính năng mới
- `fix`: Sửa lỗi
- `docs`: Thay đổi documentation
- `style`: Format, missing semi colons, etc
- `refactor`: Refactoring code
- `test`: Thêm tests
- `chore`: Maintain

### 4. Pull Request
```bash
# Push branch
git push origin feature/my-feature

# Tạo PR trên GitHub với template:
## Description
Mô tả thay đổi

## Type of change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
```

## Tools hữu ích

### 1. Code generation
```bash
# Generate mocks
mockgen -source=config.go -destination=mocks/config_mock.go

# Generate test coverage
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### 2. Dependency management
```bash
# Update dependencies
make update-deps

# Verify dependencies
go mod verify

# Tidy dependencies
go mod tidy
```

### 3. CI/CD
```yaml
# .github/workflows/ci.yml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: 1.21
    - run: make deps
    - run: make test
    - run: make lint
```

## Troubleshooting phát triển

### 1. Build errors
```bash
# Clean build cache
go clean -cache
go clean -modcache

# Rebuild
make clean
make build
```

### 2. Import issues
```bash
# Verify module
go mod verify

# Download missing deps
go mod download

# Update go.sum
go mod tidy
```

### 3. Port conflicts
```bash
# Find process using port
lsof -i :8888

# Kill process
make kill
```

## Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)
- [Awesome Go](https://awesome-go.com/)
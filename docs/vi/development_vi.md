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
# Tải bộ cài đặt từ https://go.dev/dl/
```

### Clone Repository
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
├── configs/
│   ├── config.yaml          # Cấu hình chính
│   ├── config.example.yaml  # Cấu hình mẫu
│   └── ad_domains.yaml      # Miền chặn quảng cáo
├── certs/               # Chứng chỉ CA (tạo ra)
├── build/               # Kết quả build
├── docker/              # Các file Docker
│   ├── Dockerfile           # Image production
│   ├── Dockerfile.scratch   # Image tối thiểu
│   └── docker-compose.yml   # Docker Compose config
├── scripts/             # Script tiện ích
│   ├── test/               # Script test
│   └── generate_ca.sh      # Tạo chứng chỉ CA
├── docs/                # Tài liệu
├── Makefile            # Tự động hóa build
├── go.mod              # Định nghĩa Go module
└── go.sum              # Checksum dependencies
```

## Workflow phát triển

### Chu kỳ phát triển cơ bản
```bash
# 1. Thay đổi code
vim main.go

# 2. Build và test
make build

# 3. Chạy với hot reload
make dev

# 4. Kiểm tra chức năng
make test
```

### Live reload với Air
```bash
# Cài đặt air
go install github.com/air-verse/air@latest

# Tạo .air.toml
air init

# Chạy với live reload
air
```

### Debugging

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
        "SMARTPROXY_CONFIG": "configs/config.yaml"
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

## Hướng dẫn build

### Phát triển cục bộ
```bash
# Build và chạy nhanh
make dev

# Chỉ build binary
make build

# Chạy binary hiện có
make run

# Làm sạch và build lại
make clean build
```

### Build production
```bash
# Build binary tối ưu
make build-prod

# Build cho tất cả nền tảng
make build-all

# Build nền tảng cụ thể
GOOS=linux GOARCH=amd64 make build
```

### Docker build
```bash
# Build Docker image
make docker-build

# Chạy Docker container
make docker-run

# Build image scratch tối thiểu
docker build -f docker/Dockerfile.scratch -t smartproxy:scratch .
```

## Testing

### Unit test
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

### Integration test
```bash
# Chạy tất cả test
make test

# Chạy script test cụ thể
./scripts/test/test_proxy.sh
./scripts/test/test_https.sh
./scripts/test/test_auth.sh
```

### Benchmark test
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

### Load testing
```bash
# Sử dụng hey
go install github.com/rakyll/hey@latest
hey -n 10000 -c 100 -x http://localhost:8888 http://httpbin.org/get

# Sử dụng wrk
wrk -t12 -c400 -d30s --latency http://localhost:8888

# Sử dụng Apache Bench
ab -n 10000 -c 100 -X localhost:8888 http://httpbin.org/get
```

## Quy ước code

### Định dạng
```bash
# Format code
make fmt

# Hoặc thủ công
gofmt -s -w .
go fmt ./...
```

### Linting
```bash
# Chạy linter
make lint

# Cài đặt golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

### Quy ước đặt tên
- Function: `camelCase` cho private, `PascalCase` cho exported
- Constant: `UPPER_SNAKE_CASE` hoặc `PascalCase`
- File: `snake_case.go`
- Package: chữ thường, không dấu gạch dưới

### Comment
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

## Tổng quan kiến trúc

### Các thành phần chính

1. **main.go** - Entry point ứng dụng và logic proxy chính
   - Xử lý yêu cầu HTTP/HTTPS
   - Quyết định định tuyến
   - Quản lý transport

2. **config.go** - Quản lý cấu hình
   - Phân tích YAML
   - Xác thực
   - Giá trị mặc định

3. **Transport Layer** - Xử lý kết nối
   - Direct transport cho static/CDN
   - HTTP proxy transport
   - SOCKS5 proxy transport

### Luồng yêu cầu

```
Yêu cầu client
    ↓
Kiểm tra xác thực
    ↓
Phân tích upstream từ auth
    ↓
Quyết định định tuyến
    ├─ Miền quảng cáo? → Chặn (204)
    ├─ File tĩnh? → Trực tiếp
    ├─ Miền CDN? → Trực tiếp
    └─ Khác → Upstream Proxy
```

## Hướng dẫn đóng góp

### Fork và clone
```bash
# Fork trên GitHub
# Clone fork của bạn
git clone https://github.com/yourusername/smartproxy.git
cd smartproxy
git remote add upstream https://github.com/hothuongtin/smartproxy.git
```

### Tạo branch
```bash
git checkout -b feature/my-feature
```

### Quy ước commit
```bash
# Định dạng: <type>(<scope>): <subject>
git commit -m "feat(proxy): add WebSocket support"
git commit -m "fix(config): handle missing upstream proxy"
git commit -m "docs(readme): update installation instructions"
```

Các loại:
- `feat`: Tính năng mới
- `fix`: Sửa lỗi
- `docs`: Thay đổi tài liệu
- `style`: Thay đổi code style
- `refactor`: Tái cấu trúc code
- `test`: Thêm test
- `chore`: Công việc bảo trì

### Quy trình pull request
1. Cập nhật tài liệu
2. Thêm test cho tính năng mới
3. Đảm bảo tất cả test pass
4. Cập nhật CHANGELOG.md
5. Gửi PR với mô tả rõ ràng

## Debug và profiling

### Bật debug logging
```yaml
# config.yaml
logging:
  level: debug
```

### Performance profiling
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

### Phân tích bộ nhớ
```bash
# Chạy với memory profiling
go run -memprofile=mem.prof main.go

# Phân tích
go tool pprof mem.prof
```

## Các tác vụ phát triển thường gặp

### Thêm tính năng mới

1. **Lên kế hoạch tính năng**
   - Xác định yêu cầu
   - Cân nhắc nhu cầu cấu hình
   - Lên kế hoạch test

2. **Triển khai tăng dần**
   - Bắt đầu với chức năng cốt lõi
   - Thêm hỗ trợ cấu hình
   - Thêm logging và metrics
   - Viết test

3. **Tài liệu đầy đủ**
   - Cập nhật code comment
   - Cập nhật ví dụ cấu hình
   - Cập nhật tài liệu người dùng

### Cập nhật dependencies
```bash
# Cập nhật tất cả dependencies
make update-deps

# Cập nhật dependency cụ thể
go get -u github.com/package/name

# Xác minh và dọn dẹp
go mod verify
go mod tidy
```

### Quy trình release
```bash
# Tag version
git tag -a v1.2.3 -m "Release version 1.2.3"

# Build release binary
make release

# Push tag
git push origin v1.2.3
```

## Khắc phục sự cố phát triển

### Lỗi build
```bash
# Làm sạch build cache
go clean -cache
go clean -modcache

# Build lại
make clean
make build
```

### Vấn đề import
```bash
# Xác minh module
go mod verify

# Tải dependencies thiếu
go mod download

# Cập nhật go.sum
go mod tidy
```

### Xung đột cổng
```bash
# Tìm process dùng cổng
lsof -i :8888  # macOS/Linux
netstat -ano | findstr :8888  # Windows

# Kill proxy hiện có
make kill
```

### Test thất bại
```bash
# Chạy test với output chi tiết
go test -v ./...

# Chạy test cụ thể
go test -v -run TestFunctionName

# Debug test
dlv test
```

## Mẹo tối ưu hiệu suất

1. **Profile trước** - Đừng tối ưu khi chưa có dữ liệu
2. **Giảm thiểu allocation** - Tái sử dụng object khi có thể
3. **Concurrent safe** - Sử dụng sync primitive đúng cách
4. **Buffer pool** - Tái sử dụng buffer cho hoạt động I/O
5. **Connection pooling** - Tái sử dụng kết nối HTTP

## Tài nguyên

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Proverbs](https://go-proverbs.github.io/)
- [Awesome Go](https://awesome-go.com/)
- [Go Performance Tips](https://github.com/golang/go/wiki/Performance)
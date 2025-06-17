# SmartProxy

Một proxy HTTP/HTTPS hiệu suất cao với khả năng định tuyến thông minh, chặn quảng cáo và sử dụng tài nguyên tối thiểu.

## Tính năng

- 🚀 **Hiệu suất cao**: Xử lý hàng ngàn kết nối đồng thời với connection pooling
- 🎯 **Định tuyến thông minh**: Kết nối trực tiếp cho file tĩnh và CDN
- 🚫 **Chặn quảng cáo**: Chặn domain quảng cáo và tracking với hiệu suất O(1)
- 🔒 **Hỗ trợ HTTPS**: Tùy chọn MITM để kiểm tra hoặc tunnel an toàn
- 🌈 **Log màu sắc**: Log có cấu trúc đẹp mắt với slogcolor
- 📦 **Docker image tối giản**: Image production chỉ ~15MB sử dụng distroless/scratch
- 🔧 **Cấu hình linh hoạt**: Cấu hình bằng YAML với hỗ trợ hot-reload

## Bắt đầu nhanh

### Sử dụng Make

```bash
# Build và chạy
make build
make run

# Hoặc một lệnh duy nhất
make dev
```

### Sử dụng Docker

```bash
# Sử dụng docker-compose (khuyến nghị)
docker-compose up -d

# Hoặc build và chạy thủ công
make docker-build
make docker-run
```

### Cấu hình

1. Sao chép file cấu hình mẫu:
```bash
cp config.example.yaml config.yaml
```

2. Cấu hình upstream proxy (BẮT BUỘC):
```yaml
upstream:
  proxy_url: "http://your-proxy:8080"
  username: "tùy chọn"
  password: "tùy chọn"
```

3. Chạy proxy:
```bash
make run
```

## Tùy chọn cấu hình

### Cài đặt cơ bản

```yaml
server:
  http_port: 8888              # Cổng lắng nghe proxy
  https_mitm: false            # Bật chặn HTTPS
  max_idle_conns: 10000        # Kích thước connection pool
  max_idle_conns_per_host: 100 # Giới hạn kết nối mỗi host
```

### Upstream Proxy (Bắt buộc)

```yaml
upstream:
  proxy_url: "http://proxy:8080"  # hoặc "socks5://127.0.0.1:1080"
  username: ""
  password: ""
```

### Chặn quảng cáo

```yaml
ad_blocking:
  enabled: true
  domains_file: "ad_domains.yaml"
```

## Hiệu suất

SmartProxy được tối ưu cho hoạt động hiệu suất cao:

- **Connection Pooling**: Tái sử dụng kết nối để giảm chi phí
- **Chặn quảng cáo O(1)**: Tra cứu hash map cho khớp domain tức thì
- **Định tuyến trực tiếp**: Bỏ qua upstream proxy cho nội dung tĩnh
- **Hỗ trợ HTTP/2**: Multiplexing để hiệu suất tốt hơn
- **Zero-Copy Operations**: Phân bổ bộ nhớ tối thiểu

### Benchmark

Với cài đặt mặc định:
- 10,000+ kết nối đồng thời
- 5,000+ requests/giây
- Độ trễ dưới mili giây cho kết nối trực tiếp
- ~50MB sử dụng bộ nhớ dưới tải

## Docker Images

Chúng tôi cung cấp nhiều tùy chọn Docker image:

### Distroless (Khuyến nghị)
- Kích thước: ~15MB
- Bảo mật: Không có shell, bề mặt tấn công tối thiểu
- Base: `gcr.io/distroless/static-debian12`

```bash
docker build -t smartproxy:latest .
```

### Scratch (Tối thiểu)
- Kích thước: ~12MB
- Bảo mật: Tối thiểu tuyệt đối
- Base: `scratch`

```bash
docker build -f Dockerfile.scratch -t smartproxy:scratch .
```

## Cấu hình HTTPS

### Chế độ Tunneling (Mặc định)
- Không có cảnh báo chứng chỉ
- Mã hóa end-to-end được duy trì
- Không cần cấu hình

### Chế độ MITM
Để kiểm tra HTTPS:

1. Tạo chứng chỉ CA:
```bash
make ca-cert
```

2. Bật trong cấu hình:
```yaml
server:
  https_mitm: true
  ca_cert: "certs/ca.crt"
  ca_key: "certs/ca.key"
```

3. Cài đặt CA trên thiết bị client

## Phát triển

### Yêu cầu
- Go 1.21+
- Make
- Docker (tùy chọn)

### Build

```bash
# Phát triển local
make dev

# Production build
make build

# Cross-platform builds
make build-all
```

### Testing

```bash
# Chạy tất cả test
make test

# Test chức năng cụ thể
./test_proxy.sh
./test_https.sh
```

### Chất lượng code

```bash
# Format code
make fmt

# Chạy linter
make lint
```

## Kiến trúc

SmartProxy sử dụng kiến trúc đơn giản nhưng hiệu quả:

- **Binary đơn**: Tất cả chức năng trong một file thực thi
- **Cấu hình YAML**: Dễ dàng quản lý cài đặt
- **Transport linh hoạt**: Hỗ trợ HTTP/SOCKS5 upstreams
- **Graceful Shutdown**: Dọn dẹp kết nối đúng cách

## Đóng góp

1. Fork repository
2. Tạo nhánh tính năng
3. Commit thay đổi
4. Push lên nhánh
5. Tạo Pull Request

## Giấy phép

MIT License - xem file LICENSE để biết chi tiết

## Khắc phục sự cố

### Cổng đã được sử dụng
```bash
make kill  # Tắt proxy đang chạy
make run   # Khởi động lại
```

### Lỗi chứng chỉ
- Đảm bảo chứng chỉ CA được cài trên client
- Kiểm tra ngày hết hạn chứng chỉ
- Xác minh MITM được bật trong cấu hình

### Vấn đề hiệu suất
- Tăng `max_idle_conns` cho nhiều kết nối hơn
- Kiểm tra hiệu suất upstream proxy
- Giám sát tài nguyên hệ thống

## Hỗ trợ

- Issues: [GitHub Issues](https://github.com/yourusername/smartproxy/issues)
- Tài liệu: Xem thư mục `docs/`
- FAQ: Xem `FAQ_vi.md`
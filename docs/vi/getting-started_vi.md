# Bắt đầu với SmartProxy

## Tổng quan

SmartProxy là một proxy HTTP/HTTPS hiệu suất cao với khả năng định tuyến thông minh, chặn quảng cáo và sử dụng tài nguyên tối thiểu. Hướng dẫn này sẽ giúp bạn cài đặt và chạy nhanh chóng.

## Yêu cầu hệ thống

- **Hệ điều hành**: Linux, macOS, hoặc Windows
- **Go**: 1.21+ (để build từ source)
- **Docker**: Tùy chọn, cho triển khai container
- **Bộ nhớ**: Tối thiểu 64MB, khuyến nghị 256MB+
- **CPU**: Bất kỳ CPU hiện đại nào

## Phương pháp cài đặt

### Phương pháp 1: Sử dụng file binary có sẵn

Tải phiên bản mới nhất cho nền tảng của bạn từ [trang releases](https://github.com/hothuongtin/smartproxy/releases).

```bash
# Tải và giải nén (ví dụ cho Linux)
wget https://github.com/hothuongtin/smartproxy/releases/download/vX.X.X/smartproxy-linux-amd64.tar.gz
tar -xzf smartproxy-linux-amd64.tar.gz
chmod +x smartproxy

# Chạy
./smartproxy
```

### Phương pháp 2: Sử dụng Docker (Khuyến nghị)

```bash
# Sử dụng docker-compose (khuyến nghị)
cd docker && docker-compose up -d

# Hoặc build và chạy thủ công
docker build -t smartproxy:latest .
docker run -d \
  --name smartproxy \
  -p 8888:8888 \
  -v $(pwd)/configs/config.yaml:/app/config.yaml:ro \
  smartproxy:latest
```

### Phương pháp 3: Build từ source code

```bash
# Clone repository
git clone https://github.com/hothuongtin/smartproxy.git
cd smartproxy

# Build bằng Make
make build

# Chạy
make run

# Hoặc một lệnh cho phát triển
make dev
```

## Hướng dẫn nhanh

### Bước 1: Cấu hình

Sao chép file cấu hình mẫu:

```bash
cp configs/config.example.yaml configs/config.yaml
```

Cấu hình cơ bản (`configs/config.yaml`):

```yaml
server:
  http_port: 8888              # Cổng lắng nghe proxy
  https_mitm: false            # Bật chặn HTTPS (cần chứng chỉ CA)
  max_idle_conns: 10000        # Kích thước connection pool
  max_idle_conns_per_host: 100 # Giới hạn connection mỗi host

ad_blocking:
  enabled: true
  domains_file: "configs/ad_domains.yaml"

direct_patterns:
  static_files:
    - .js
    - .css
    - .jpg
    - .png
    - .gif
    - .ico
    - .woff
    - .woff2
  
  cdn_domains:
    - cloudflare.com
    - cdn.jsdelivr.net
    - cdnjs.cloudflare.com
    - unpkg.com

logging:
  level: info  # debug, info, warn, error
  colored: true
```

### Bước 2: Cấu hình xác thực proxy

SmartProxy sử dụng xác thực thông minh để cấu hình động các upstream proxy:

```bash
# Định dạng
Username: <schema>  # http hoặc socks5
Password: <base64-encoded-upstream>

# Ví dụ: HTTP proxy không có xác thực
echo -n "proxy.example.com:8080" | base64
# Kết quả: cHJveHkuZXhhbXBsZS5jb206ODA4MA==

# Ví dụ: SOCKS5 proxy với xác thực
echo -n "socks.example.com:1080:user:pass" | base64
# Kết quả: c29ja3MuZXhhbXBsZS5jb206MTA4MDp1c2VyOnBhc3M=
```

### Bước 3: Chạy SmartProxy

```bash
# Sử dụng Make
make run

# Hoặc trực tiếp
./smartproxy

# Với đường dẫn config tùy chỉnh
SMARTPROXY_CONFIG=/path/to/config.yaml ./smartproxy
```

### Bước 4: Cấu hình client

#### Cấu hình trình duyệt

Cấu hình trình duyệt sử dụng `http://localhost:8888` làm proxy HTTP/HTTPS.

Ví dụ với xác thực:
```
Proxy: http://localhost:8888
Username: http
Password: cHJveHkuZXhhbXBsZS5jb206ODA4MA==
```

#### Dòng lệnh

```bash
# Sử dụng curl
curl -x http://http:cHJveHkuZXhhbXBsZS5jb206ODA4MA==@localhost:8888 http://ipinfo.io

# Sử dụng biến môi trường
export http_proxy=http://http:cHJveHkuZXhhbXBsZS5jb206ODA4MA==@localhost:8888
export https_proxy=http://http:cHJveHkuZXhhbXBsZS5jb206ODA4MA==@localhost:8888
curl http://ipinfo.io
```

## Cấu hình HTTPS

### Chế độ Tunneling (Mặc định)

Theo mặc định, SmartProxy tunnel lưu lượng HTTPS mà không giải mã:

- Không có cảnh báo chứng chỉ
- Mã hóa end-to-end được duy trì
- Không cần cấu hình

### Chế độ MITM (Nâng cao)

Để kiểm tra lưu lượng HTTPS và bật các tính năng như chặn quảng cáo trên các trang HTTPS:

#### Bước 1: Tạo chứng chỉ CA

```bash
make ca-cert
# Hoặc thủ công
./scripts/generate_ca.sh
```

Lệnh này tạo:
- `certs/ca.crt` - Chứng chỉ CA (cài đặt trên client)
- `certs/ca.key` - Khóa riêng (giữ bảo mật)

#### Bước 2: Bật MITM trong cấu hình

```yaml
server:
  https_mitm: true
  ca_cert: "certs/ca.crt"
  ca_key: "certs/ca.key"
```

#### Bước 3: Cài đặt chứng chỉ CA trên client

**macOS:**
1. Double-click `certs/ca.crt`
2. Thêm vào System keychain
3. Trust cho SSL

**Windows:**
1. Double-click `certs/ca.crt`
2. Cài đặt vào "Trusted Root Certification Authorities"

**Linux:**
```bash
sudo cp certs/ca.crt /usr/local/share/ca-certificates/smartproxy.crt
sudo update-ca-certificates
```

**iOS:**
1. Email hoặc AirDrop `ca.crt` đến thiết bị
2. Cài đặt profile
3. Vào Settings > General > About > Certificate Trust Settings
4. Bật cho SmartProxy CA

**Android:**
1. Copy `ca.crt` vào thiết bị
2. Settings > Security > Install from storage
3. Chọn "CA certificate"

## Triển khai Docker

### Sử dụng Docker Compose (Khuyến nghị)

```yaml
version: '3.8'

services:
  smartproxy:
    image: smartproxy:latest
    container_name: smartproxy
    restart: unless-stopped
    ports:
      - "8888:8888"
    volumes:
      - ./configs/config.yaml:/app/config.yaml:ro
      - ./configs/ad_domains.yaml:/app/ad_domains.yaml:ro
      - ./certs:/app/certs:ro  # Chỉ khi sử dụng MITM
    environment:
      - SMARTPROXY_CONFIG=/app/config.yaml
    security_opt:
      - no-new-privileges:true
    read_only: true
    tmpfs:
      - /tmp
```

### Tùy chọn Docker Image

**Distroless (Khuyến nghị):**
- Kích thước: ~15MB
- Bảo mật: Không có shell, bề mặt tấn công tối thiểu
- Base: `gcr.io/distroless/static-debian12`

```bash
docker build -f docker/Dockerfile -t smartproxy:latest .
```

**Scratch (Tối thiểu):**
- Kích thước: ~12MB
- Bảo mật: Tối thiểu tuyệt đối
- Base: `scratch`

```bash
docker build -f docker/Dockerfile.scratch -t smartproxy:scratch .
```

## Xác minh cài đặt

Kiểm tra cài đặt proxy:

```bash
# Test HTTP
curl -x http://localhost:8888 http://httpbin.org/get

# Test HTTPS
curl -x http://localhost:8888 https://httpbin.org/get

# Test với xác thực
curl -x http://http:cHJveHkuZXhhbXBsZS5jb206ODA4MA==@localhost:8888 http://ipinfo.io
```

## Các vấn đề thường gặp

### Cổng đã được sử dụng

```bash
# Kill proxy hiện tại
make kill  # hoặc pkill -f smartproxy

# Khởi động lại
make run
```

### Lỗi xác thực

Luôn sử dụng `http://` cho URL proxy, ngay cả khi truy cập các trang HTTPS:

```bash
# ✅ ĐÚNG
curl -x http://http:PASSWORD@localhost:8888 https://example.com

# ❌ SAI - Không dùng https:// cho URL proxy
curl -x https://http:PASSWORD@localhost:8888 https://example.com
```

### Lỗi chứng chỉ trong chế độ MITM

- Đảm bảo chứng chỉ CA được cài đặt và tin cậy đúng cách
- Kiểm tra ngày hết hạn chứng chỉ
- Xác minh MITM đã được bật trong config

## Các bước tiếp theo

- [Hướng dẫn cấu hình](configuration_vi.md) - Các tùy chọn cấu hình chi tiết
- [Tính năng](features_vi.md) - Tìm hiểu về định tuyến thông minh và chặn quảng cáo
- [Xác thực](authentication_vi.md) - Cài đặt xác thực nâng cao
- [Khắc phục sự cố](troubleshooting_vi.md) - Các vấn đề thường gặp và giải pháp
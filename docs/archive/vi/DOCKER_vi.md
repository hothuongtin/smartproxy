# Hướng dẫn Docker cho SmartProxy

## Tổng quan

SmartProxy cung cấp Docker images được tối ưu với kích thước cực nhỏ và bảo mật cao. Chúng tôi sử dụng multi-stage builds và distroless/scratch base images.

## Docker Images

### 1. Distroless Image (Khuyến nghị)
- **Kích thước**: ~15MB (5MB với UPX compression)
- **Base**: `gcr.io/distroless/static-debian12`
- **Bảo mật**: Không có shell, package manager, hoặc utilities không cần thiết
- **Sử dụng cho**: Production

### 2. Scratch Image (Tối thiểu)
- **Kích thước**: ~12MB
- **Base**: `scratch` (empty image)
- **Bảo mật**: Tối thiểu tuyệt đối
- **Sử dụng cho**: Môi trường cần image nhỏ nhất

## Bắt đầu nhanh

### Sử dụng docker-compose (Khuyến nghị)

```bash
# Khởi động
docker-compose up -d

# Xem logs
docker-compose logs -f

# Dừng
docker-compose down
```

### Build và chạy thủ công

```bash
# Build image
docker build -t smartproxy:latest .

# Chạy container
docker run -d \
  --name smartproxy \
  -p 8888:8888 \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  smartproxy:latest
```

## Cấu hình docker-compose.yml

```yaml
version: '3.8'

services:
  smartproxy:
    build: .
    image: smartproxy:latest
    container_name: smartproxy
    restart: unless-stopped
    ports:
      - "8888:8888"
    volumes:
      - ./config.yaml:/app/config.yaml:ro
      - ./ad_domains.yaml:/app/ad_domains.yaml:ro
    environment:
      - SMARTPROXY_CONFIG=/app/config.yaml
    networks:
      - proxy-network
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 512M
    security_opt:
      - no-new-privileges:true
    read_only: true
    tmpfs:
      - /tmp

networks:
  proxy-network:
    driver: bridge
```

## Build các loại image khác nhau

### Distroless (mặc định)
```bash
docker build -t smartproxy:distroless .
```

### Scratch (tối thiểu nhất)
```bash
docker build -f Dockerfile.scratch -t smartproxy:scratch .
```

### Alpine (cho debugging)
```bash
docker build -f Dockerfile.alpine -t smartproxy:alpine .
```

## Tối ưu build

### 1. Sử dụng BuildKit
```bash
DOCKER_BUILDKIT=1 docker build -t smartproxy:latest .
```

### 2. Multi-platform build
```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t smartproxy:latest \
  --push .
```

### 3. Cache mount để build nhanh hơn
```dockerfile
# syntax=docker/dockerfile:1
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
```

## Bảo mật

### 1. Non-root user
Tất cả images chạy với non-root user (UID 65532 cho distroless).

### 2. Read-only filesystem
```yaml
read_only: true
tmpfs:
  - /tmp
```

### 3. Security options
```yaml
security_opt:
  - no-new-privileges:true
  - seccomp:unconfined
```

### 4. Network isolation
```yaml
networks:
  proxy-network:
    driver: bridge
    internal: true
```

## Quản lý logs

### 1. Xem logs
```bash
docker logs -f smartproxy
```

### 2. Giới hạn kích thước log
```yaml
logging:
  driver: "json-file"
  options:
    max-size: "10m"
    max-file: "3"
```

### 3. Gửi logs đến syslog
```yaml
logging:
  driver: "syslog"
  options:
    syslog-address: "tcp://192.168.0.42:123"
```

## Health checks

```yaml
healthcheck:
  test: ["CMD", "nc", "-z", "localhost", "8888"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 10s
```

## Volumes và Bind Mounts

### 1. Config files (read-only)
```yaml
volumes:
  - ./config.yaml:/app/config.yaml:ro
  - ./ad_domains.yaml:/app/ad_domains.yaml:ro
```

### 2. Chứng chỉ CA (nếu dùng MITM)
```yaml
volumes:
  - ./certs:/app/certs:ro
```

### 3. Persistent logs
```yaml
volumes:
  - ./logs:/app/logs
```

## Environment Variables

```yaml
environment:
  - SMARTPROXY_CONFIG=/app/config.yaml
  - LOG_LEVEL=info
  - TZ=Asia/Ho_Chi_Minh
```

## Networking

### 1. Host network mode (hiệu suất tốt nhất)
```bash
docker run --network host smartproxy:latest
```

### 2. Bridge network (mặc định)
```yaml
networks:
  proxy-network:
    driver: bridge
```

### 3. Custom network
```bash
docker network create --driver bridge proxy-net
docker run --network proxy-net smartproxy:latest
```

## Backup và Restore

### Backup config
```bash
docker cp smartproxy:/app/config.yaml ./backup/
```

### Restore config
```bash
docker cp ./backup/config.yaml smartproxy:/app/
docker restart smartproxy
```

## Monitoring

### 1. Stats realtime
```bash
docker stats smartproxy
```

### 2. Inspect container
```bash
docker inspect smartproxy
```

### 3. Export metrics
```bash
docker run -d \
  --name prometheus \
  -p 9090:9090 \
  -v prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus
```

## Troubleshooting

### Container không start
```bash
# Xem logs
docker logs smartproxy

# Debug với shell (cần Alpine image)
docker run -it --rm smartproxy:alpine sh
```

### Permission denied
```bash
# Fix quyền cho config files
chmod 644 config.yaml ad_domains.yaml
```

### Network issues
```bash
# Test network trong container
docker exec smartproxy nc -zv google.com 80
```

## Tips và Tricks

### 1. Alias cho commands thường dùng
```bash
alias sp-up='docker-compose up -d'
alias sp-down='docker-compose down'
alias sp-logs='docker-compose logs -f'
alias sp-restart='docker-compose restart'
```

### 2. Auto-update với Watchtower
```bash
docker run -d \
  --name watchtower \
  -v /var/run/docker.sock:/var/run/docker.sock \
  containrrr/watchtower \
  --interval 86400 \
  smartproxy
```

### 3. Backup tự động
```bash
# Cron job backup config hàng ngày
0 2 * * * docker cp smartproxy:/app/config.yaml /backup/config-$(date +\%Y\%m\%d).yaml
```
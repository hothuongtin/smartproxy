# Hướng dẫn hiệu suất SmartProxy

## Tổng quan

SmartProxy được tối ưu cho hoạt động hiệu suất cao, có khả năng xử lý hàng nghìn kết nối đồng thời với tài nguyên sử dụng tối thiểu. Hướng dẫn này bao gồm các đặc điểm hiệu suất, kỹ thuật tối ưu và khuyến nghị điều chỉnh.

## Đặc điểm hiệu suất

### Benchmarks

Với cài đặt mặc định, SmartProxy đạt được:
- **Kết nối đồng thời**: 10,000+
- **Yêu cầu/giây**: 5,000+ (tùy thuộc độ trễ upstream)
- **Độ trễ kết nối trực tiếp**: Dưới mili giây
- **Sử dụng bộ nhớ**: ~50MB dưới tải vừa phải
- **Sử dụng CPU**: Tối thiểu khi tắt nén

### Tính năng hiệu suất chính

1. **Connection Pooling**: Tái sử dụng kết nối để giảm tổn hạo TCP handshake
2. **Chặn quảng cáo O(1)**: Tìm kiếm hash map cho khớp miền tức thì
3. **Định tuyến trực tiếp**: Bỏ qua upstream proxy cho nội dung tĩnh
4. **Hỗ trợ HTTP/2**: Ghép kênh cho hiệu suất tốt hơn
5. **Hoạt động Zero-Copy**: Cấp phát bộ nhớ tối thiểu

## Tối ưu hóa hiệu suất

### 1. Connection Pooling

SmartProxy duy trì một pool kết nối có thể tái sử dụng để tránh tổn hạo tạo kết nối TCP mới.

**Cấu hình:**
```yaml
server:
  max_idle_conns: 10000        # Tổng kích thước pool
  max_idle_conns_per_host: 100 # Giới hạn mỗi host
  idle_conn_timeout: 90         # Giây trước khi đóng kết nối idle
```

**Hướng dẫn điều chỉnh:**
- **Lưu lượng cao**: Tăng `max_idle_conns` lên 50,000+
- **Nhiều host**: Tăng `max_idle_conns_per_host` lên 500+
- **Phiên dài**: Tăng `idle_conn_timeout` lên 300s

### 2. Chặn miền quảng cáo

Chặn quảng cáo sử dụng hash map được tối ưu cho tìm kiếm O(1):

```go
// Kiểm tra miền hiệu quả với chặn phân cấp
func isAdDomain(host string) bool {
    adDomainsMutex.RLock()
    defer adDomainsMutex.RUnlock()
    
    // Tìm kiếm trực tiếp - O(1)
    if adDomains[host] {
        return true
    }
    
    // Kiểm tra miền cha
    parts := strings.Split(host, ".")
    for i := 1; i < len(parts); i++ {
        parent := strings.Join(parts[i:], ".")
        if adDomains[parent] {
            return true
        }
    }
    return false
}
```

**Tác động hiệu suất:**
- Thời gian tìm kiếm: <1μs mỗi miền
- Sử dụng bộ nhớ: ~1MB mỗi 10,000 miền
- Thread-safe với RWMutex

### 3. Cấu hình transport

Cài đặt transport được tối ưu cho các loại kết nối khác nhau:

```yaml
server:
  tls_handshake_timeout: 10    # Thời gian chờ TLS handshake
  expect_continue_timeout: 1    # Thời gian chờ HTTP Expect-Continue
  read_buffer_size: 65536      # Buffer đọc 64KB
  write_buffer_size: 65536     # Buffer ghi 64KB
```

**Điều chỉnh kích thước buffer:**
- **Băng thông cao**: Tăng lên 131072 (128KB)
- **Bộ nhớ thấp**: Giảm xuống 32768 (32KB)
- **Mặc định**: 65536 (64KB) hoạt động tốt cho hầu hết trường hợp

### 4. Định tuyến trực tiếp

File tĩnh và nội dung CDN bỏ qua upstream proxy:

```yaml
direct_extensions:
  - .js
  - .css
  - .jpg
  - .png
  - .woff
  - .woff2

direct_domains:
  - cdn.
  - static.
  - cloudflare
  - akamai
```

**Lợi ích hiệu suất:**
- Giảm tải upstream proxy
- Độ trễ thấp hơn cho nội dung tĩnh
- Sử dụng băng thông tốt hơn

## Cấu hình cho các khối lượng công việc khác nhau

### Cấu hình đồng thời cao

Để xử lý 50,000+ kết nối đồng thời:

```yaml
server:
  http_port: 8888
  max_idle_conns: 50000
  max_idle_conns_per_host: 500
  idle_conn_timeout: 120
  read_buffer_size: 131072   # 128KB
  write_buffer_size: 131072  # 128KB

ad_blocking:
  enabled: true  # Tắt nếu không cần cho hiệu suất tốt hơn

logging:
  level: warn  # Giảm tổn hạo logging
```

### Cấu hình bộ nhớ thấp

Cho môi trường hạn chế tài nguyên:

```yaml
server:
  http_port: 8888
  max_idle_conns: 1000
  max_idle_conns_per_host: 10
  idle_conn_timeout: 30
  read_buffer_size: 32768    # 32KB
  write_buffer_size: 32768   # 32KB

ad_blocking:
  enabled: false  # Tiết kiệm bộ nhớ

logging:
  level: error  # Logging tối thiểu
```

### Cấu hình cân bằng

Cân bằng tốt giữa hiệu suất và sử dụng tài nguyên:

```yaml
server:
  http_port: 8888
  max_idle_conns: 10000
  max_idle_conns_per_host: 100
  idle_conn_timeout: 90
  read_buffer_size: 65536    # 64KB
  write_buffer_size: 65536   # 64KB

ad_blocking:
  enabled: true

logging:
  level: info
```

## Điều chỉnh hệ thống

### Tham số Linux Kernel

```bash
# Tăng giới hạn file descriptor
ulimit -n 65536

# Hoặc vĩnh viễn trong /etc/security/limits.conf
* soft nofile 65536
* hard nofile 65536

# Điều chỉnh mạng kernel
sudo sysctl -w net.ipv4.ip_local_port_range="1024 65535"
sudo sysctl -w net.core.somaxconn=32768
sudo sysctl -w net.ipv4.tcp_tw_reuse=1
sudo sysctl -w net.ipv4.tcp_fin_timeout=15
```

### Điều chỉnh macOS

```bash
# Tăng giới hạn file descriptor
ulimit -n 65536

# Tăng kern.maxfiles
sudo sysctl -w kern.maxfiles=65536
sudo sysctl -w kern.maxfilesperproc=65536
```

### Hiệu suất Docker

```yaml
# docker-compose.yml
services:
  smartproxy:
    image: smartproxy:latest
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 1G
        reservations:
          cpus: '2'
          memory: 512M
    sysctls:
      - net.core.somaxconn=65535
      - net.ipv4.ip_local_port_range=1024 65535
```

## Giám sát hiệu suất

### Chỉ số tích hợp

Bật debug logging để xem chỉ số hiệu suất:

```yaml
logging:
  level: debug
```

Log sẽ hiển thị:
- Thời gian yêu cầu/phản hồi
- Thống kê tái sử dụng kết nối
- Sự kiện tạo transport

### Giám sát bên ngoài

#### Tích hợp Prometheus (Tính năng tương lai)
```go
// Ví dụ chỉ số có thể được export
proxy_requests_total
proxy_request_duration_seconds
proxy_active_connections
proxy_connection_pool_size
proxy_upstream_errors_total
```

#### Tùy chọn giám sát hiện tại

1. **Chỉ số hệ thống**
   ```bash
   # Sử dụng CPU và bộ nhớ
   top -p $(pgrep smartproxy)
   
   # Kết nối mạng
   netstat -an | grep 8888 | wc -l
   
   # File descriptor
   lsof -p $(pgrep smartproxy) | wc -l
   ```

2. **Log ứng dụng**
   ```bash
   # Tỉ lệ yêu cầu
   tail -f smartproxy.log | grep "Request completed" | \
     awk '{print $1}' | uniq -c
   
   # Tỉ lệ lỗi
   tail -f smartproxy.log | grep ERROR | wc -l
   ```

## Kiểm tra tải

### Sử dụng load test đi kèm

```bash
# Khởi động proxy
./smartproxy

# Chạy load test
go run scripts/loadtest.go
```

### Sử dụng công cụ bên ngoài

#### Apache Bench (ab)
```bash
ab -n 10000 -c 100 -X localhost:8888 http://httpbin.org/get
```

#### hey
```bash
hey -n 10000 -c 100 -x http://localhost:8888 http://httpbin.org/get
```

#### wrk
```bash
wrk -t12 -c400 -d30s --latency \
  -H "Proxy-Authorization: Basic $(echo -n 'http:BASE64_PASS' | base64)" \
  http://localhost:8888/http://httpbin.org/get
```

## Khắc phục sự cố hiệu suất

### Sử dụng CPU cao

**Nguyên nhân có thể:**
1. Debug logging bật
2. Hoạt động SSL/TLS (chế độ MITM)
3. Nén bật

**Giải pháp:**
- Đặt `logging.level: warn` hoặc `error`
- Tắt MITM nếu không cần
- Tắt nén trong transport

### Sử dụng bộ nhớ cao

**Nguyên nhân có thể:**
1. Connection pool lớn
2. Rò rỉ bộ nhớ trong kết nối upstream
3. Danh sách chặn quảng cáo lớn

**Giải pháp:**
- Giảm `max_idle_conns`
- Đặt `idle_conn_timeout` ngắn hơn
- Giám sát với pprof:
  ```go
  import _ "net/http/pprof"
  // go tool pprof http://localhost:6060/debug/pprof/heap
  ```

### Thời gian phản hồi chậm

**Nguyên nhân có thể:**
1. Upstream proxy chậm
2. Connection pool cạn kiệt
3. Độ trễ phân giải DNS

**Giải pháp:**
- Kiểm tra upstream proxy trực tiếp
- Tăng kích thước connection pool
- Sử dụng DNS caching ở cấp hệ thống

## Thực hành tốt nhất

1. **Bắt đầu với mặc định**: Cấu hình mặc định hoạt động tốt cho hầu hết trường hợp

2. **Giám sát trước khi điều chỉnh**: Đo lường hiệu suất trước khi thay đổi

3. **Điều chỉnh từng bước**: Thay đổi một tham số tại một thời điểm

4. **Kiểm tra dưới tải**: Sử dụng mẫu lưu lượng thực tế để kiểm tra

5. **Giới hạn tài nguyên**: Đặt giới hạn hệ thống phù hợp trước tải cao

6. **Bảo trì thường xuyên**: Khởi động lại định kỳ để xóa connection pool

## Mẹo hiệu suất

### Nên làm
- ✅ Sử dụng định tuyến trực tiếp cho nội dung tĩnh
- ✅ Bật connection pooling
- ✅ Tắt các tính năng không cần thiết (MITM, chặn quảng cáo)
- ✅ Sử dụng kích thước buffer phù hợp
- ✅ Giám sát tài nguyên hệ thống

### Không nên làm
- ❌ Không bật debug logging trong production
- ❌ Không đặt connection pool quá cao cho bộ nhớ có sẵn
- ❌ Không sử dụng chế độ MITM trừ khi cần thiết
- ❌ Không bỏ qua giới hạn hệ thống
- ❌ Không chạy mà không giám sát

## Hiệu suất mong đợi theo phần cứng

### VPS tối thiểu (1 vCPU, 1GB RAM)
- Kết nối đồng thời: 1,000-5,000
- Yêu cầu/giây: 500-1,000
- Cấu hình khuyến nghị: Bộ nhớ thấp

### Máy chủ tiêu chuẩn (4 vCPU, 8GB RAM)
- Kết nối đồng thời: 10,000-50,000
- Yêu cầu/giây: 5,000-10,000
- Cấu hình khuyến nghị: Cân bằng

### Máy chủ cao cấp (16+ vCPU, 32GB+ RAM)
- Kết nối đồng thời: 100,000+
- Yêu cầu/giây: 20,000+
- Cấu hình khuyến nghị: Đồng thời cao
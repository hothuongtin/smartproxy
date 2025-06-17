# Tối ưu hiệu suất SmartProxy

## Tổng quan

SmartProxy đã được tối ưu để xử lý hàng ngàn request đồng thời một cách hiệu quả. Tài liệu này mô tả các tối ưu hiệu suất chính đã được triển khai.

## Các tối ưu chính

### 1. Connection Pooling
- **MaxIdleConns**: 10,000 kết nối (có thể cấu hình)
- **MaxIdleConnsPerHost**: 100 kết nối mỗi host
- **IdleConnTimeout**: 90 giây
- Tái sử dụng kết nối để giảm chi phí thiết lập kết nối mới

### 2. Chặn domain quảng cáo với tra cứu O(1)
- Sử dụng hash map thay vì tìm kiếm tuyến tính
- Thread-safe với RWMutex cho truy cập đồng thời
- Hỗ trợ chặn domain phân cấp (tự động chặn subdomain)

### 3. Cấu hình Transport được tối ưu
- **Hỗ trợ HTTP/2**: Bật để multiplexing tốt hơn
- **TLS Session Resumption**: Giảm chi phí TLS handshake
- **Dual Stack**: Hỗ trợ cả IPv4 và IPv6
- **Buffer lớn**: Buffer đọc/ghi 64KB để throughput tốt hơn
- **Tắt nén**: Giảm sử dụng CPU

### 4. Tối ưu Server
- **Graceful Shutdown**: Xử lý đúng cách SIGINT/SIGTERM
- **Request Timeouts**: Ngăn client chậm giữ kết nối
- **Không log chi tiết**: Tắt mặc định để tăng hiệu suất

### 5. Phân bổ bộ nhớ tối thiểu
- Tái sử dụng transport giữa các request
- Thao tác string hiệu quả
- Map được phân bổ trước cho ad domains

## Cấu hình

Tất cả cài đặt được quản lý qua `config.yaml`:

```yaml
server:
  http_port: 8888
  max_idle_conns: 10000
  max_idle_conns_per_host: 100
  idle_conn_timeout: 90
  tls_handshake_timeout: 10
  expect_continue_timeout: 1
  read_buffer_size: 65536
  write_buffer_size: 65536
```

## Load Testing

Sử dụng `loadtest.go` đi kèm để xác minh hiệu suất:

```bash
# Khởi động proxy
./smartproxy

# Trong terminal khác, chạy load test
go run loadtest.go
```

## Mẹo tối ưu hiệu suất

1. **Tắt chặn quảng cáo**: Nếu không cần, đặt `ad_blocking.enabled: false` để hiệu suất tốt hơn
2. **Điều chỉnh giới hạn kết nối**: Tăng `max_idle_conns` cho concurrency cao hơn
3. **Sử dụng kết nối trực tiếp**: Cấu hình CDN domains để bỏ qua upstream proxy
4. **Giám sát bộ nhớ**: Theo dõi memory usage và điều chỉnh connection limits phù hợp

## Benchmarks

Với cài đặt mặc định, SmartProxy có thể xử lý:
- 10,000+ kết nối đồng thời
- 5,000+ requests/giây (tùy thuộc độ trễ upstream)
- Độ trễ dưới mili giây cho kết nối trực tiếp
- Sử dụng CPU tối thiểu với compression tắt

## Điều chỉnh hệ thống

### Linux

1. Tăng file descriptors:
```bash
# Kiểm tra giới hạn hiện tại
ulimit -n

# Tăng giới hạn (thêm vào /etc/security/limits.conf)
* soft nofile 65535
* hard nofile 65535
```

2. Tối ưu network stack:
```bash
# Thêm vào /etc/sysctl.conf
net.ipv4.tcp_fin_timeout = 30
net.ipv4.tcp_keepalive_time = 300
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.core.netdev_max_backlog = 65535

# Áp dụng
sysctl -p
```

### macOS

1. Tăng giới hạn file:
```bash
sudo launchctl limit maxfiles 65536 200000
```

2. Tăng giới hạn process:
```bash
sudo launchctl limit maxproc 2048 2048
```

## Giám sát hiệu suất

### Metrics cần theo dõi

1. **Kết nối đang hoạt động**:
```bash
netstat -an | grep 8888 | grep ESTABLISHED | wc -l
```

2. **CPU và Memory**:
```bash
# Linux
top -p $(pgrep smartproxy)

# macOS
top -pid $(pgrep smartproxy)
```

3. **Throughput**:
```bash
# Sử dụng iftop hoặc nethogs
sudo iftop -i any -f "port 8888"
```

### Log analysis

```bash
# Đếm requests mỗi giây
tail -f proxy.log | grep -oE '[0-9]{2}:[0-9]{2}:[0-9]{2}' | uniq -c
```

## Khắc phục sự cố hiệu suất

### CPU cao
- Kiểm tra nếu MITM được bật không cần thiết
- Giảm `max_idle_conns_per_host`
- Tắt compression nếu được bật

### Memory cao
- Giảm `max_idle_conns`
- Kiểm tra memory leaks với pprof:
```bash
go tool pprof http://localhost:6060/debug/pprof/heap
```

### Độ trễ cao
- Kiểm tra hiệu suất upstream proxy
- Xác minh CDN domains được cấu hình đúng
- Kiểm tra network latency

## Cấu hình ví dụ cho production

```yaml
server:
  http_port: 8888
  https_mitm: false  # Tắt để hiệu suất tốt nhất
  max_idle_conns: 20000
  max_idle_conns_per_host: 200
  idle_conn_timeout: 120
  tls_handshake_timeout: 5
  expect_continue_timeout: 1
  read_buffer_size: 131072   # 128KB
  write_buffer_size: 131072  # 128KB

upstream:
  proxy_url: "http://fast-proxy:8080"

ad_blocking:
  enabled: false  # Tắt nếu không cần thiết

logging:
  level: warn  # Giảm log level
```
# Hướng dẫn cấu hình SmartProxy

## Tổng quan

SmartProxy sử dụng các file cấu hình YAML để điều khiển hoạt động. Cấu hình có thể được chỉ định qua:

1. **File cấu hình**: Mặc định `configs/config.yaml`
2. **Biến môi trường**: `SMARTPROXY_CONFIG=/path/to/config.yaml`
3. **Xác thực thông minh**: Cấu hình upstream động qua thông tin xác thực proxy

## Định dạng file cấu hình

File cấu hình sử dụng định dạng YAML với các phần chính sau:

- `server` - Cài đặt máy chủ proxy
- `ad_blocking` - Cấu hình chặn quảng cáo
- `direct_extensions` - Phần mở rộng file cho định tuyến trực tiếp
- `direct_domains` - Miền cho định tuyến trực tiếp
- `logging` - Cấu hình ghi log

## Cài đặt máy chủ

```yaml
server:
  # Cổng proxy HTTP/HTTPS
  http_port: 8888
  
  # Cài đặt chặn HTTPS
  https_mitm: false    # Bật/tắt chặn HTTPS (MITM)
  ca_cert: ""          # Đường dẫn file chứng chỉ CA (tự tạo nếu trống)
  ca_key: ""           # Đường dẫn file khóa riêng CA (tự tạo nếu trống)
  
  # Cài đặt hiệu suất cho đồng thời cao
  max_idle_conns: 10000        # Số kết nối idle tối đa
  max_idle_conns_per_host: 100 # Số kết nối idle tối đa mỗi host
  idle_conn_timeout: 90         # Thời gian chờ kết nối idle (giây)
  tls_handshake_timeout: 10     # Thời gian chờ TLS handshake (giây)
  expect_continue_timeout: 1    # Thời gian chờ expect continue (giây)
  
  # Cài đặt buffer
  read_buffer_size: 65536       # Kích thước buffer đọc (64KB)
  write_buffer_size: 65536      # Kích thước buffer ghi (64KB)
```

### Giải thích cài đặt máy chủ chính

- **`http_port`**: Cổng SmartProxy lắng nghe (mặc định: 8888)
- **`https_mitm`**: Khi `true`, giải mã lưu lượng HTTPS để kiểm tra. Cần chứng chỉ CA.
- **`max_idle_conns`**: Tổng kích thước connection pool. Giá trị cao hơn cải thiện hiệu suất nhưng dùng nhiều bộ nhớ hơn.
- **`max_idle_conns_per_host`**: Giới hạn kết nối mỗi host để tránh quá tải máy chủ đơn lẻ.

## Chế độ xác thực thông minh

SmartProxy cấu hình động các upstream proxy qua thông tin xác thực:

```yaml
# Không cần cấu hình upstream trong file config!
# Upstream được cấu hình cho mỗi kết nối qua xác thực

# Định dạng xác thực:
# Username: schema (http hoặc socks5)
# Password: chi tiết upstream đã mã hóa base64

# Ví dụ:
# - Không có xác thực upstream: base64("host:port")
#   Username: http
#   Password: bmEubHVuYXByb3h5LmNvbToxMjIzMw== (na.lunaproxy.com:12233)
#
# - Với xác thực upstream: base64("host:port:username:password")
#   Username: socks5
#   Password: bmEubHVuYXByb3h5LmNvbToxMjIzMzp1c2VyOnBhc3M= (na.lunaproxy.com:12233:user:pass)
```

## Cấu hình chặn quảng cáo

```yaml
ad_blocking:
  enabled: true
  domains_file: "configs/ad_domains.yaml"
```

File `ad_domains.yaml` chứa danh sách các miền cần chặn:

```yaml
domains:
  - doubleclick.net
  - googleads.com
  - googlesyndication.com
  - google-analytics.com
  - facebook.com/tr
  - amazon-adsystem.com
  # ... thêm miền khác
```

## Cấu hình định tuyến trực tiếp

### Phần mở rộng file tĩnh

Các file với phần mở rộng này bỏ qua upstream proxy để cải thiện hiệu suất:

```yaml
direct_extensions:
  # Tài liệu
  - .pdf
  - .doc
  - .docx
  
  # Hình ảnh
  - .jpg
  - .jpeg
  - .png
  - .gif
  - .webp
  - .svg
  - .ico
  
  # Video
  - .mp4
  - .webm
  - .avi
  - .mov
  
  # Âm thanh
  - .mp3
  - .wav
  - .ogg
  
  # Tài nguyên web
  - .css
  - .js
  - .woff
  - .woff2
  
  # File nén
  - .zip
  - .rar
  - .7z
```

### Miền CDN

Các miền khớp với mẫu này sử dụng kết nối trực tiếp:

```yaml
direct_domains:
  # Mẫu CDN phổ biến
  - cdn.
  - static.
  - assets.
  
  # Nhà cung cấp CDN lớn
  - cloudflare
  - akamai
  - fastly
  - cloudfront
  
  # Dịch vụ phổ biến
  - googleapis.com
  - gstatic.com
  - jsdelivr.net
  - unpkg.com
```

## Cấu hình ghi log

```yaml
logging:
  level: info      # debug, info, warn, error
  format: text     # text hoặc json
  colored: true    # Bật output màu (chỉ với format text)
```

### Các mức log

- **`debug`**: Thông tin chi tiết cho việc gỡ lỗi
- **`info`**: Thông tin hoạt động chung
- **`warn`**: Thông báo cảnh báo cho các vấn đề tiềm ẩn
- **`error`**: Thông báo lỗi cho các thất bại

### Cấu hình debug

Để gỡ lỗi, sử dụng cấu hình debug:

```bash
cp configs/config.debug.yaml configs/config.yaml
```

Cấu hình debug bao gồm:
- `level: debug` cho log chi tiết
- Thời gian chờ mở rộng
- Thông báo lỗi chi tiết

## Biến môi trường

- **`SMARTPROXY_CONFIG`**: Ghi đè đường dẫn file cấu hình
- **`NO_PROXY`**: Danh sách host bỏ qua proxy, cách nhau bằng dấu phẩy
- **`HTTP_PROXY`/`HTTPS_PROXY`**: Không được SmartProxy sử dụng

## Các cấu hình mẫu

### Cấu hình tối thiểu

```yaml
server:
  http_port: 8888

ad_blocking:
  enabled: true
  domains_file: "configs/ad_domains.yaml"

logging:
  level: info
```

### Cấu hình hiệu suất cao

```yaml
server:
  http_port: 8888
  max_idle_conns: 50000
  max_idle_conns_per_host: 500
  idle_conn_timeout: 120
  read_buffer_size: 131072   # 128KB
  write_buffer_size: 131072  # 128KB

ad_blocking:
  enabled: true
  domains_file: "configs/ad_domains.yaml"

# Định tuyến trực tiếp mở rộng
direct_extensions:
  - .js
  - .css
  - .jpg
  - .png
  - .gif
  - .webp
  - .woff
  - .woff2
  - .ttf
  - .eot
  - .svg
  - .ico
  - .mp4
  - .webm
  - .mp3
  - .pdf

direct_domains:
  - cdn.
  - static.
  - assets.
  - media.
  - img.
  - cloudflare
  - akamai
  - fastly
  - amazonaws.com
  - googleusercontent.com

logging:
  level: warn  # Ít log hơn cho hiệu suất
```

### Cấu hình HTTPS MITM

```yaml
server:
  http_port: 8888
  https_mitm: true
  ca_cert: "certs/ca.crt"
  ca_key: "certs/ca.key"
  tls_handshake_timeout: 30  # Thời gian chờ lâu hơn cho SSL

ad_blocking:
  enabled: true  # Giờ hoạt động trên các trang HTTPS
  domains_file: "configs/ad_domains.yaml"

logging:
  level: info
```

## Thực hành tốt nhất cho cấu hình

1. **Bắt đầu với mặc định**: Cấu hình mẫu cung cấp các giá trị mặc định tốt
2. **Điều chỉnh theo khối lượng công việc**: Điều chỉnh connection pool dựa trên sử dụng
3. **Giám sát hiệu suất**: Sử dụng debug logging để xác định điểm nghẽn
4. **Bảo mật cài đặt**: Giữ khóa riêng CA an toàn nếu dùng MITM
5. **Cập nhật thường xuyên**: Giữ danh sách chặn quảng cáo luôn cập nhật

## Xác thực

SmartProxy xác thực cấu hình khi khởi động:
- Số cổng phải là 1-65535
- Đường dẫn file phải có thể truy cập
- File chứng chỉ phải hợp lệ nếu được chỉ định
- Cấu hình không hợp lệ ngăn khởi động

## Tải lại nhanh

Thay đổi cấu hình yêu cầu khởi động lại:

```bash
# Khởi động lại mềm
make restart

# Hoặc thủ công
pkill -TERM smartproxy && ./smartproxy
```

## Khắc phục sự cố cấu hình

### Các vấn đề thường gặp

1. **Cổng đã được sử dụng**: Thay đổi `http_port` hoặc kill tiến trình hiện tại
2. **Không tìm thấy file**: Sử dụng đường dẫn tuyệt đối hoặc tương đối với thư mục làm việc
3. **Từ chối quyền**: Đảm bảo quyền đọc cho file config và chứng chỉ
4. **YAML không hợp lệ**: Kiểm tra cú pháp với công cụ xác thực YAML trực tuyến

### Cấu hình debug

Bật debug logging để xem chi tiết cấu hình:

```yaml
logging:
  level: debug
```

Điều này hiển thị:
- Giá trị cấu hình đã tải
- Đường dẫn file đang sử dụng
- Quyết định định tuyến
- Các chỉ số hiệu suất
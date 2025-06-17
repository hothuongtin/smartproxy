# HTTPS Upstream Routing trong SmartProxy

## Tổng Quan

SmartProxy hiện tại định tuyến đúng cách các kết nối HTTPS thông qua upstream proxy đã cấu hình khi sử dụng chế độ smart authentication. Tài liệu này giải thích cách xử lý kết nối HTTPS.

## Cách Hoạt Động

### 1. HTTPS Request từ Client
Khi client thực hiện HTTPS request:
```
Client → CONNECT example.com:443 HTTP/1.1
         Proxy-Authorization: Basic [credentials]
```

### 2. Phân Tích Authentication
SmartProxy trích xuất cấu hình upstream proxy từ authentication:
- **Username**: Giao thức (`http` hoặc `socks5`)
- **Password**: Chi tiết upstream proxy được mã hóa Base64

### 3. Định Tuyến Kết Nối

#### Đối với Domains Thông Thường:
```
Client → SmartProxy → Upstream Proxy → Target Server
```
Kết nối HTTPS được tunnel thông qua upstream proxy đã cấu hình.

#### Đối với CDN Domains:
```
Client → SmartProxy → Target Server (Direct)
```
CDN domains bỏ qua upstream proxy để có hiệu suất tốt hơn.

### 4. Chi Tiết Implementation

#### HTTP Upstream Proxy
Đối với HTTP upstream proxies, SmartProxy:
1. Kết nối tới upstream proxy
2. Gửi CONNECT request để thiết lập tunnel
3. Bao gồm proxy authentication nếu được cấu hình
4. Chuyển tiếp traffic đã mã hóa

#### SOCKS5 Upstream Proxy
Đối với SOCKS5 upstream proxies, SmartProxy:
1. Sử dụng Go SOCKS5 client library
2. Thiết lập kết nối SOCKS5 với authentication
3. Tunnel HTTPS traffic thông qua SOCKS5 proxy

## Ví Dụ Cấu Hình

### HTTP Proxy không có Authentication
```bash
# Upstream: proxy.example.com:8080
username="http"
password=$(echo -n "proxy.example.com:8080" | base64)
curl -x "http://${username}:${password}@localhost:8888" https://example.com
```

### HTTP Proxy có Authentication
```bash
# Upstream: proxy.example.com:8080 với user:pass
username="http"
password=$(echo -n "proxy.example.com:8080:user:pass" | base64)
curl -x "http://${username}:${password}@localhost:8888" https://example.com
```

### SOCKS5 Proxy
```bash
# Upstream: socks5.example.com:1080
username="socks5"
password=$(echo -n "socks5.example.com:1080" | base64)
curl -x "http://${username}:${password}@localhost:8888" https://example.com
```

## Testing

Sử dụng script test được cung cấp để xác minh HTTPS upstream routing:
```bash
./scripts/test/test_https_upstream.sh
```

## Cân Nhắc Hiệu Suất

1. **Tối ưu CDN**: CDN domains (ví dụ: cdn.jsdelivr.net, cdnjs.cloudflare.com) tự động sử dụng kết nối trực tiếp
2. **Connection Pooling**: Upstream connections được pool và tái sử dụng để hiệu quả
3. **Timeout Handling**: Timeout mặc định 30 giây cho upstream connections

## Khắc Phục Sự Cố

### HTTPS Connection Thất Bại
- Xác minh upstream proxy có thể truy cập được
- Kiểm tra upstream proxy hỗ trợ HTTPS/CONNECT method
- Đảm bảo thông tin đăng nhập chính xác

### Connection Timeout
- Kiểm tra kết nối mạng tới upstream proxy
- Xác minh upstream proxy không chặn IP của bạn
- Thử tăng timeout trong configs/config.yaml

### CDN Vẫn Sử Dụng Upstream
- Kiểm tra domain có trong danh sách CDN trong configs/config.yaml không
- Thêm custom CDN domains vào danh sách direct_domains

## Ghi Chú Bảo Mật

1. **End-to-End Encryption**: HTTPS traffic vẫn được mã hóa thông qua toàn bộ chuỗi
2. **Không MITM**: Với `https_mitm: false` (mặc định), SmartProxy không thể xem nội dung HTTPS
3. **Bảo Vệ Credential**: Sử dụng HTTPS tới SmartProxy nếu chạy trên mạng không tin cậy
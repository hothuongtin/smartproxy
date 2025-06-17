# Tính năng SmartProxy

## Tổng quan

SmartProxy được thiết kế là một proxy HTTP/HTTPS thông minh tối ưu hóa lưu lượng web thông qua quyết định định tuyến thông minh, chặn quảng cáo và xử lý kết nối hiệu suất cao.

## Tính năng cốt lõi

### 1. Logic định tuyến thông minh

SmartProxy tự động xác định đường đi tốt nhất cho mỗi yêu cầu:

#### Phát hiện file tĩnh
- **Là gì**: Tự động phát hiện file tĩnh bằng phần mở rộng
- **Cách thức**: Tìm kiếm O(1) của phần mở rộng file từ URL path
- **File**: `.js`, `.css`, `.jpg`, `.png`, `.gif`, `.pdf`, v.v.
- **Lợi ích**: Kết nối trực tiếp cho tải nhanh hơn
- **Ví dụ**:
  ```
  http://example.com/app.js?v=123 → Kết nối trực tiếp
  http://example.com/api/data → Upstream proxy
  ```

#### Phát hiện miền CDN
- **Là gì**: Xác định miền CDN cho định tuyến trực tiếp
- **Cách thức**: So khớp mẫu trên tên miền
- **Mẫu**: `cdn.`, `static.`, `assets.`, các nhà cung cấp CDN lớn
- **Lợi ích**: Bỏ qua proxy cho nội dung đã được tối ưu
- **Ví dụ**:
  ```
  https://cdn.example.com → Kết nối trực tiếp
  https://api.example.com → Upstream proxy
  ```

#### Luồng quyết định định tuyến thông minh
```
Yêu cầu → Có phải quảng cáo? → Chặn (204 No Content)
         ↓ Không
         Là file tĩnh? → Kết nối trực tiếp
         ↓ Không
         Là miền CDN? → Kết nối trực tiếp
         ↓ Không
         Định tuyến qua upstream proxy
```

### 2. Chặn quảng cáo

#### Chặn hiệu suất cao
- **Tìm kiếm O(1)**: Hash map cho khớp miền tức thì
- **Chặn phân cấp**: Tự động chặn subdomain
- **Thread-Safe**: Truy cập đồng thời với RWMutex
- **Ví dụ**:
  ```yaml
  # Chặn "doubleclick.net" cũng chặn:
  - ads.doubleclick.net
  - static.doubleclick.net
  - any.subdomain.doubleclick.net
  ```

#### Chi tiết triển khai
- Tải miền từ `ad_domains.yaml`
- Trả về 204 No Content cho miền bị chặn
- Hoạt động trên HTTP và HTTPS (với MITM bật)
- Không cấp phát bộ nhớ cho tìm kiếm

### 3. Connection Pooling

#### Lợi ích hiệu suất
- **Tái sử dụng kết nối**: Giảm tổn hạo TCP handshake
- **Giới hạn có thể cấu hình**: Điều chỉnh theo khối lượng công việc
- **Giới hạn mỗi host**: Ngăn quá tải máy chủ đơn lẻ

#### Cấu hình mặc định
```yaml
max_idle_conns: 10000        # Tổng kích thước pool
max_idle_conns_per_host: 100 # Giới hạn mỗi host
idle_conn_timeout: 90         # Giây trước khi đóng
```

#### Loại kết nối
1. **Direct Transport**: Cho file tĩnh và CDN
2. **HTTP Proxy Transport**: Cho upstream proxy HTTP
3. **SOCKS5 Transport**: Cho upstream proxy SOCKS5

### 4. Xử lý HTTPS

#### Chế độ Tunneling (Mặc định)
- **Cách hoạt động**: Tạo tunnel mã hóa mà không kiểm tra
- **Lợi ích**:
  - Không có cảnh báo chứng chỉ
  - Mã hóa end-to-end thực sự
  - Không cần cấu hình
- **Hạn chế**:
  - Không thể chặn quảng cáo trên HTTPS
  - Không thể phát hiện file tĩnh trên HTTPS

#### Chế độ MITM (Nâng cao)
- **Cách hoạt động**: Giải mã, kiểm tra và mã hóa lại
- **Lợi ích**:
  - Chặn quảng cáo trên các trang HTTPS
  - Phát hiện file tĩnh cho HTTPS
  - Trí tuệ định tuyến đầy đủ
- **Yêu cầu**:
  - Cài đặt chứng chỉ CA
  - Xác thực cho tất cả yêu cầu

### 5. Xác thực thông minh

#### Cấu hình upstream động
Thay vì cấu hình upstream tĩnh, SmartProxy sử dụng thông tin xác thực để cấu hình động upstream proxy cho mỗi kết nối.

**Định dạng**:
```
Username: <schema>  # http hoặc socks5
Password: <base64-encoded-upstream>
```

**Ví dụ**:
```bash
# HTTP proxy không có xác thực
Username: http
Password: cHJveHkuZXhhbXBsZS5jb206ODA4MA== # proxy.example.com:8080

# SOCKS5 với xác thực
Username: socks5  
Password: c29ja3MuZXhhbXBsZS5jb206MTA4MDp1c2VyOnBhc3M= # socks.example.com:1080:user:pass
```

#### Lợi ích
- Nhiều upstream proxy mà không cần khởi động lại
- Cấu hình proxy cho mỗi client
- Xoay vòng proxy dễ dàng
- Không cần thay đổi file config

### 6. Hỗ trợ HTTP/2

- **Ghép kênh**: Nhiều yêu cầu qua một kết nối
- **Nén header**: Giảm tổn hạo
- **Server Push**: Không sử dụng nhưng được hỗ trợ
- **Tự động**: Thương lượng qua ALPN

### 7. Tối ưu hóa hiệu suất

#### Hoạt động Zero-Copy
- `io.Copy` hiệu quả cho truyền dữ liệu
- Cấp phát bộ nhớ tối thiểu
- Tái sử dụng buffer khi có thể

#### Buffer tối ưu
```yaml
read_buffer_size: 65536   # Mặc định 64KB
write_buffer_size: 65536  # Mặc định 64KB
```

#### Xử lý yêu cầu đồng thời
- Go routine cho mỗi yêu cầu
- I/O không chặn
- Quản lý kết nối mềm

### 8. Ghi log toàn diện

#### Ghi log có cấu trúc
- Sử dụng `slog` cho output có cấu trúc
- Output màu cho dễ đọc
- Tùy chọn định dạng JSON

#### Các mức log
- **Debug**: Quyết định định tuyến chi tiết
- **Info**: Tóm tắt yêu cầu/phản hồi  
- **Warn**: Vấn đề tiềm ẩn
- **Error**: Lỗi và thất bại

#### Ví dụ output
```
12:34:56.789 INFO Incoming request method=GET url=http://example.com client=127.0.0.1:54321
12:34:56.790 DEBUG URL identified as static file extension=.js action=direct_connection
12:34:56.891 INFO Request completed status=200 duration=101ms
```

### 9. Tắt mềm

- Xử lý tín hiệu SIGINT/SIGTERM
- Chờ các kết nối hoạt động
- Đóng transport đúng cách
- Ngăn mất dữ liệu

### 10. Hỗ trợ đa nền tảng

#### Nền tảng được hỗ trợ
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)
- FreeBSD (amd64)

#### Hỗ trợ Docker
- Image Distroless (~15MB)
- Image Scratch (~12MB)
- Build đa kiến trúc

## So sánh tính năng

| Tính năng | SmartProxy | Proxy truyền thống |
|-----------|------------|-------------------|
| Phát hiện file tĩnh | ✅ Tự động | ❌ Tất cả qua proxy |
| Định tuyến CDN | ✅ Tự động | ❌ Tất cả qua proxy |
| Chặn quảng cáo | ✅ Hiệu suất O(1) | ⚠️ Tùy thuộc |
| Connection pooling | ✅ 10,000+ kết nối | ⚠️ Hạn chế |
| Upstream động | ✅ Mỗi kết nối | ❌ Config tĩnh |
| Tùy chọn HTTPS | ✅ Tunnel hoặc MITM | ⚠️ Thường một |
| Sử dụng bộ nhớ | ✅ ~50MB khi tải | ⚠️ Thường cao hơn |

## Các trường hợp sử dụng

### 1. Môi trường phát triển
- Định tuyến API call qua proxy
- Kết nối trực tiếp cho asset
- Debug với chế độ MITM

### 2. Duyệt web không quảng cáo
- Chặn quảng cáo ở cấp proxy
- Hoạt động cho tất cả thiết bị
- Không cần tiện ích trình duyệt

### 3. Cài đặt đa proxy
- Proxy khác nhau cho mỗi client
- Xoay vòng proxy dễ dàng
- Không thay đổi cấu hình

### 4. Tối ưu hóa hiệu suất
- Kết nối CDN trực tiếp
- Kết nối transport được cache
- Giảm độ trễ

## Tính năng tương lai

Các cải tiến dự kiến:
- Sửa đổi yêu cầu/phản hồi
- Quy tắc định tuyến tùy chỉnh
- Chỉ số và giám sát
- Hỗ trợ WebSocket
- Hỗ trợ HTTP/3
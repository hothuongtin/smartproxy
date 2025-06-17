# SmartProxy Debug Logging

SmartProxy bao gồm hệ thống debug logging toàn diện để giúp khắc phục sự cố và hiểu rõ hành vi của proxy.

## Bắt Đầu Nhanh

### Kích Hoạt Chế Độ Debug

1. **Sử dụng Makefile (Khuyến nghị)**
```bash
make debug
```

2. **Sử dụng file config**
Chỉnh sửa `configs/config.yaml`:
```yaml
logging:
  level: debug
```

3. **Sử dụng debug config có sẵn**
```bash
cp configs/config.debug.yaml configs/config.yaml
./build/smartproxy
```

## Những Gì Debug Mode Ghi Log

### Xử Lý Request
- **Xác thực**: Phân tích thông tin đăng nhập, xác thực schema, giải mã base64
- **Định tuyến**: Logic quyết định kết nối trực tiếp vs upstream
- **Thời gian**: Thời lượng request/response

### Ví Dụ Debug Output

```
15:04:05.123 INFO Starting high-performance proxy server address=:8888 mode=smart_proxy_auth
15:04:05.124 DEBUG Configuration details http_port=8888 https_mitm=false max_idle_conns=10000
15:04:05.125 DEBUG Sample ad domains samples=[doubleclick.net, googleads.com] total=1000

# Request đến
15:04:10.234 DEBUG Incoming request method=GET url=http://example.com host=example.com
15:04:10.235 DEBUG Parsing upstream from authentication username=http password_length=32
15:04:10.236 DEBUG Decoded upstream configuration decoded=proxy.example.com:8080 schema=http
15:04:10.237 DEBUG Authentication successful upstream_type=http upstream_host=proxy.example.com

# Quyết định định tuyến
15:04:10.238 DEBUG Routing decision for request url=http://example.com/script.js
15:04:10.239 DEBUG URL identified as static file extension=.js action=direct_connection
15:04:10.240 DEBUG Using direct connection reason=static_file_or_cdn

# Response
15:04:10.345 DEBUG Direct request completed status=200 duration=105ms
15:04:10.346 DEBUG Response received status=200 content_type=application/javascript
```

### Thông Tin Được Ghi Log

#### 1. **Khởi Động Server**
- Chi tiết cấu hình
- Cài đặt hiệu suất
- Extensions và domains được load
- Trạng thái chặn quảng cáo

#### 2. **Xác Thực**
- Phân tích Basic auth
- Xác thực schema (http/socks5)
- Giải mã Base64
- Phân tích upstream

#### 3. **Logic Định Tuyến**
- Phát hiện file tĩnh
- Khớp domain CDN
- Chặn domain quảng cáo
- Lựa chọn upstream

#### 4. **Xử Lý Kết Nối**
- Tạo transport
- Cache kết nối
- Thiết lập SOCKS5/HTTP proxy
- Chi tiết lỗi

#### 5. **Metrics Hiệu Suất**
- Thời gian request/response
- Cache hits/misses
- Tái sử dụng kết nối

## Danh Mục Debug Logging

### Phát Hiện File Tĩnh
```
DEBUG URL identified as static file url=http://cdn.com/app.js extension=.js
DEBUG URL not a static file url=http://api.com/data path=/data
# Với MITM được bật, hoạt động cho HTTPS:
DEBUG URL identified as static file url=https://cdn.com/app.js extension=.js
```

### Phát Hiện CDN
```
DEBUG Domain identified as CDN host=cdn.example.com pattern=cdn.
DEBUG Domain not a CDN host=api.example.com
```

### Chặn Quảng Cáo
```
DEBUG Domain blocked (exact match) host=doubleclick.net action=blocked
DEBUG Domain blocked (parent match) host=ads.google.com blocked_parent=google.com
```

### Upstream Proxy
```
DEBUG Creating new transport type=http host=proxy.com port=8080
DEBUG Using cached transport cache_key=http:proxy.com:8080
DEBUG Upstream request completed status=200 duration=250ms
```

### HTTPS với MITM
```
DEBUG MITM authentication check passed upstream=http://proxy.com:8080
DEBUG Intercepting HTTPS request method=GET url=https://example.com/script.js
DEBUG URL identified as static file extension=.js action=direct_connection
DEBUG Using direct connection for HTTPS static file url=https://cdn.com/app.js
DEBUG Using upstream proxy for HTTPS request url=https://api.com/data
```

## Tác Động Hiệu Suất

Debug logging có tác động hiệu suất tối thiểu:
- Structured logging với slog
- Kiểm tra debug có điều kiện
- Định dạng chuỗi hiệu quả
- Không có code debug trong hot paths

## Khắc Phục Sự Cố

### Vấn Đề Thường Gặp

1. **Không có debug output**
   - Xác minh `logging.level: debug` trong config
   - Kiểm tra đường dẫn config file
   - Đảm bảo logger initialization

2. **Quá nhiều output**
   - Lọc theo component
   - Sử dụng grep cho các pattern cụ thể
   - Điều chỉnh log level thành info/warn

3. **Thiếu thông tin**
   - Kiểm tra feature có debug logging không
   - Gửi issue để bổ sung logging

### Lọc Debug Output

```bash
# Chỉ authentication logs
./build/smartproxy 2>&1 | grep "auth"

# Chỉ routing decisions
./build/smartproxy 2>&1 | grep -E "routing|direct|upstream"

# Chỉ errors
./build/smartproxy 2>&1 | grep -E "ERROR|WARN"
```

## Cân Nhắc Bảo Mật

Debug mode có thể ghi log thông tin nhạy cảm:
- Thông tin đăng nhập proxy đã giải mã
- URL request
- Authentication headers

**Không bao giờ sử dụng debug mode trong production!**

## Đóng Góp

Để thêm debug logging:

1. Sử dụng global logger
```go
logger.Debug("Mô tả hoạt động",
    "key1", value1,
    "key2", value2)
```

2. Kiểm tra debug level cho các operation tốn kém
```go
if logger.Enabled(nil, slog.LevelDebug) {
    // Logic debug tốn kém
}
```

3. Tuân theo naming conventions
- Sử dụng message mô tả rõ ràng
- Bao gồm context liên quan
- Tránh ghi log dữ liệu nhạy cảm
# Hướng dẫn khắc phục sự cố SmartProxy

## Câu hỏi thường gặp

### Xác thực và hành vi trình duyệt

#### H: Tại sao curl hiển thị lỗi xác thực nhưng Chrome vẫn hoạt động?

**Đ:** Chrome sử dụng kết nối bền vững (HTTP keep-alive). Khi bạn nhập thông tin sai:
- curl tạo kết nối mới → kiểm tra xác thực → hiển thị lỗi
- Chrome tái sử dụng kết nối đã xác thực → tiếp tục hoạt động

Đây là hành vi chuẩn của HTTP, không phải lỗi.

**Giải pháp:**
1. Đóng Chrome hoàn toàn để ngắt kết nối cũ
2. Hoặc vào `chrome://net-internals/#sockets` → "Flush socket pools"
3. Mở lại Chrome và nhập thông tin mới

#### H: Làm thế nào để kiểm tra xác thực đúng cách?

**Với curl:**
```bash
# Kiểm tra với thông tin chính xác
correct_password=$(echo -n "proxy.example.com:8080" | base64)
curl -x http://http:${correct_password}@localhost:8888 http://ipinfo.io

# Buộc tạo kết nối mới
curl -x http://http:wrongpassword@localhost:8888 -H "Connection: close" http://httpbin.org/ip
```

**Với trình duyệt:**
1. Đóng tất cả cửa sổ trình duyệt
2. Xóa dữ liệu duyệt (Ctrl+Shift+Delete)
3. Mở với profile mới:
   ```bash
   # macOS
   open -na "Google Chrome" --args --user-data-dir="/tmp/chrome_test" --proxy-server="http://localhost:8888"
   
   # Windows
   chrome.exe --user-data-dir="%TEMP%\chrome_test" --proxy-server="http://localhost:8888"
   ```

### Các lỗi thường gặp và giải pháp

#### "illegal base64 data at input byte X"

**Nguyên nhân:** Vấn đề mã hóa base64
- Bị ngắt dòng với newline (thường tại 76 ký tự)
- Có khoảng trắng thừa
- Có ký tự không hợp lệ

**Giải pháp:** SmartProxy tự động xử lý base64 bị ngắt. Tạo base64 đúng:
```bash
# Sử dụng -n để tránh newline cuối
echo -n "host:port:user:pass" | base64

# Xác minh mã hóa của bạn
echo "YOUR_BASE64_STRING" | base64 -d
```

#### "LibreSSL: error:1404B42E:SSL routines:ST_CONNECT:tlsv1 alert protocol version"

**Nguyên nhân:** Sử dụng `https://` trong URL proxy

**Giải pháp:** Luôn sử dụng `http://` cho URL proxy:
```bash
# ✅ ĐÚNG - Luôn dùng http:// cho URL proxy
curl -x http://http:PASSWORD@localhost:8888 https://example.com

# ❌ SAI - Không dùng https:// cho URL proxy
curl -x https://http:PASSWORD@localhost:8888 https://example.com
```

#### "Yêu cầu xác thực" (407 Proxy Authentication Required)

**Nguyên nhân:** Không cung cấp thông tin xác thực

**Giải pháp:** Cung cấp username và password:
```bash
# Dòng lệnh
curl -x http://http:BASE64_PASSWORD@localhost:8888 http://example.com

# Biến môi trường
export http_proxy=http://http:BASE64_PASSWORD@localhost:8888
export https_proxy=http://http:BASE64_PASSWORD@localhost:8888
```

#### "Thông tin xác thực không hợp lệ" (403 Forbidden)

**Nguyên nhân:**
- Username sai (phải là `http` hoặc `socks5`)
- Mã hóa base64 không hợp lệ
- Chi tiết upstream proxy không chính xác

**Giải pháp:** Xác minh thông tin của bạn:
```bash
# Kiểm tra username đúng
# Phải là "http" hoặc "socks5"

# Xác minh mã hóa base64
echo -n "proxy.example.com:8080" | base64
# Sau đó giải mã để xác minh
echo "cHJveHkuZXhhbXBsZS5jb206ODA4MA==" | base64 -d
```

#### Cổng đã được sử dụng

**Lỗi:** `bind: address already in use`

**Giải pháp:**
```bash
# Tìm process dùng cổng
lsof -i :8888  # macOS/Linux
netstat -ano | findstr :8888  # Windows

# Kill proxy hiện có
make kill
# Hoặc
pkill -f smartproxy

# Khởi động lại
make run
```

### Vấn đề HTTPS và chứng chỉ

#### H: Tại sao ip-api.com trả về lỗi "SSL unavailable"?

**Đ:** API miễn phí của ip-api.com không hỗ trợ HTTPS. Sử dụng HTTP thay vì:
- ❌ Sai: `https://ip-api.com/json`
- ✅ Đúng: `http://ip-api.com/json`

Đây không phải lỗi proxy - đây là giới hạn của gói miễn phí.

#### Cảnh báo chứng chỉ trong chế độ MITM

**Nguyên nhân:** Chứng chỉ CA chưa được cài đặt hoặc tin cậy đúng cách

**Giải pháp:**
1. Tạo chứng chỉ CA:
   ```bash
   make ca-cert
   ```

2. Cài đặt chứng chỉ CA trên client:
   - **macOS:** Double-click `certs/ca.crt` → Thêm vào System keychain → Trust cho SSL
   - **Windows:** Double-click `certs/ca.crt` → Cài vào "Trusted Root Certification Authorities"
   - **Linux:** `sudo cp certs/ca.crt /usr/local/share/ca-certificates/smartproxy.crt && sudo update-ca-certificates`

3. Khởi động lại trình duyệt sau khi cài đặt

#### H: Làm thế nào truy cập các trang HTTPS khi MITM bị tắt?

**Đ:** Khi `https_mitm: false` (mặc định), SmartProxy tunnel kết nối HTTPS mà không giải mã. Điều này hoạt động hoàn hảo cho tất cả các trang HTTPS:
- Thiết lập tunnel sử dụng phương thức CONNECT
- Định tuyến qua upstream proxy đã cấu hình
- Duy trì mã hóa end-to-end
- Không có cảnh báo chứng chỉ

### Vấn đề hiệu suất

#### Thời gian phản hồi chậm

**Nguyên nhân và giải pháp có thể:**

1. **Upstream proxy chậm**
   - Kiểm tra upstream proxy trực tiếp
   - Thử upstream proxy khác
   - Kiểm tra độ trễ mạng

2. **Connection pool cạn kiệt**
   ```yaml
   server:
     max_idle_conns: 50000  # Tăng cho tải cao
     max_idle_conns_per_host: 500
   ```

3. **Quá nhiều logging**
   ```yaml
   logging:
     level: warn  # Giảm từ debug/info
   ```

4. **Giới hạn hệ thống**
   ```bash
   # Kiểm tra giới hạn file descriptor
   ulimit -n
   
   # Tăng giới hạn
   ulimit -n 65536
   ```

#### Sử dụng bộ nhớ cao

**Giải pháp:**
1. Giảm kích thước connection pool
2. Bật memory profiling:
   ```go
   import _ "net/http/pprof"
   // Truy cập http://localhost:6060/debug/pprof/heap
   ```
3. Kiểm tra rò rỉ kết nối trong log

### Debug Logging

#### Bật chế độ debug

**Phương pháp 1: File cấu hình**
```yaml
logging:
  level: debug
```

**Phương pháp 2: Sử dụng config debug**
```bash
cp configs/config.debug.yaml configs/config.yaml
./smartproxy
```

**Phương pháp 3: Makefile**
```bash
make debug
```

#### Chế độ debug hiển thị gì

- Chi tiết xác thực
- Quyết định định tuyến
- Xử lý kết nối
- Chỉ số hiệu suất
- Chi tiết lỗi

#### Ví dụ output debug
```
15:04:10.234 DEBUG Incoming request method=GET url=http://example.com
15:04:10.235 DEBUG Parsing upstream from authentication username=http
15:04:10.236 DEBUG Authentication successful upstream_type=http
15:04:10.237 DEBUG URL identified as static file extension=.js
15:04:10.238 DEBUG Using direct connection reason=static_file
```

#### Lọc output debug

```bash
# Chỉ log xác thực
./smartproxy 2>&1 | grep "auth"

# Chỉ quyết định định tuyến
./smartproxy 2>&1 | grep -E "routing|direct|upstream"

# Chỉ lỗi
./smartproxy 2>&1 | grep -E "ERROR|WARN"
```

### Vấn đề trường hợp sử dụng cụ thể

#### H: Tại sao một số yêu cầu HTTPS nhanh hơn các yêu cầu khác?

SmartProxy tự động sử dụng kết nối trực tiếp cho:
- File tĩnh (hình ảnh, CSS, JS, v.v.)
- Miền CDN đã biết
- Nội dung không cần lọc

Điều này bỏ qua upstream proxy để hiệu suất tốt hơn.

#### H: Làm thế nào để tắt chặn quảng cáo?

Chỉnh sửa `configs/config.yaml`:
```yaml
ad_blocking:
  enabled: false
```

#### H: Tôi có thể sử dụng SOCKS5 proxy làm upstream?

Có! Sử dụng xác thực thông minh:
```bash
# SOCKS5 không có xác thực
echo -n "socks5.example.com:1080" | base64
# Sử dụng với username: socks5

# SOCKS5 với xác thực
echo -n "socks5.example.com:1080:user:pass" | base64
# Sử dụng với username: socks5
```

#### H: Làm thế nào để thay đổi cổng lắng nghe?

Chỉnh sửa `configs/config.yaml`:
```yaml
server:
  http_port: 9999  # Thay đổi từ mặc định 8888
```

### Cân nhắc bảo mật

#### Bảo mật chế độ debug

Chế độ debug có thể ghi lại thông tin nhạy cảm:
- Thông tin proxy đã giải mã
- URL yêu cầu
- Header xác thực

**Không bao giờ sử dụng chế độ debug trong production!**

#### Khuyến nghị cài đặt bảo mật

1. **Chỉ bind vào localhost** (nếu không phục vụ mạng)
2. **Sử dụng quy tắc firewall** để hạn chế truy cập
3. **Xoay vòng thông tin xác thực** thường xuyên
4. **Giám sát log** cho hoạt động đáng ngờ
5. **Giữ khóa riêng CA an toàn** nếu dùng MITM

### Nhận trợ giúp

Nếu bạn vẫn gặp vấn đề:

1. **Kiểm tra log** với chế độ debug bật
2. **Chạy script test** trong `scripts/test/`
3. **Tìm kiếm issue hiện có** trên GitHub
4. **Tạo issue mới** với:
   - Phiên bản SmartProxy
   - Cấu hình (không có dữ liệu nhạy cảm)
   - Log debug
   - Các bước tái tạo

### Các lệnh hữu ích

```bash
# Kiểm tra proxy đang chạy
ps aux | grep smartproxy

# Kiểm tra cổng lắng nghe
netstat -tlnp | grep 8888  # Linux
lsof -i :8888  # macOS

# Kiểm tra kết nối cơ bản
telnet localhost 8888

# Test với curl chi tiết
curl -v -x http://localhost:8888 http://httpbin.org/get

# Giám sát log thời gian thực
tail -f smartproxy.log | grep -E "ERROR|WARN"
```
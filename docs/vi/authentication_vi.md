# Hướng dẫn xác thực SmartProxy

## Tổng quan

SmartProxy sử dụng hệ thống xác thực đột phá cho phép cấu hình upstream proxy động thông qua thông tin xác thực. Điều này loại bỏ nhu cầu cấu hình upstream tĩnh trong file config và cho phép chọn proxy cho mỗi kết nối.

## Cách hoạt động

Thay vì cấu hình upstream proxy trong file cấu hình, bạn mã hóa chi tiết upstream trong thông tin xác thực proxy:

- **Username**: Schema của upstream proxy (`http` hoặc `socks5`)
- **Password**: Chi tiết upstream proxy đã mã hóa base64

## Định dạng xác thực

### Cho upstream KHÔNG CÓ xác thực

```
Định dạng: base64("host:port")
```

Ví dụ:
```bash
echo -n "proxy.example.com:8080" | base64
# Kết quả: cHJveHkuZXhhbXBsZS5jb206ODA4MA==
```

### Cho upstream CÓ xác thực

```
Định dạng: base64("host:port:username:password")
```

Ví dụ:
```bash
echo -n "proxy.example.com:8080:myuser:mypass" | base64
# Kết quả: cHJveHkuZXhhbXBsZS5jb206ODA4MDpteXVzZXI6bXlwYXNz
```

## Ví dụ sử dụng

### Dòng lệnh (curl)

```bash
# HTTP upstream không có xác thực
curl -x http://http:cHJveHkuZXhhbXBsZS5jb206ODA4MA==@localhost:8888 http://ipinfo.io

# SOCKS5 upstream với xác thực
curl -x http://socks5:cHJveHkuZXhhbXBsZS5jb206MTA4MDp1c2VyOnBhc3M=@localhost:8888 http://ipinfo.io
```

### Cấu hình trình duyệt

#### Chrome/Edge
1. Vào Settings → Advanced → System → Open proxy settings
2. Cấu hình:
   ```
   Server: localhost
   Port: 8888
   Username: http (hoặc socks5)
   Password: [chi tiết upstream đã mã hóa base64]
   ```

#### Firefox
1. Vào Settings → Network Settings
2. Manual proxy configuration:
   ```
   HTTP Proxy: localhost    Port: 8888
   HTTPS Proxy: localhost   Port: 8888
   ```
3. Nhập thông tin xác thực khi được nhắc

### Ngôn ngữ lập trình

#### Python
```python
import requests
import base64

# Mã hóa chi tiết upstream
upstream = "proxy.example.com:8080:myuser:mypass"
password = base64.b64encode(upstream.encode()).decode()

# Cấu hình proxy
proxies = {
    'http': f'http://http:{password}@localhost:8888',
    'https': f'http://http:{password}@localhost:8888'
}

response = requests.get('http://ipinfo.io', proxies=proxies)
print(response.text)
```

#### Node.js
```javascript
const axios = require('axios');

// Mã hóa chi tiết upstream
const upstream = "proxy.example.com:8080";
const password = Buffer.from(upstream).toString('base64');

// Cấu hình proxy
const proxyConfig = {
    host: 'localhost',
    port: 8888,
    auth: {
        username: 'http',
        password: password
    }
};

axios.get('http://ipinfo.io', {
    proxy: proxyConfig
}).then(response => {
    console.log(response.data);
});
```

#### Go
```go
import (
    "encoding/base64"
    "net/http"
    "net/url"
)

// Mã hóa chi tiết upstream
upstream := "proxy.example.com:8080"
password := base64.StdEncoding.EncodeToString([]byte(upstream))

// Cấu hình proxy
proxyURL, _ := url.Parse(fmt.Sprintf("http://http:%s@localhost:8888", password))
client := &http.Client{
    Transport: &http.Transport{
        Proxy: http.ProxyURL(proxyURL),
    },
}

resp, _ := client.Get("http://ipinfo.io")
```

## Tính năng nâng cao

### Nhiều upstream proxy

Sử dụng các upstream proxy khác nhau cho các yêu cầu khác nhau mà không cần khởi động lại SmartProxy:

```bash
# Proxy Mỹ
us_password=$(echo -n "us.proxy.com:8080" | base64)
curl -x http://http:${us_password}@localhost:8888 http://ipinfo.io

# Proxy Châu Âu
eu_password=$(echo -n "eu.proxy.com:8080" | base64)
curl -x http://http:${eu_password}@localhost:8888 http://ipinfo.io

# Proxy Châu Á với xác thực
asia_password=$(echo -n "asia.proxy.com:8080:user:pass" | base64)
curl -x http://http:${asia_password}@localhost:8888 http://ipinfo.io
```

### Xoay vòng proxy

Dễ dàng xoay vòng qua nhiều proxy:

```python
import random
import base64
import requests

proxies_list = [
    "proxy1.com:8080",
    "proxy2.com:8080:user:pass",
    "proxy3.com:8080"
]

# Chọn proxy ngẫu nhiên
upstream = random.choice(proxies_list)
password = base64.b64encode(upstream.encode()).decode()

# Sử dụng proxy đã chọn
proxies = {
    'http': f'http://http:{password}@localhost:8888',
    'https': f'http://http:{password}@localhost:8888'
}

response = requests.get('http://ipinfo.io', proxies=proxies)
```

## Hành vi xác thực trình duyệt

### HTTP Keep-Alive và xác thực

Trình duyệt sử dụng kết nối bền vững (HTTP keep-alive) có thể gây hành vi không mong muốn khi test:

1. **Xác thực chỉ được kiểm tra trên kết nối mới**
2. **Kết nối hiện tại tiếp tục hoạt động ngay cả với thông tin sai**
3. **Đây là hành vi chuẩn của HTTP/1.1 và HTTP/2**

### Kiểm tra xác thực trong trình duyệt

#### Phương pháp 1: Đóng tất cả kết nối
1. Đóng Chrome/Firefox hoàn toàn (tất cả cửa sổ)
2. Hoặc sử dụng: `chrome://net-internals/#sockets` → "Flush socket pools"
3. Mở lại trình duyệt và test với thông tin mới

#### Phương pháp 2: Profile mới
```bash
# Windows
chrome.exe --user-data-dir="%TEMP%\chrome_test" --proxy-server="http://localhost:8888"

# macOS
open -na "Google Chrome" --args --user-data-dir="/tmp/chrome_test" --proxy-server="http://localhost:8888"

# Linux
google-chrome --user-data-dir="/tmp/chrome_test" --proxy-server="http://localhost:8888"
```

#### Phương pháp 3: Chế độ ẩn danh/riêng tư
- Mở cửa sổ ẩn danh/riêng tư
- Cấu hình cài đặt proxy
- Kiểm tra xác thực

## Các vấn đề thường gặp và giải pháp

### "illegal base64 data at input byte X"

Lỗi này xảy ra khi:
- Base64 bị ngắt dòng với newline (thường tại 76 ký tự)
- Có khoảng trắng thừa
- Có ký tự không hợp lệ

**Giải pháp**: SmartProxy tự động xử lý base64 bị ngắt bằng cách loại bỏ khoảng trắng và newline.

Để tạo base64 đúng:
```bash
# Sử dụng -n để tránh newline cuối
echo -n "host:port:user:pass" | base64
```

### "LibreSSL: error:1404B42E:SSL routines:ST_CONNECT:tlsv1 alert protocol version"

**Nguyên nhân**: Sử dụng `https://` trong URL proxy

**Giải pháp**: Luôn sử dụng `http://` cho URL proxy:
```bash
# ✅ ĐÚNG - Luôn dùng http:// cho URL proxy (ngay cả cho trang HTTPS)
curl -x http://http:PASSWORD@localhost:8888 https://ipinfo.io

# ❌ SAI - Không dùng https:// cho URL proxy
curl -x https://http:PASSWORD@localhost:8888 http://ipinfo.io
```

### Yêu cầu xác thực (407)

**Nguyên nhân**: Không cung cấp thông tin xác thực

**Giải pháp**: Đảm bảo bạn cung cấp username và password trong yêu cầu

### Thông tin xác thực không hợp lệ (403)

**Nguyên nhân**: 
- Username sai (phải là `http` hoặc `socks5`)
- Mã hóa base64 không hợp lệ
- Chi tiết upstream proxy không chính xác

**Giải pháp**: Xác minh mã hóa của bạn:
```bash
# Kiểm tra base64
echo -n "proxy.example.com:8080" | base64
# Giải mã để xác minh
echo "cHJveHkuZXhhbXBsZS5jb206ODA4MA==" | base64 -d
```

## Cân nhắc bảo mật

1. **Base64 KHÔNG phải mã hóa** - Đó chỉ là encoding. Sử dụng kết nối HTTPS đến SmartProxy nếu bảo mật là mối quan tâm.

2. **Lưu trữ thông tin xác thực** - Lưu trữ thông tin xác thực đã mã hóa base64 an toàn:
   - Sử dụng biến môi trường
   - Sử dụng hệ thống quản lý khóa an toàn
   - Không commit thông tin xác thực vào quản lý phiên bản

3. **Kiểm soát truy cập** - Cân nhắc các biện pháp bảo mật bổ sung:
   - Chỉ chạy SmartProxy trên localhost
   - Sử dụng quy tắc firewall để hạn chế truy cập
   - Triển khai giới hạn tần suất cho xác thực thất bại

4. **Chế độ HTTPS MITM** - Khi bật, yêu cầu xác thực cho tất cả yêu cầu để đảm bảo sử dụng proxy an toàn

## Thực hành tốt nhất

1. **Sử dụng mật khẩu mạnh** - Cho upstream proxy hỗ trợ xác thực

2. **Giám sát log** - Bật logging để theo dõi xác thực thất bại:
   ```yaml
   logging:
     level: info
   ```

3. **Kiểm tra kỹ lưỡng** - Xác minh xác thực hoạt động trước khi sử dụng production

4. **Xoay vòng thông tin xác thực** - Thay đổi thông tin xác thực upstream proxy thường xuyên

5. **Sử dụng HTTPS** - Khi kết nối đến SmartProxy qua mạng:
   ```nginx
   # Nginx reverse proxy với SSL
   server {
       listen 443 ssl;
       ssl_certificate /path/to/cert.pem;
       ssl_certificate_key /path/to/key.pem;
       
       location / {
           proxy_pass http://localhost:8888;
           proxy_set_header Proxy-Authorization $http_proxy_authorization;
       }
   }
   ```
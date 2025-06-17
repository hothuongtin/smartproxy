# Xác thực Smart Proxy

SmartProxy hiện hỗ trợ cấu hình upstream động thông qua thông tin xác thực. Thay vì hardcode cài đặt upstream proxy trong file cấu hình, bạn có thể mã hóa chúng trong thông tin đăng nhập.

## Cách hoạt động

Khi kết nối tới SmartProxy, cung cấp thông tin xác thực với:
- **Username**: Schema của upstream proxy (`http` hoặc `socks5`)
- **Password**: Chi tiết upstream proxy được mã hóa base64

## Định dạng cấu hình

### Upstream KHÔNG có xác thực:
```
Định dạng: base64("host:port")
```

Ví dụ:
```bash
echo -n "na.lunaproxy.com:12233" | base64
# Kết quả: bmEubHVuYXByb3h5LmNvbToxMjIzMw==
```

### Upstream CÓ xác thực:
```
Định dạng: base64("host:port:username:password")
```

Ví dụ:
```bash
echo -n "na.lunaproxy.com:12233:user-usa1az_H5xzU:Ajs76x6a76ax" | base64
# Kết quả: bmEubHVuYXByb3h5LmNvbToxMjIzMzp1c2VyLXVzYTFhel9INXh6VTpBanM3Nng2YTc2YXg=
```

## Ví dụ sử dụng

### cURL
```bash
# HTTP upstream không có auth
curl -x http://http:bmEubHVuYXByb3h5LmNvbToxMjIzMw==@localhost:8888 http://ipinfo.io

# SOCKS5 upstream có auth
curl -x http://socks5:bmEubHVuYXByb3h5LmNvbToxMjIzMzp1c2VyOnBhc3M=@localhost:8888 http://ipinfo.io
```

### Chrome/Firefox
Cài đặt proxy:
```
Server: localhost
Port: 8888
Username: http (hoặc socks5)
Password: [chi tiết upstream đã mã hóa base64]
```

### Python Requests
```python
import requests
import base64

# Mã hóa chi tiết upstream
upstream = "na.lunaproxy.com:12233:myuser:mypass"
password = base64.b64encode(upstream.encode()).decode()

# Cấu hình proxy
proxies = {
    'http': f'http://http:{password}@localhost:8888',
    'https': f'http://http:{password}@localhost:8888'
}

response = requests.get('http://ipinfo.io', proxies=proxies)
print(response.text)
```

### Node.js
```javascript
const axios = require('axios');

// Mã hóa chi tiết upstream
const upstream = "na.lunaproxy.com:12233";
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

## Tính năng nâng cao

### Hỗ trợ nhiều Upstream
Bạn có thể sử dụng các upstream proxy khác nhau cho các request khác nhau bằng cách thay đổi thông tin xác thực:

```bash
# Proxy US
curl -x http://http:dXMucHJveHkuY29tOjgwODA=@localhost:8888 http://ipinfo.io

# Proxy EU
curl -x http://http:ZXUucHJveHkuY29tOjgwODA=@localhost:8888 http://ipinfo.io
```

### Định tuyến tự động
SmartProxy vẫn duy trì định tuyến thông minh:
- File tĩnh (.js, .css, hình ảnh) → Kết nối trực tiếp
- Domain CDN → Kết nối trực tiếp
- Domain quảng cáo → Chặn
- Request khác → Upstream proxy bạn chỉ định

## Bảo mật

1. **Base64 KHÔNG phải mã hóa** - Chỉ là encoding. Sử dụng HTTPS kết nối tới SmartProxy nếu cần bảo mật.

2. **Lưu trữ thông tin** - Lưu trữ thông tin base64 an toàn, không để trong file plain text.

3. **Kiểm soát truy cập** - Cân nhắc thêm kiểm soát truy cập nếu chạy SmartProxy trên mạng public.

## Xử lý lỗi

### Lỗi thông tin không hợp lệ
Đảm bảo mã hóa base64 đúng:
```bash
# Đúng
echo -n "host:port" | base64

# Sai (có thêm newline)
echo "host:port" | base64
```

### Kết nối bị từ chối
Kiểm tra chi tiết upstream proxy đúng và proxy có thể truy cập được.

### Yêu cầu xác thực
Đảm bảo bạn đang cung cấp header Proxy-Authorization với request.

### Chrome vẫn hoạt động với mật khẩu sai
Chrome cache thông tin xác thực. Xem [BROWSER_AUTH_NOTES_vi.md](BROWSER_AUTH_NOTES_vi.md) để biết cách xử lý.
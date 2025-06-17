# SmartProxy FAQ - Câu hỏi thường gặp

## H: Tại sao curl báo lỗi xác thực nhưng Chrome vẫn hoạt động?

**Đ:** Chrome sử dụng kết nối persistent (keep-alive). Khi bạn nhập sai mật khẩu:
- curl tạo kết nối mới → kiểm tra xác thực → báo lỗi
- Chrome tái sử dụng kết nối cũ đã xác thực → vẫn hoạt động

Đây là hành vi chuẩn của HTTP, không phải lỗi.

**Giải pháp:**
1. Đóng hoàn toàn Chrome để ngắt kết nối cũ
2. Hoặc vào chrome://net-internals/#sockets → "Flush socket pools"
3. Mở lại Chrome và nhập thông tin mới

Xem [BROWSER_AUTH_NOTES_vi.md](BROWSER_AUTH_NOTES_vi.md) để biết chi tiết.

## H: Tại sao ip-api.com trả về lỗi "SSL unavailable"?

**Đ:** API miễn phí của ip-api.com không hỗ trợ HTTPS. Sử dụng HTTP thay vì:
- ❌ Sai: `https://ip-api.com/json`
- ✅ Đúng: `http://ip-api.com/json`

Đây không phải lỗi của proxy - đó là giới hạn của gói miễn phí của họ.

## H: Làm thế nào để truy cập các trang HTTPS khi tắt MITM?

**Đ:** Khi `https_mitm: false` (mặc định), SmartProxy tunnel các kết nối HTTPS mà không giải mã. Điều này hoạt động hoàn hảo cho tất cả các trang HTTPS tiêu chuẩn. Proxy sẽ:
- Thiết lập tunnel sử dụng phương thức CONNECT
- Định tuyến kết nối HTTPS qua upstream proxy đã cấu hình (nếu có)
- Chuyển tiếp traffic được mã hóa mà không kiểm tra
- Duy trì mã hóa end-to-end

**Lưu ý:** Kết nối HTTPS giờ đây sử dụng đúng upstream proxy được cấu hình qua xác thực, ngoại trừ các domain CDN sẽ dùng kết nối trực tiếp để có hiệu suất tốt hơn.

## H: Sự khác biệt giữa MITM bật và tắt là gì?

### MITM Tắt (Mặc định):
- ✅ Không có cảnh báo chứng chỉ
- ✅ Quyền riêng tư đầy đủ - proxy không thể xem nội dung HTTPS
- ✅ Hoạt động ngay lập tức mà không cần cấu hình
- ❌ Không thể chặn quảng cáo trên trang HTTPS
- ❌ Không thể kiểm tra traffic HTTPS

### MITM Bật:
- ✅ Có thể chặn quảng cáo trên trang HTTPS
- ✅ Có thể kiểm tra và lọc nội dung HTTPS
- ❌ Yêu cầu cài đặt chứng chỉ CA
- ❌ Có thể có vấn đề về quyền riêng tư
- ❌ Cảnh báo chứng chỉ nếu CA không được tin cậy

## H: Làm thế nào để kiểm tra proxy có hoạt động không?

Chạy script test đi kèm:
```bash
./test_proxy.sh
```

Hoặc kiểm tra thủ công:
```bash
# Test HTTP
curl -x http://localhost:8888 http://httpbin.org/get

# Test HTTPS
curl -x http://localhost:8888 https://httpbin.org/get

# Kiểm tra IP qua proxy
curl -x http://localhost:8888 http://ip-api.com/json
```

## H: Tại sao một số request HTTPS nhanh hơn các request khác?

SmartProxy tự động sử dụng kết nối trực tiếp cho:
- File tĩnh (hình ảnh, CSS, JS, v.v.)
- Domain CDN đã biết
- Nội dung không cần lọc

Điều này bỏ qua upstream proxy để có hiệu suất tốt hơn.

## H: Làm thế nào để tắt chặn quảng cáo?

Chỉnh sửa `config.yaml`:
```yaml
ad_blocking:
  enabled: false
```

## H: Có thể sử dụng proxy SOCKS5 làm upstream không?

Có! Cấu hình trong `config.yaml`:
```yaml
upstream:
  proxy_url: "socks5://127.0.0.1:1080"
  username: "tùy chọn"
  password: "tùy chọn"
```

## H: Làm thế nào để thay đổi cổng lắng nghe?

Chỉnh sửa `config.yaml`:
```yaml
server:
  http_port: 9999  # Thay đổi từ mặc định 8888
```

## H: Proxy không khởi động, báo lỗi gì?

### Lỗi: "Upstream proxy is required"
- **Nguyên nhân**: Không cấu hình upstream proxy
- **Giải pháp**: Thêm `upstream.proxy_url` trong config.yaml

### Lỗi: "bind: address already in use"
- **Nguyên nhân**: Cổng đã được sử dụng
- **Giải pháp**: 
  ```bash
  make kill  # Tắt proxy cũ
  make run   # Chạy lại
  ```

### Lỗi chứng chỉ TLS
- **Nguyên nhân**: MITM được bật nhưng client không tin cậy CA
- **Giải pháp**: Cài đặt chứng chỉ CA hoặc tắt MITM

## H: Làm thế nào để giám sát hiệu suất?

1. Xem log với màu sắc:
```bash
tail -f proxy.log
```

2. Kiểm tra số lượng kết nối:
```bash
netstat -an | grep 8888 | wc -l
```

3. Monitor CPU và memory:
```bash
top -p $(pgrep smartproxy)
```

## H: Docker image nào nên sử dụng?

- **Production**: Sử dụng image distroless mặc định (~15MB)
- **Debug**: Sử dụng image Alpine nếu cần shell
- **Tối thiểu tuyệt đối**: Sử dụng scratch image (~12MB)

## H: Có thể chạy nhiều instance không?

Có, nhưng cần:
1. Sử dụng cổng khác nhau cho mỗi instance
2. Sử dụng file config khác nhau
3. Cân nhắc sử dụng load balancer phía trước

## H: Làm thế nào để đóng góp?

1. Fork repository
2. Tạo branch mới: `git checkout -b feature/ten-tinh-nang`
3. Commit: `git commit -m 'Thêm tính năng X'`
4. Push: `git push origin feature/ten-tinh-nang`
5. Tạo Pull Request

## H: Hỗ trợ IPv6 không?

Có, SmartProxy hỗ trợ dual-stack (IPv4 và IPv6) mặc định.
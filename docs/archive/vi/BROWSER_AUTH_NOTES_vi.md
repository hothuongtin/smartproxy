# Hành vi Xác thực của Trình duyệt

## Sự khác biệt giữa Chrome và curl

Khi test xác thực proxy, bạn có thể thấy hành vi khác nhau giữa curl và Chrome.

### Vấn đề
- curl hiển thị lỗi xác thực với mật khẩu sai: `errorMsg: Account password authentication failed`
- Chrome vẫn hoạt động bình thường với cùng mật khẩu sai đó

### Nguyên nhân thực sự

**Kết nối HTTP Persistent (Keep-Alive)**
- Chrome sử dụng kết nối persistent/keep-alive
- Xác thực chỉ kiểm tra khi thiết lập kết nối mới
- Kết nối đã thiết lập sẽ được tái sử dụng cho nhiều request
- curl mặc định tạo kết nối mới cho mỗi request

Đây là hành vi chuẩn của HTTP/1.1 và HTTP/2, không phải lỗi.

### Cách xử lý

1. **Đóng tất cả kết nối Chrome**
   - Đóng hoàn toàn Chrome (tất cả cửa sổ)
   - Hoặc: chrome://net-internals/#sockets → "Close idle sockets"
   - Hoặc: chrome://net-internals/#sockets → "Flush socket pools"

2. **Test với kết nối mới**
   - Mở Chrome lại
   - Nhập thông tin xác thực mới
   - Chrome sẽ sử dụng thông tin mới cho kết nối mới

3. **Force new connection trong curl**
   ```bash
   # Disable keep-alive
   curl -x http://http:wrongpassword@localhost:8888 \
        -H "Connection: close" \
        http://httpbin.org/ip
   ```

### Cách Test Đúng

#### Trong Chrome
1. Đóng hoàn toàn Chrome
2. Xóa dữ liệu duyệt web:
   - Nhấn `Ctrl+Shift+Delete` (Windows/Linux) hoặc `Cmd+Shift+Delete` (Mac)
   - Chọn "Toàn bộ thời gian"
   - Chọn "Cookie và dữ liệu trang web khác" và "Hình ảnh và tệp đã lưu trong bộ nhớ cache"
   - Nhấn "Xóa dữ liệu"
3. Mở Chrome với profile mới:
   ```bash
   # Windows
   chrome.exe --user-data-dir="%TEMP%\chrome_test" --proxy-server="http://localhost:8888"
   
   # macOS
   open -na "Google Chrome" --args --user-data-dir="/tmp/chrome_test" --proxy-server="http://localhost:8888"
   
   # Linux
   google-chrome --user-data-dir="/tmp/chrome_test" --proxy-server="http://localhost:8888"
   ```
4. Nhập thông tin sai khi được yêu cầu
5. Chrome sẽ hiển thị lỗi xác thực

#### Trong curl
```bash
# Test với mật khẩu sai
curl -x http://http:wrongpassword@localhost:8888 http://httpbin.org/ip

# Test với định dạng đúng
correct_password=$(echo -n "proxy.example.com:8080" | base64)
curl -x http://http:${correct_password}@localhost:8888 http://httpbin.org/ip
```

### Chế độ Debug

Bật debug logging để xem chi tiết:
```bash
make debug
```

Tìm các dòng log:
```
DEBUG Incoming request method=CONNECT
DEBUG No authentication for CONNECT host=httpbin.org:443
DEBUG Failed to parse upstream from CONNECT auth error=...
```

### Ý nghĩa Bảo mật

Xác thực nghiêm ngặt của curl mới là hành vi đúng. Cache của Chrome tiện lợi nhưng có thể che giấu lỗi khi test.

Cho môi trường production:
- Luôn test với profile trình duyệt mới
- Theo dõi log proxy cho lỗi xác thực
- Cân nhắc giới hạn số lần thử sai
- Dùng mật khẩu mạnh, riêng biệt cho mỗi upstream
# Hướng dẫn cấu hình HTTPS

## Tổng quan

SmartProxy hỗ trợ hai chế độ xử lý traffic HTTPS:

1. **Chế độ Tunneling** (mặc định) - Kết nối HTTPS được tunnel mà không giải mã
2. **Chế độ MITM** - Kết nối HTTPS được giải mã để kiểm tra và lọc

## Chế độ Tunneling (Khuyến nghị)

Đây là chế độ mặc định và an toàn nhất. Traffic HTTPS được tunnel qua proxy mà không bị giải mã.

**config.yaml:**
```yaml
server:
  https_mitm: false  # Đây là mặc định
```

**Ưu điểm:**
- Không có cảnh báo chứng chỉ
- Mã hóa end-to-end được duy trì đầy đủ
- Hoạt động ngay lập tức

**Hạn chế:**
- Không thể kiểm tra nội dung HTTPS
- Chặn quảng cáo chỉ hoạt động trên trang HTTP
- Không thể áp dụng lọc nội dung cho HTTPS

## Chế độ MITM (Nâng cao)

Chế độ này cho phép proxy giải mã và kiểm tra traffic HTTPS. Yêu cầu cài đặt chứng chỉ CA trên thiết bị client và xác thực cho tất cả các yêu cầu.

### Bước 1: Tạo chứng chỉ CA

```bash
./generate_ca.sh
```

Lệnh này tạo ra:
- `certs/ca.crt` - Chứng chỉ CA (cài đặt trên client)
- `certs/ca.key` - Private key (giữ bảo mật)

### Bước 2: Cập nhật cấu hình

**config.yaml:**
```yaml
server:
  https_mitm: true
  ca_cert: "certs/ca.crt"
  ca_key: "certs/ca.key"
```

### Bước 3: Cài đặt chứng chỉ CA trên Client

#### macOS:
1. Double-click vào `certs/ca.crt`
2. Thêm vào System keychain
3. Trust cho SSL

#### Windows:
1. Double-click vào `certs/ca.crt`
2. Cài đặt vào "Trusted Root Certification Authorities"

#### Linux:
```bash
sudo cp certs/ca.crt /usr/local/share/ca-certificates/smartproxy.crt
sudo update-ca-certificates
```

#### iOS:
1. Gửi email hoặc AirDrop `ca.crt` đến thiết bị
2. Cài đặt profile
3. Vào Settings > General > About > Certificate Trust Settings
4. Bật cho SmartProxy CA

#### Android:
1. Copy `ca.crt` vào thiết bị
2. Settings > Security > Install from storage
3. Chọn "CA certificate"

## Chọn chế độ phù hợp

### Sử dụng chế độ Tunneling khi:
- Bạn muốn bảo mật tối đa
- Không cần kiểm tra nội dung HTTPS
- Muốn không cần cấu hình trên client
- Quyền riêng tư là ưu tiên

### Sử dụng chế độ MITM khi:
- Cần chặn quảng cáo trên trang HTTPS
- Muốn kiểm tra traffic HTTPS
- Cần lọc nội dung cho HTTPS
- Có thể quản lý chứng chỉ CA trên tất cả client
- Cần phát hiện file tĩnh cho HTTPS
- Cần định tuyến thông minh đầy đủ cho yêu cầu HTTPS

## Xác thực trong chế độ MITM

Khi MITM được bật, tất cả các yêu cầu cần xác thực proxy hợp lệ:
- Đảm bảo chỉ người dùng được ủy quyền mới có thể kiểm tra traffic HTTPS
- Ngăn chặn các cuộc tấn công MITM trái phép
- Thông tin xác thực cũng cấu hình định tuyến upstream proxy

## Cân nhắc về bảo mật

1. **Giữ ca.key an toàn** - Bất kỳ ai có file này đều có thể mạo danh mọi website
2. **Sử dụng mật khẩu mạnh** - Bảo vệ file CA key
3. **Giới hạn tin cậy CA** - Chỉ cài đặt trên thiết bị bạn kiểm soát
4. **Xoay vòng thường xuyên** - Tạo lại chứng chỉ định kỳ
5. **Kiểm tra log** - Giám sát log proxy để phát hiện hoạt động đáng ngờ

## Khắc phục sự cố

### Cảnh báo chứng chỉ
- Đảm bảo chứng chỉ CA được cài đặt và tin cậy đúng cách
- Kiểm tra ngày hết hạn chứng chỉ
- Xác minh hostname khớp

### Kết nối thất bại
- Kiểm tra firewall rules
- Xác minh proxy đang chạy
- Test với `curl --proxy http://localhost:8888 https://example.com`

### Vấn đề hiệu suất
- Tắt MITM nếu không cần thiết
- Kiểm tra CPU usage trong quá trình TLS operations
- Cân nhắc hardware acceleration

## Lệnh hữu ích

### Kiểm tra chứng chỉ
```bash
# Xem thông tin chứng chỉ
openssl x509 -in certs/ca.crt -text -noout

# Kiểm tra ngày hết hạn
openssl x509 -in certs/ca.crt -dates -noout
```

### Test HTTPS qua proxy
```bash
# Test với curl
curl -x http://localhost:8888 https://httpbin.org/get

# Test với wget
https_proxy=http://localhost:8888 wget -O- https://httpbin.org/get
```

### Debug TLS
```bash
# Bật debug log
export SSLKEYLOGFILE=~/sslkeys.log
# Sau đó sử dụng Wireshark để phân tích
```
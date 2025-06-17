# Smart Proxy Authentication

SmartProxy now supports dynamic upstream configuration through authentication credentials. Instead of hardcoding upstream proxy settings in the configuration file, you can encode them in the authentication credentials.

## How It Works

When connecting to SmartProxy, provide authentication credentials where:
- **Username**: The upstream proxy schema (`http` or `socks5`)
- **Password**: Base64 encoded upstream proxy details

## Configuration Format

### For upstream WITHOUT authentication:
```
Format: base64("host:port")
```

Example:
```bash
echo -n "na.lunaproxy.com:12233" | base64
# Output: bmEubHVuYXByb3h5LmNvbToxMjIzMw==
```

### For upstream WITH authentication:
```
Format: base64("host:port:username:password")
```

Example:
```bash
echo -n "na.lunaproxy.com:12233:user-usa1az_H5xzU:Ajs76x6a76ax" | base64
# Output: bmEubHVuYXByb3h5LmNvbToxMjIzMzp1c2VyLXVzYTFhel9INXh6VTpBanM3Nng2YTc2YXg=
```

## Usage Examples

### cURL
```bash
# HTTP upstream without auth
curl -x http://http:bmEubHVuYXByb3h5LmNvbToxMjIzMw==@localhost:8888 http://ipinfo.io

# SOCKS5 upstream with auth
curl -x http://socks5:bmEubHVuYXByb3h5LmNvbToxMjIzMzp1c2VyOnBhc3M=@localhost:8888 http://ipinfo.io
```

### Chrome/Firefox
Set proxy to:
```
Server: localhost
Port: 8888
Username: http (or socks5)
Password: [your base64 encoded upstream details]
```

### Python Requests
```python
import requests
import base64

# Encode upstream details
upstream = "na.lunaproxy.com:12233:myuser:mypass"
password = base64.b64encode(upstream.encode()).decode()

# Configure proxy
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

// Encode upstream details
const upstream = "na.lunaproxy.com:12233";
const password = Buffer.from(upstream).toString('base64');

// Configure proxy
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

## Advanced Features

### Multiple Upstream Support
You can use different upstream proxies for different requests by changing the authentication credentials:

```bash
# US proxy
curl -x http://http:dXMucHJveHkuY29tOjgwODA=@localhost:8888 http://ipinfo.io

# EU proxy
curl -x http://http:ZXUucHJveHkuY29tOjgwODA=@localhost:8888 http://ipinfo.io
```

### Automatic Routing
SmartProxy still maintains its intelligent routing:
- Static files (.js, .css, images) → Direct connection
- CDN domains → Direct connection
- Ad domains → Blocked
- Other requests → Your specified upstream proxy

## Security Considerations

1. **Base64 is NOT encryption** - It's just encoding. Use HTTPS connections to SmartProxy if security is a concern.

2. **Credential Storage** - Store your base64 encoded credentials securely, not in plain text files.

3. **Access Control** - Consider implementing additional access control if running SmartProxy on a public network.

## Troubleshooting

### Invalid credentials error
Make sure your base64 encoding is correct:
```bash
# Correct
echo -n "host:port" | base64

# Wrong (includes newline)
echo "host:port" | base64
```

### Connection refused
Verify the upstream proxy details are correct and the proxy is accessible.

### Authentication required
Make sure you're providing the Proxy-Authorization header with your requests.

### Chrome still works with wrong password
Chrome caches proxy credentials. See [BROWSER_AUTH_NOTES.md](BROWSER_AUTH_NOTES.md) for details.
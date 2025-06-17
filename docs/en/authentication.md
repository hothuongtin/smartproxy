# SmartProxy Authentication Guide

## Overview

SmartProxy uses an innovative authentication system that allows dynamic upstream proxy configuration through authentication credentials. This eliminates the need for static upstream configuration in config files and enables per-connection proxy selection.

## How It Works

Instead of configuring upstream proxies in the configuration file, you encode the upstream details in the proxy authentication credentials:

- **Username**: The upstream proxy schema (`http` or `socks5`)
- **Password**: Base64 encoded upstream proxy details

## Authentication Format

### For Upstream WITHOUT Authentication

```
Format: base64("host:port")
```

Example:
```bash
echo -n "proxy.example.com:8080" | base64
# Output: cHJveHkuZXhhbXBsZS5jb206ODA4MA==
```

### For Upstream WITH Authentication

```
Format: base64("host:port:username:password")
```

Example:
```bash
echo -n "proxy.example.com:8080:myuser:mypass" | base64
# Output: cHJveHkuZXhhbXBsZS5jb206ODA4MDpteXVzZXI6bXlwYXNz
```

## Usage Examples

### Command Line (curl)

```bash
# HTTP upstream without auth
curl -x http://http:cHJveHkuZXhhbXBsZS5jb206ODA4MA==@localhost:8888 http://ipinfo.io

# SOCKS5 upstream with auth
curl -x http://socks5:cHJveHkuZXhhbXBsZS5jb206MTA4MDp1c2VyOnBhc3M=@localhost:8888 http://ipinfo.io
```

### Browser Configuration

#### Chrome/Edge
1. Go to Settings → Advanced → System → Open proxy settings
2. Configure:
   ```
   Server: localhost
   Port: 8888
   Username: http (or socks5)
   Password: [your base64 encoded upstream details]
   ```

#### Firefox
1. Go to Settings → Network Settings
2. Manual proxy configuration:
   ```
   HTTP Proxy: localhost    Port: 8888
   HTTPS Proxy: localhost   Port: 8888
   ```
3. Enter credentials when prompted

### Programming Languages

#### Python
```python
import requests
import base64

# Encode upstream details
upstream = "proxy.example.com:8080:myuser:mypass"
password = base64.b64encode(upstream.encode()).decode()

# Configure proxy
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

// Encode upstream details
const upstream = "proxy.example.com:8080";
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

#### Go
```go
import (
    "encoding/base64"
    "net/http"
    "net/url"
)

// Encode upstream details
upstream := "proxy.example.com:8080"
password := base64.StdEncoding.EncodeToString([]byte(upstream))

// Configure proxy
proxyURL, _ := url.Parse(fmt.Sprintf("http://http:%s@localhost:8888", password))
client := &http.Client{
    Transport: &http.Transport{
        Proxy: http.ProxyURL(proxyURL),
    },
}

resp, _ := client.Get("http://ipinfo.io")
```

## Advanced Features

### Multiple Upstream Proxies

Use different upstream proxies for different requests without restarting SmartProxy:

```bash
# US proxy
us_password=$(echo -n "us.proxy.com:8080" | base64)
curl -x http://http:${us_password}@localhost:8888 http://ipinfo.io

# EU proxy
eu_password=$(echo -n "eu.proxy.com:8080" | base64)
curl -x http://http:${eu_password}@localhost:8888 http://ipinfo.io

# Asia proxy with auth
asia_password=$(echo -n "asia.proxy.com:8080:user:pass" | base64)
curl -x http://http:${asia_password}@localhost:8888 http://ipinfo.io
```

### Proxy Rotation

Easily rotate through multiple proxies:

```python
import random
import base64
import requests

proxies_list = [
    "proxy1.com:8080",
    "proxy2.com:8080:user:pass",
    "proxy3.com:8080"
]

# Select random proxy
upstream = random.choice(proxies_list)
password = base64.b64encode(upstream.encode()).decode()

# Use selected proxy
proxies = {
    'http': f'http://http:{password}@localhost:8888',
    'https': f'http://http:{password}@localhost:8888'
}

response = requests.get('http://ipinfo.io', proxies=proxies)
```

## Browser Authentication Behavior

### HTTP Keep-Alive and Authentication

Browsers use persistent connections (HTTP keep-alive) which can cause unexpected behavior during testing:

1. **Authentication is checked only on new connections**
2. **Existing connections continue to work even with wrong credentials**
3. **This is standard HTTP/1.1 and HTTP/2 behavior**

### Testing Authentication in Browsers

#### Method 1: Close All Connections
1. Close Chrome/Firefox completely (all windows)
2. Or use: `chrome://net-internals/#sockets` → "Flush socket pools"
3. Reopen browser and test with new credentials

#### Method 2: Fresh Profile
```bash
# Windows
chrome.exe --user-data-dir="%TEMP%\chrome_test" --proxy-server="http://localhost:8888"

# macOS
open -na "Google Chrome" --args --user-data-dir="/tmp/chrome_test" --proxy-server="http://localhost:8888"

# Linux
google-chrome --user-data-dir="/tmp/chrome_test" --proxy-server="http://localhost:8888"
```

#### Method 3: Incognito/Private Mode
- Open incognito/private window
- Configure proxy settings
- Test authentication

## Common Issues and Solutions

### "illegal base64 data at input byte X"

This error occurs when:
- Base64 is wrapped with newlines (common at 76 characters)
- Extra whitespace is added
- Invalid characters are present

**Solution**: SmartProxy automatically handles wrapped base64 by removing whitespace and newlines.

To create proper base64:
```bash
# Use -n to avoid trailing newline
echo -n "host:port:user:pass" | base64
```

### "LibreSSL: error:1404B42E:SSL routines:ST_CONNECT:tlsv1 alert protocol version"

**Cause**: Using `https://` in the proxy URL

**Solution**: Always use `http://` for the proxy URL:
```bash
# ✅ CORRECT - Always use http:// for proxy URL (even for HTTPS sites)
curl -x http://http:PASSWORD@localhost:8888 https://ipinfo.io

# ❌ WRONG - Don't use https:// for proxy URL
curl -x https://http:PASSWORD@localhost:8888 http://ipinfo.io
```

### Authentication Required (407)

**Cause**: No credentials provided

**Solution**: Ensure you're providing username and password in the request

### Invalid Credentials (403)

**Cause**: 
- Wrong username (must be `http` or `socks5`)
- Invalid base64 encoding
- Upstream proxy details incorrect

**Solution**: Verify your encoding:
```bash
# Test your base64
echo -n "proxy.example.com:8080" | base64
# Decode to verify
echo "cHJveHkuZXhhbXBsZS5jb206ODA4MA==" | base64 -d
```

## Security Considerations

1. **Base64 is NOT encryption** - It's just encoding. Use HTTPS connections to SmartProxy if security is a concern.

2. **Credential Storage** - Store your base64 encoded credentials securely:
   - Use environment variables
   - Use secure key management systems
   - Don't commit credentials to version control

3. **Access Control** - Consider additional security measures:
   - Run SmartProxy on localhost only
   - Use firewall rules to restrict access
   - Implement rate limiting for failed auth attempts

4. **HTTPS MITM Mode** - When enabled, requires authentication for all requests to ensure secure proxy usage

## Best Practices

1. **Use Strong Passwords** - For upstream proxies that support authentication

2. **Monitor Logs** - Enable logging to track authentication failures:
   ```yaml
   logging:
     level: info
   ```

3. **Test Thoroughly** - Verify authentication works before production use

4. **Rotate Credentials** - Change upstream proxy credentials regularly

5. **Use HTTPS** - When connecting to SmartProxy over a network:
   ```nginx
   # Nginx reverse proxy with SSL
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
# HTTPS Upstream Routing in SmartProxy

## Overview

SmartProxy now properly routes HTTPS connections through the configured upstream proxy when using smart authentication mode. This document explains how HTTPS connections are handled.

## How It Works

### 1. Client HTTPS Request
When a client makes an HTTPS request:
```
Client → CONNECT example.com:443 HTTP/1.1
         Proxy-Authorization: Basic [credentials]
```

### 2. Authentication Parsing
SmartProxy extracts the upstream proxy configuration from the authentication:
- **Username**: Protocol (`http` or `socks5`)
- **Password**: Base64-encoded upstream proxy details

### 3. Connection Routing

#### For Regular Domains:
```
Client → SmartProxy → Upstream Proxy → Target Server
```
The HTTPS connection is tunneled through the configured upstream proxy.

#### For CDN Domains:
```
Client → SmartProxy → Target Server (Direct)
```
CDN domains bypass the upstream proxy for better performance.

### 4. Implementation Details

#### HTTP Upstream Proxy
For HTTP upstream proxies, SmartProxy:
1. Connects to the upstream proxy
2. Sends a CONNECT request to establish the tunnel
3. Includes proxy authentication if configured
4. Forwards the encrypted traffic

#### SOCKS5 Upstream Proxy
For SOCKS5 upstream proxies, SmartProxy:
1. Uses the Go SOCKS5 client library
2. Establishes a SOCKS5 connection with authentication
3. Tunnels the HTTPS traffic through the SOCKS5 proxy

## Configuration Examples

### HTTP Proxy without Authentication
```bash
# Upstream: proxy.example.com:8080
username="http"
password=$(echo -n "proxy.example.com:8080" | base64)
curl -x "http://${username}:${password}@localhost:8888" https://example.com
```

### HTTP Proxy with Authentication
```bash
# Upstream: proxy.example.com:8080 with user:pass
username="http"
password=$(echo -n "proxy.example.com:8080:user:pass" | base64)
curl -x "http://${username}:${password}@localhost:8888" https://example.com
```

### SOCKS5 Proxy
```bash
# Upstream: socks5.example.com:1080
username="socks5"
password=$(echo -n "socks5.example.com:1080" | base64)
curl -x "http://${username}:${password}@localhost:8888" https://example.com
```

## Testing

Use the provided test script to verify HTTPS upstream routing:
```bash
./scripts/test/test_https_upstream.sh
```

## Performance Considerations

1. **CDN Optimization**: CDN domains (e.g., cdn.jsdelivr.net, cdnjs.cloudflare.com) automatically use direct connections
2. **Connection Pooling**: Upstream connections are pooled and reused for efficiency
3. **Timeout Handling**: Default 30-second timeout for upstream connections

## Troubleshooting

### HTTPS Connection Fails
- Verify upstream proxy is accessible
- Check upstream proxy supports HTTPS/CONNECT method
- Ensure authentication credentials are correct

### Connection Times Out
- Check network connectivity to upstream proxy
- Verify upstream proxy is not blocking your IP
- Try increasing timeout in configs/config.yaml

### CDN Still Uses Upstream
- Check if domain is in the CDN list in configs/config.yaml
- Add custom CDN domains to direct_domains list

## Security Notes

1. **End-to-End Encryption**: HTTPS traffic remains encrypted through the entire chain
2. **No MITM**: With `https_mitm: false` (default), SmartProxy cannot see HTTPS content
3. **Credential Protection**: Use HTTPS to SmartProxy if running on untrusted networks
# SmartProxy FAQ

## Q: Why does ip-api.com return "SSL unavailable" error?

**A:** The free ip-api.com API doesn't support HTTPS. Use HTTP instead:
- ❌ Wrong: `https://ip-api.com/json`
- ✅ Correct: `http://ip-api.com/json`

This is not a proxy issue - it's a limitation of their free tier.

## Q: How can I access HTTPS sites with MITM disabled?

**A:** When `https_mitm: false` (default), SmartProxy tunnels HTTPS connections without decryption. This works perfectly for all standard HTTPS sites. The proxy:
- Establishes a tunnel using the CONNECT method
- Forwards encrypted traffic without inspection
- Maintains end-to-end encryption

## Q: What's the difference between MITM enabled vs disabled?

### MITM Disabled (Default):
- ✅ No certificate warnings
- ✅ Full privacy - proxy can't see HTTPS content
- ✅ Works immediately without configuration
- ❌ Can't block ads on HTTPS sites
- ❌ Can't inspect HTTPS traffic

### MITM Enabled:
- ✅ Can block ads on HTTPS sites
- ✅ Can inspect and filter HTTPS content
- ❌ Requires CA certificate installation
- ❌ Potential privacy concerns
- ❌ Certificate warnings if CA not trusted

## Q: How do I test if the proxy is working?

Run the included test script:
```bash
./test_proxy.sh
```

Or manually test:
```bash
# Test HTTP
curl -x http://localhost:8888 http://httpbin.org/get

# Test HTTPS
curl -x http://localhost:8888 https://httpbin.org/get

# Check your IP through proxy
curl -x http://localhost:8888 http://ip-api.com/json
```

## Q: Why are some HTTPS requests faster than others?

SmartProxy automatically uses direct connections for:
- Static files (images, CSS, JS, etc.)
- Known CDN domains
- Content that doesn't need filtering

This bypasses the upstream proxy for better performance.

## Q: How can I disable ad blocking?

Edit `config.yaml`:
```yaml
ad_blocking:
  enabled: false
```

## Q: Can I use a SOCKS5 proxy as upstream?

Yes! Configure it in `config.yaml`:
```yaml
upstream:
  proxy_url: "socks5://127.0.0.1:1080"
  username: "optional"
  password: "optional"
```

## Q: How do I change the listening port?

Edit `config.yaml`:
```yaml
server:
  http_port: 9999  # Change from default 8888
```
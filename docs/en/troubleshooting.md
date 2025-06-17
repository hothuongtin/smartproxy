# SmartProxy Troubleshooting Guide

## Frequently Asked Questions

### Authentication and Browser Behavior

#### Q: Why does curl show authentication error but Chrome still works?

**A:** Chrome uses persistent connections (HTTP keep-alive). When you enter wrong credentials:
- curl creates new connection → checks auth → shows error
- Chrome reuses existing authenticated connection → continues working

This is standard HTTP behavior, not a bug.

**Solution:**
1. Close Chrome completely to disconnect old connections
2. Or go to `chrome://net-internals/#sockets` → "Flush socket pools"
3. Reopen Chrome and enter new credentials

#### Q: How do I properly test authentication?

**With curl:**
```bash
# Test with correct credentials
correct_password=$(echo -n "proxy.example.com:8080" | base64)
curl -x http://http:${correct_password}@localhost:8888 http://ipinfo.io

# Force new connection
curl -x http://http:wrongpassword@localhost:8888 -H "Connection: close" http://httpbin.org/ip
```

**With browser:**
1. Close all browser windows
2. Clear browsing data (Ctrl+Shift+Delete)
3. Open with fresh profile:
   ```bash
   # macOS
   open -na "Google Chrome" --args --user-data-dir="/tmp/chrome_test" --proxy-server="http://localhost:8888"
   
   # Windows
   chrome.exe --user-data-dir="%TEMP%\chrome_test" --proxy-server="http://localhost:8888"
   ```

### Common Errors and Solutions

#### "illegal base64 data at input byte X"

**Cause:** Base64 encoding issues
- Wrapped with newlines (common at 76 characters)
- Extra whitespace
- Invalid characters

**Solution:** SmartProxy automatically handles wrapped base64. Create proper base64:
```bash
# Use -n to avoid trailing newline
echo -n "host:port:user:pass" | base64

# Verify your encoding
echo "YOUR_BASE64_STRING" | base64 -d
```

#### "LibreSSL: error:1404B42E:SSL routines:ST_CONNECT:tlsv1 alert protocol version"

**Cause:** Using `https://` in the proxy URL

**Solution:** Always use `http://` for the proxy URL:
```bash
# ✅ CORRECT - Always use http:// for proxy URL
curl -x http://http:PASSWORD@localhost:8888 https://example.com

# ❌ WRONG - Don't use https:// for proxy URL
curl -x https://http:PASSWORD@localhost:8888 https://example.com
```

#### "Authentication required" (407 Proxy Authentication Required)

**Cause:** No credentials provided

**Solution:** Provide username and password:
```bash
# Command line
curl -x http://http:BASE64_PASSWORD@localhost:8888 http://example.com

# Environment variables
export http_proxy=http://http:BASE64_PASSWORD@localhost:8888
export https_proxy=http://http:BASE64_PASSWORD@localhost:8888
```

#### "Invalid credentials" (403 Forbidden)

**Causes:**
- Wrong username (must be `http` or `socks5`)
- Invalid base64 encoding
- Upstream proxy details incorrect

**Solution:** Verify your credentials:
```bash
# Check username is correct
# Must be either "http" or "socks5"

# Verify base64 encoding
echo -n "proxy.example.com:8080" | base64
# Then decode to verify
echo "cHJveHkuZXhhbXBsZS5jb206ODA4MA==" | base64 -d
```

#### Port Already in Use

**Error:** `bind: address already in use`

**Solution:**
```bash
# Find process using port
lsof -i :8888  # macOS/Linux
netstat -ano | findstr :8888  # Windows

# Kill existing proxy
make kill
# Or
pkill -f smartproxy

# Start fresh
make run
```

### HTTPS and Certificate Issues

#### Q: Why does ip-api.com return "SSL unavailable" error?

**A:** The free ip-api.com API doesn't support HTTPS. Use HTTP instead:
- ❌ Wrong: `https://ip-api.com/json`
- ✅ Correct: `http://ip-api.com/json`

This is not a proxy issue - it's a limitation of their free tier.

#### Certificate Warnings in MITM Mode

**Cause:** CA certificate not properly installed or trusted

**Solution:**
1. Generate CA certificate:
   ```bash
   make ca-cert
   ```

2. Install CA certificate on client:
   - **macOS:** Double-click `certs/ca.crt` → Add to System keychain → Trust for SSL
   - **Windows:** Double-click `certs/ca.crt` → Install to "Trusted Root Certification Authorities"
   - **Linux:** `sudo cp certs/ca.crt /usr/local/share/ca-certificates/smartproxy.crt && sudo update-ca-certificates`

3. Restart browser after installation

#### Q: How can I access HTTPS sites with MITM disabled?

**A:** When `https_mitm: false` (default), SmartProxy tunnels HTTPS connections without decryption. This works perfectly for all HTTPS sites:
- Establishes tunnel using CONNECT method
- Routes through configured upstream proxy
- Maintains end-to-end encryption
- No certificate warnings

### Performance Issues

#### Slow Response Times

**Possible causes and solutions:**

1. **Upstream proxy is slow**
   - Test upstream proxy directly
   - Try different upstream proxy
   - Check network latency

2. **Connection pool exhausted**
   ```yaml
   server:
     max_idle_conns: 50000  # Increase for high load
     max_idle_conns_per_host: 500
   ```

3. **Too much logging**
   ```yaml
   logging:
     level: warn  # Reduce from debug/info
   ```

4. **System limits**
   ```bash
   # Check file descriptor limit
   ulimit -n
   
   # Increase limit
   ulimit -n 65536
   ```

#### High Memory Usage

**Solutions:**
1. Reduce connection pool size
2. Enable memory profiling:
   ```go
   import _ "net/http/pprof"
   // Access http://localhost:6060/debug/pprof/heap
   ```
3. Check for connection leaks in logs

### Debug Logging

#### Enable Debug Mode

**Method 1: Configuration file**
```yaml
logging:
  level: debug
```

**Method 2: Use debug config**
```bash
cp configs/config.debug.yaml configs/config.yaml
./smartproxy
```

**Method 3: Makefile**
```bash
make debug
```

#### What Debug Mode Shows

- Authentication details
- Routing decisions
- Connection handling
- Performance metrics
- Error details

#### Example Debug Output
```
15:04:10.234 DEBUG Incoming request method=GET url=http://example.com
15:04:10.235 DEBUG Parsing upstream from authentication username=http
15:04:10.236 DEBUG Authentication successful upstream_type=http
15:04:10.237 DEBUG URL identified as static file extension=.js
15:04:10.238 DEBUG Using direct connection reason=static_file
```

#### Filtering Debug Output

```bash
# Only authentication logs
./smartproxy 2>&1 | grep "auth"

# Only routing decisions
./smartproxy 2>&1 | grep -E "routing|direct|upstream"

# Errors only
./smartproxy 2>&1 | grep -E "ERROR|WARN"
```

### Specific Use Case Issues

#### Q: Why are some HTTPS requests faster than others?

SmartProxy automatically uses direct connections for:
- Static files (images, CSS, JS, etc.)
- Known CDN domains
- Content that doesn't need filtering

This bypasses the upstream proxy for better performance.

#### Q: How do I disable ad blocking?

Edit `configs/config.yaml`:
```yaml
ad_blocking:
  enabled: false
```

#### Q: Can I use a SOCKS5 proxy as upstream?

Yes! Use smart authentication:
```bash
# SOCKS5 without auth
echo -n "socks5.example.com:1080" | base64
# Use with username: socks5

# SOCKS5 with auth
echo -n "socks5.example.com:1080:user:pass" | base64
# Use with username: socks5
```

#### Q: How do I change the listening port?

Edit `configs/config.yaml`:
```yaml
server:
  http_port: 9999  # Change from default 8888
```

### Security Considerations

#### Debug Mode Security

Debug mode may log sensitive information:
- Decoded proxy credentials
- Request URLs
- Authentication headers

**Never use debug mode in production!**

#### Secure Setup Recommendations

1. **Bind to localhost only** (if not serving network)
2. **Use firewall rules** to restrict access
3. **Rotate credentials** regularly
4. **Monitor logs** for suspicious activity
5. **Keep CA private key secure** if using MITM

### Getting Help

If you're still having issues:

1. **Check logs** with debug mode enabled
2. **Run test scripts** in `scripts/test/`
3. **Search existing issues** on GitHub
4. **Create new issue** with:
   - SmartProxy version
   - Configuration (without sensitive data)
   - Debug logs
   - Steps to reproduce

### Useful Commands

```bash
# Check if proxy is running
ps aux | grep smartproxy

# Check listening ports
netstat -tlnp | grep 8888  # Linux
lsof -i :8888  # macOS

# Test basic connectivity
telnet localhost 8888

# Test with verbose curl
curl -v -x http://localhost:8888 http://httpbin.org/get

# Monitor real-time logs
tail -f smartproxy.log | grep -E "ERROR|WARN"
```
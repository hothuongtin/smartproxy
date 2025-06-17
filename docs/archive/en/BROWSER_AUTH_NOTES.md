# Browser Authentication Behavior

## Chrome vs curl Authentication Differences

When testing proxy authentication, you might notice different behavior between curl and Chrome.

### Issue
- curl shows authentication error with wrong password: `errorMsg: Account password authentication failed`
- Chrome continues to work normally with the same wrong password

### The Real Cause

**HTTP Persistent Connections (Keep-Alive)**
- Chrome uses persistent/keep-alive connections
- Authentication is only checked when establishing new connections
- Established connections are reused for multiple requests
- curl by default creates a new connection for each request

This is standard HTTP/1.1 and HTTP/2 behavior, not a bug.

### How to Handle

1. **Close all Chrome connections**
   - Close Chrome completely (all windows)
   - Or: chrome://net-internals/#sockets → "Close idle sockets"
   - Or: chrome://net-internals/#sockets → "Flush socket pools"

2. **Test with new connection**
   - Reopen Chrome
   - Enter new authentication credentials
   - Chrome will use new credentials for new connections

3. **Force new connection in curl**
   ```bash
   # Disable keep-alive
   curl -x http://http:wrongpassword@localhost:8888 \
        -H "Connection: close" \
        http://httpbin.org/ip
   ```

### Testing Authentication Properly

#### In Chrome
1. Close all Chrome windows completely
2. Clear browsing data:
   - Press `Ctrl+Shift+Delete` (Windows/Linux) or `Cmd+Shift+Delete` (Mac)
   - Select "All time"
   - Check "Cookies and other site data" and "Cached images and files"
   - Click "Clear data"
3. Open Chrome with fresh profile:
   ```bash
   # Windows
   chrome.exe --user-data-dir="%TEMP%\chrome_test" --proxy-server="http://localhost:8888"
   
   # macOS
   open -na "Google Chrome" --args --user-data-dir="/tmp/chrome_test" --proxy-server="http://localhost:8888"
   
   # Linux
   google-chrome --user-data-dir="/tmp/chrome_test" --proxy-server="http://localhost:8888"
   ```
4. Enter wrong credentials when prompted
5. Chrome should show authentication error

#### In curl
```bash
# Test with wrong password
curl -x http://http:wrongpassword@localhost:8888 http://httpbin.org/ip

# Test with correct format
correct_password=$(echo -n "proxy.example.com:8080" | base64)
curl -x http://http:${correct_password}@localhost:8888 http://httpbin.org/ip
```

### Debug Mode

Enable debug logging to see authentication details:
```bash
make debug
```

Look for these log entries:
```
DEBUG Incoming request method=CONNECT
DEBUG No authentication for CONNECT host=httpbin.org:443
DEBUG Failed to parse upstream from CONNECT auth error=...
```

### Security Implications

The strict authentication in curl is the correct behavior. Chrome's credential caching is a convenience feature but can mask authentication issues during testing.

For production use:
- Always test with fresh browser profiles
- Monitor proxy logs for authentication failures
- Consider implementing rate limiting for failed auth attempts
- Use strong, unique passwords for each upstream proxy
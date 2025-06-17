# SmartProxy Debug Logging

SmartProxy includes comprehensive debug logging to help troubleshoot issues and understand proxy behavior.

## Quick Start

### Enable Debug Mode

1. **Using Makefile (Recommended)**
```bash
make debug
```

2. **Using config file**
Edit `config.yaml`:
```yaml
logging:
  level: debug
```

3. **Using debug config**
```bash
cp config.debug.yaml config.yaml
./smartproxy
```

## What Debug Mode Logs

### Request Processing
- **Authentication**: Credential parsing, schema validation, base64 decoding
- **Routing**: Decision logic for direct vs upstream connections
- **Timing**: Request/response durations

### Example Debug Output

```
15:04:05.123 INFO Starting high-performance proxy server address=:8888 mode=smart_proxy_auth
15:04:05.124 DEBUG Configuration details http_port=8888 https_mitm=false max_idle_conns=10000
15:04:05.125 DEBUG Sample ad domains samples=[doubleclick.net, googleads.com] total=1000

# Incoming request
15:04:10.234 DEBUG Incoming request method=GET url=http://example.com host=example.com
15:04:10.235 DEBUG Parsing upstream from authentication username=http password_length=32
15:04:10.236 DEBUG Decoded upstream configuration decoded=proxy.example.com:8080 schema=http
15:04:10.237 DEBUG Authentication successful upstream_type=http upstream_host=proxy.example.com

# Routing decision
15:04:10.238 DEBUG Routing decision for request url=http://example.com/script.js
15:04:10.239 DEBUG URL identified as static file extension=.js action=direct_connection
15:04:10.240 DEBUG Using direct connection reason=static_file_or_cdn

# Response
15:04:10.345 DEBUG Direct request completed status=200 duration=105ms
15:04:10.346 DEBUG Response received status=200 content_type=application/javascript
```

### Logged Information

#### 1. **Server Startup**
- Configuration details
- Performance settings
- Loaded extensions and domains
- Ad blocking status

#### 2. **Authentication**
- Basic auth parsing
- Schema validation (http/socks5)
- Base64 decoding
- Upstream parsing

#### 3. **Routing Logic**
- Static file detection
- CDN domain matching
- Ad domain blocking
- Upstream selection

#### 4. **Connection Handling**
- Transport creation
- Connection caching
- SOCKS5/HTTP proxy setup
- Error details

#### 5. **Performance Metrics**
- Request/response timing
- Cache hits/misses
- Connection reuse

## Debug Logging Categories

### Static File Detection
```
DEBUG URL identified as static file url=http://cdn.com/app.js extension=.js
DEBUG URL not a static file url=http://api.com/data path=/data
# With MITM enabled, works for HTTPS too:
DEBUG URL identified as static file url=https://cdn.com/app.js extension=.js
```

### CDN Detection
```
DEBUG Domain identified as CDN host=cdn.example.com pattern=cdn.
DEBUG Domain not a CDN host=api.example.com
```

### Ad Blocking
```
DEBUG Domain blocked (exact match) host=doubleclick.net action=blocked
DEBUG Domain blocked (parent match) host=ads.google.com blocked_parent=google.com
```

### Upstream Proxy
```
DEBUG Creating new transport type=http host=proxy.com port=8080
DEBUG Using cached transport cache_key=http:proxy.com:8080
DEBUG Upstream request completed status=200 duration=250ms
```

### HTTPS with MITM
```
DEBUG MITM authentication check passed upstream=http://proxy.com:8080
DEBUG Intercepting HTTPS request method=GET url=https://example.com/script.js
DEBUG URL identified as static file extension=.js action=direct_connection
DEBUG Using direct connection for HTTPS static file url=https://cdn.com/app.js
DEBUG Using upstream proxy for HTTPS request url=https://api.com/data
```

## Performance Impact

Debug logging has minimal performance impact:
- Structured logging with slog
- Conditional debug checks
- Efficient string formatting
- No debug code in hot paths

## Troubleshooting

### Common Issues

1. **No debug output**
   - Verify `logging.level: debug` in config
   - Check config file path
   - Ensure logger initialization

2. **Too much output**
   - Filter by component
   - Use grep for specific patterns
   - Adjust log level to info/warn

3. **Missing information**
   - Check if feature has debug logging
   - Submit issue for additional logging

### Filtering Debug Output

```bash
# Only authentication logs
./smartproxy 2>&1 | grep "auth"

# Only routing decisions
./smartproxy 2>&1 | grep -E "routing|direct|upstream"

# Errors only
./smartproxy 2>&1 | grep -E "ERROR|WARN"
```

## Security Considerations

Debug mode may log sensitive information:
- Decoded proxy credentials
- Request URLs
- Authentication headers

**Never use debug mode in production!**

## Contributing

To add debug logging:

1. Use the global logger
```go
logger.Debug("Operation description",
    "key1", value1,
    "key2", value2)
```

2. Check debug level for expensive operations
```go
if logger.Enabled(nil, slog.LevelDebug) {
    // Expensive debug logic
}
```

3. Follow naming conventions
- Use descriptive messages
- Include relevant context
- Avoid logging sensitive data
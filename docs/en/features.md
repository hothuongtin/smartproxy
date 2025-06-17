# SmartProxy Features

## Overview

SmartProxy is designed to be an intelligent HTTP/HTTPS proxy that optimizes web traffic through smart routing decisions, ad blocking, and high-performance connection handling.

## Core Features

### 1. Smart Routing Logic

SmartProxy automatically determines the best path for each request:

#### Static File Detection
- **What**: Automatically detects static files by extension
- **How**: O(1) lookup of file extensions from URL path
- **Files**: `.js`, `.css`, `.jpg`, `.png`, `.gif`, `.pdf`, etc.
- **Benefit**: Direct connections for faster loading
- **Example**:
  ```
  http://example.com/app.js?v=123 → Direct connection
  http://example.com/api/data → Upstream proxy
  ```

#### CDN Domain Detection
- **What**: Identifies CDN domains for direct routing
- **How**: Pattern matching on domain names
- **Patterns**: `cdn.`, `static.`, `assets.`, major CDN providers
- **Benefit**: Bypasses proxy for already-optimized content
- **Example**:
  ```
  https://cdn.example.com → Direct connection
  https://api.example.com → Upstream proxy
  ```

#### Intelligent Routing Decision Flow
```
Request → Is it an ad? → Block (204 No Content)
         ↓ No
         Is it static file? → Direct connection
         ↓ No
         Is it CDN domain? → Direct connection
         ↓ No
         Route through upstream proxy
```

### 2. Ad Blocking

#### High-Performance Blocking
- **O(1) Lookup**: Hash map for instant domain matching
- **Hierarchical Blocking**: Blocks subdomains automatically
- **Thread-Safe**: Concurrent access with RWMutex
- **Example**:
  ```yaml
  # Blocking "doubleclick.net" also blocks:
  - ads.doubleclick.net
  - static.doubleclick.net
  - any.subdomain.doubleclick.net
  ```

#### Implementation Details
- Loads domains from `ad_domains.yaml`
- Returns 204 No Content for blocked domains
- Works on HTTP and HTTPS (with MITM enabled)
- Zero memory allocation for lookups

### 3. Connection Pooling

#### Performance Benefits
- **Reuse Connections**: Reduces TCP handshake overhead
- **Configurable Limits**: Tune for your workload
- **Per-Host Limits**: Prevent overwhelming single servers

#### Default Configuration
```yaml
max_idle_conns: 10000        # Total pool size
max_idle_conns_per_host: 100 # Per-host limit
idle_conn_timeout: 90         # Seconds before closing
```

#### Connection Types
1. **Direct Transport**: For static files and CDNs
2. **HTTP Proxy Transport**: For HTTP upstream proxies
3. **SOCKS5 Transport**: For SOCKS5 upstream proxies

### 4. HTTPS Handling

#### Tunneling Mode (Default)
- **How It Works**: Creates encrypted tunnel without inspection
- **Benefits**:
  - No certificate warnings
  - True end-to-end encryption
  - Zero configuration
- **Limitations**:
  - Cannot block ads on HTTPS
  - Cannot detect static files on HTTPS

#### MITM Mode (Advanced)
- **How It Works**: Decrypts, inspects, and re-encrypts
- **Benefits**:
  - Ad blocking on HTTPS sites
  - Static file detection for HTTPS
  - Full routing intelligence
- **Requirements**:
  - CA certificate installation
  - Authentication for all requests

### 5. Smart Authentication

#### Dynamic Upstream Configuration
Instead of static upstream configuration, SmartProxy uses authentication credentials to dynamically configure upstream proxies per connection.

**Format**:
```
Username: <schema>  # http or socks5
Password: <base64-encoded-upstream>
```

**Examples**:
```bash
# HTTP proxy without auth
Username: http
Password: cHJveHkuZXhhbXBsZS5jb206ODA4MA== # proxy.example.com:8080

# SOCKS5 with auth
Username: socks5  
Password: c29ja3MuZXhhbXBsZS5jb206MTA4MDp1c2VyOnBhc3M= # socks.example.com:1080:user:pass
```

#### Benefits
- Multiple upstream proxies without restart
- Per-client proxy configuration
- Easy proxy rotation
- No config file changes needed

### 6. HTTP/2 Support

- **Multiplexing**: Multiple requests over single connection
- **Header Compression**: Reduced overhead
- **Server Push**: Not used but supported
- **Automatic**: Negotiated via ALPN

### 7. Performance Optimizations

#### Zero-Copy Operations
- Efficient `io.Copy` for data transfer
- Minimal memory allocations
- Reused buffers where possible

#### Optimized Buffers
```yaml
read_buffer_size: 65536   # 64KB default
write_buffer_size: 65536  # 64KB default
```

#### Concurrent Request Handling
- Go routines for each request
- Non-blocking I/O
- Graceful connection management

### 8. Comprehensive Logging

#### Structured Logging
- Uses `slog` for structured output
- Colored output for readability
- JSON format option available

#### Log Levels
- **Debug**: Detailed routing decisions
- **Info**: Request/response summaries  
- **Warn**: Potential issues
- **Error**: Failures and errors

#### Example Output
```
12:34:56.789 INFO Incoming request method=GET url=http://example.com client=127.0.0.1:54321
12:34:56.790 DEBUG URL identified as static file extension=.js action=direct_connection
12:34:56.891 INFO Request completed status=200 duration=101ms
```

### 9. Graceful Shutdown

- Handles SIGINT/SIGTERM signals
- Waits for active connections
- Closes transports properly
- Prevents data loss

### 10. Cross-Platform Support

#### Supported Platforms
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)
- FreeBSD (amd64)

#### Docker Support
- Distroless images (~15MB)
- Scratch images (~12MB)
- Multi-arch builds

## Feature Comparison

| Feature | SmartProxy | Traditional Proxy |
|---------|------------|------------------|
| Static file detection | ✅ Automatic | ❌ All through proxy |
| CDN routing | ✅ Automatic | ❌ All through proxy |
| Ad blocking | ✅ O(1) performance | ⚠️ Varies |
| Connection pooling | ✅ 10,000+ connections | ⚠️ Limited |
| Dynamic upstream | ✅ Per-connection | ❌ Static config |
| HTTPS options | ✅ Tunnel or MITM | ⚠️ Usually one |
| Memory usage | ✅ ~50MB under load | ⚠️ Often higher |

## Use Cases

### 1. Development Environment
- Route API calls through proxy
- Direct connection for assets
- Debug with MITM mode

### 2. Ad-Free Browsing
- Block ads at proxy level
- Works for all devices
- No browser extensions needed

### 3. Multi-Proxy Setup
- Different proxies per client
- Easy proxy rotation
- No configuration changes

### 4. Performance Optimization
- Direct CDN connections
- Cached transport connections
- Reduced latency

## Future Features

Planned enhancements:
- Request/response modification
- Custom routing rules
- Metrics and monitoring
- WebSocket support
- HTTP/3 support
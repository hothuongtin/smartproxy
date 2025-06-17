# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Build and Run
```bash
# Build the proxy
make build

# Run the proxy  
make run

# Build for all platforms
make build-all

# Clean build artifacts
make clean

# Download dependencies
make deps
```

### Testing
```bash
# Run functional tests
./test_proxy.sh          # Main test suite
./test_https.sh          # HTTPS functionality
./test_http_js.sh        # HTTP JavaScript handling
./test_js_direct.sh      # Direct connection tests

# Generate CA certificate for HTTPS MITM
./generate_ca.sh
```

### Docker
```bash
# Build Docker image
make docker-build

# Run Docker container
make docker-run

# Use docker-compose
docker-compose up -d

# Build minimal scratch image
docker build -f Dockerfile.scratch -t smartproxy:scratch .
```

### Development Workflow
```bash
# Kill running proxy
make kill

# Run in development mode
make dev

# Format code
make fmt

# Run linter
make lint

# Test with curl
curl -x http://localhost:8888 http://httpbin.org/get
```

## Architecture

### Core Design
SmartProxy is a high-performance HTTP/HTTPS proxy with intelligent routing. All business logic resides in `main.go` with configuration handling in `config.go`.

### Request Flow
1. Client connects to proxy (default port 8888)
2. Proxy determines request type:
   - **Static files** (.js, .css, images, etc.) → Direct connection
   - **CDN domains** → Direct connection  
   - **Ad domains** → Blocked (204 No Content)
   - **Other requests** → Upstream proxy (REQUIRED)

### Key Components

**Transport Management**
- `createOptimizedTransport()` - Creates high-performance direct transport with connection pooling
- `createHTTPProxyTransport()` - Creates transport for HTTP upstream proxy
- `createSOCKS5ProxyTransport()` - Creates transport for SOCKS5 upstream proxy

**Routing Logic**
- `isStaticFile()` - Checks if URL matches static file extensions (handles query params correctly)
- `isCDNDomain()` - Checks if domain is a known CDN
- `isAdDomain()` - O(1) lookup in ad domains map

**HTTPS Handling**
- **MITM Disabled** (default): Tunnels HTTPS without decryption
- **MITM Enabled**: Decrypts HTTPS for inspection (requires CA cert)

### Configuration

**YAML Structure**
- `config.yaml` - Main configuration
  - Server settings (port, MITM, performance)
  - Upstream proxy (MANDATORY)
  - Direct routing patterns
- `ad_domains.yaml` - Ad blocking domains

**Important Settings**
- `upstream.proxy_url` - REQUIRED, no fallback to direct
- `server.https_mitm` - Controls HTTPS interception
- `server.max_idle_conns` - Connection pool size (default: 10000)

### Performance Optimizations
- Connection pooling with configurable limits
- O(1) ad domain lookups using hash maps
- Direct connections for static content
- HTTP/2 support
- Optimized buffer sizes (64KB)
- Graceful shutdown handling

### Concurrency Model
- Go's built-in goroutines handle each request
- `sync.RWMutex` protects ad domains map
- Transport instances are shared and thread-safe
- No explicit worker pool needed

### Error Handling
- Upstream proxy required - fails fast if not configured
- HTTPS tunneling errors logged but don't break connections
- CA certificate errors provide clear instructions

## Important Notes

1. **Upstream Proxy Required**: The proxy will not start without `upstream.proxy_url` configured
2. **Static File Detection**: Uses `url.Parse()` which automatically strips query parameters and fragments
3. **Ad Blocking**: Only works on HTTP or when HTTPS MITM is enabled
4. **Environment Variables**: `SMARTPROXY_CONFIG` overrides config file path
5. **Port Conflicts**: Default port 8888, kill existing processes with `pkill -f smartproxy`
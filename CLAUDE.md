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
./scripts/test/test_proxy.sh          # Main test suite
./scripts/test/test_https.sh          # HTTPS functionality
./scripts/test/test_http_js.sh        # HTTP JavaScript handling
./scripts/test/test_js_direct.sh      # Direct connection tests

# Generate CA certificate for HTTPS MITM
./scripts/setup/generate_ca.sh
```

### Docker
```bash
# Build Docker image
make docker-build

# Run Docker container
make docker-run

# Use docker-compose
cd docker && docker-compose up -d

# Build minimal scratch image
docker build -f docker/Dockerfile.scratch -t smartproxy:scratch .
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
SmartProxy is a high-performance HTTP/HTTPS proxy with intelligent routing using modular architecture. The project follows standard Go layout with separate packages for different concerns.

### Module Structure
- **cmd/smartproxy/** - Main application entry point
- **internal/config/** - Configuration management and validation
- **internal/logger/** - Structured logging with slogcolor
- **internal/proxy/** - Core proxy logic, routing, and transport management

### Request Flow
1. Client connects to proxy (default port 8888)
2. Proxy authenticates client and extracts upstream from credentials
3. Proxy determines request type:
   - **Static files** (.js, .css, images, etc.) → Direct connection
   - **CDN domains** → Direct connection  
   - **Ad domains** → Blocked (204 No Content)
   - **Other requests** → Upstream proxy (from authentication)

### Key Components

**Server Management** (`internal/proxy/server.go`)
- `NewServer()` - Creates configured proxy server with all components
- `Start()` - Starts HTTP proxy server with graceful shutdown support
- Authentication and upstream routing handlers

**Transport Management** (`internal/proxy/transport.go`)
- `CreateOptimizedTransport()` - Creates high-performance direct transport with connection pooling
- `CreateHTTPProxyTransport()` - Creates transport for HTTP upstream proxy
- `CreateSOCKS5ProxyTransport()` - Creates transport for SOCKS5 upstream proxy
- `DialThroughHTTPProxy()` - Direct HTTPS upstream tunneling for HTTP proxies
- `DialThroughSOCKS5Proxy()` - Direct HTTPS upstream tunneling for SOCKS5 proxies

**Routing Logic** (`internal/proxy/routing.go`)
- `IsStaticFile()` - Checks if URL matches static file extensions (handles query params correctly)
- `IsCDNDomain()` - Checks if domain is a known CDN
- `IsAdDomain()` - O(1) lookup in ad domains map

**HTTPS Handling**
- **MITM Disabled** (default): Tunnels HTTPS without decryption through upstream proxies
- **MITM Enabled**: Decrypts HTTPS for inspection (requires CA cert)
- **ConnectDial Handler**: Routes HTTPS CONNECT requests through configured upstream proxies

### Configuration

**YAML Structure**
- `configs/config.yaml` - Main configuration
  - Server settings (port, MITM, performance)
  - Direct routing patterns
  - Logging settings
- `configs/ad_domains.yaml` - Ad blocking domains

**Important Settings**
- `server.https_mitm` - Controls HTTPS interception
- `server.max_idle_conns` - Connection pool size (default: 10000)
- `logging.level` - Log verbosity (debug, info, warn, error)

**Smart Authentication**
- Username: Schema (http or socks5)
- Password: Base64(host:port[:user:pass])

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
- Authentication required - returns 407 if no credentials provided
- Invalid credentials - returns 403 with clear error
- HTTPS tunneling errors logged but don't break connections
- CA certificate errors provide clear instructions

## Important Notes

1. **Smart Proxy Authentication**: Upstream proxy configured dynamically via authentication credentials, no config needed
2. **Static File Detection**: Uses `url.Parse()` which automatically strips query parameters and fragments
3. **Ad Blocking**: Only works on HTTP or when HTTPS MITM is enabled
4. **MITM Mode**: Now requires authentication for all requests, ensuring secure proxy usage
5. **Static File Detection**: Works seamlessly with MITM mode enabled, properly routing static files directly
6. **Environment Variables**: `SMARTPROXY_CONFIG` overrides config file path
7. **Port Conflicts**: Default port 8888, kill existing processes with `pkill -f smartproxy`

## Documentation Updates

When making changes to the codebase, remember to update the relevant documentation:

### Documentation Structure
```
docs/
├── en/                        # English documentation
│   ├── getting-started.md     # Installation and setup
│   ├── configuration.md       # Configuration reference
│   ├── features.md           # Core features explained
│   ├── authentication.md     # Smart proxy authentication
│   ├── development.md        # Development guide
│   ├── troubleshooting.md    # FAQ and debugging
│   └── performance.md        # Performance optimization
├── vi/                       # Vietnamese documentation
│   └── (same structure as en/)
└── examples/                 # Example configurations
    ├── config.example.yaml
    └── config.debug.yaml
```

### What to Update
- **Configuration Changes**: Update `configuration.md` and example configs
- **New Features**: Update `features.md` and relevant guides
- **API/Auth Changes**: Update `authentication.md`
- **Performance**: Update `performance.md`
- **Common Issues**: Update `troubleshooting.md`
- **Build/Dev Process**: Update `development.md`

### Key Files
- `README.md` / `README_vi.md` - Keep overview current, link to docs/
- `configs/config.example.yaml` - Example configuration
- `docker/docker-compose.yml` - Docker Compose config
- `Makefile` - Build automation

**Always keep documentation in sync with code changes!**

## Memories
- Update related documents each time they are edited.
- Do not create a socks5 server for this project, the socks5 server only declares upstream
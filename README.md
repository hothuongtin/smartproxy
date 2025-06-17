# SmartProxy

A high-performance HTTP/HTTPS proxy with intelligent routing, ad blocking, and minimal resource usage.

## Features

- 🚀 **High Performance**: Handles thousands of concurrent connections with connection pooling
- 🎯 **Intelligent Routing**: Direct connections for static files and CDNs
- 🚫 **Ad Blocking**: Block ads and tracking domains with O(1) lookup performance
- 🔒 **HTTPS Support**: Optional MITM for inspection or secure tunneling
- 🌈 **Colored Logging**: Beautiful structured logs with slogcolor
- 📦 **Minimal Docker Images**: ~15MB production images using distroless/scratch
- 🔧 **Flexible Configuration**: YAML-based configuration with hot-reload support

## Quick Start

### Using Make

```bash
# Build and run
make build
make run

# Or in one command
make dev
```

### Using Docker

```bash
# Using docker-compose (recommended)
docker-compose up -d

# Or build and run manually
make docker-build
make docker-run
```

### Configuration

1. Copy the example config:
```bash
cp config.example.yaml config.yaml
```

2. Configure your upstream proxy (REQUIRED):
```yaml
upstream:
  proxy_url: "http://your-proxy:8080"
  username: "optional"
  password: "optional"
```

3. Run the proxy:
```bash
make run
```

## Configuration Options

### Basic Settings

```yaml
server:
  http_port: 8888              # Proxy listen port
  https_mitm: false            # Enable HTTPS interception
  max_idle_conns: 10000        # Connection pool size
  max_idle_conns_per_host: 100 # Per-host connection limit
```

### Upstream Proxy (Required)

```yaml
upstream:
  proxy_url: "http://proxy:8080"  # or "socks5://127.0.0.1:1080"
  username: ""
  password: ""
```

### Ad Blocking

```yaml
ad_blocking:
  enabled: true
  domains_file: "ad_domains.yaml"
```

## Performance

SmartProxy is optimized for high-performance operation:

- **Connection Pooling**: Reuses connections to reduce overhead
- **O(1) Ad Blocking**: Hash map lookups for instant domain matching
- **Direct Routing**: Bypasses upstream proxy for static content
- **HTTP/2 Support**: Multiplexing for better performance
- **Zero-Copy Operations**: Minimal memory allocations

### Benchmarks

With default settings:
- 10,000+ concurrent connections
- 5,000+ requests/second
- Sub-millisecond overhead for direct connections
- ~50MB memory usage under load

## Docker Images

We provide multiple Docker image options:

### Distroless (Recommended)
- Size: ~15MB
- Security: No shell, minimal attack surface
- Base: `gcr.io/distroless/static-debian12`

```bash
docker build -t smartproxy:latest .
```

### Scratch (Minimal)
- Size: ~12MB
- Security: Absolutely minimal
- Base: `scratch`

```bash
docker build -f Dockerfile.scratch -t smartproxy:scratch .
```

## HTTPS Configuration

### Tunneling Mode (Default)
- No certificate warnings
- End-to-end encryption maintained
- Zero configuration required

### MITM Mode
For HTTPS inspection:

1. Generate CA certificate:
```bash
make ca-cert
```

2. Enable in config:
```yaml
server:
  https_mitm: true
  ca_cert: "certs/ca.crt"
  ca_key: "certs/ca.key"
```

3. Install CA on client devices

## Development

### Prerequisites
- Go 1.21+
- Make
- Docker (optional)

### Building

```bash
# Local development
make dev

# Production build
make build

# Cross-platform builds
make build-all
```

### Testing

```bash
# Run all tests
make test

# Test specific functionality
./test_proxy.sh
./test_https.sh
```

### Code Quality

```bash
# Format code
make fmt

# Run linter
make lint
```

## Architecture

SmartProxy uses a simple yet effective architecture:

- **Single Binary**: All functionality in one executable
- **YAML Configuration**: Easy to manage settings
- **Pluggable Transports**: Support for HTTP/SOCKS5 upstreams
- **Graceful Shutdown**: Proper connection cleanup

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

MIT License - see LICENSE file for details

## Troubleshooting

### Port Already in Use
```bash
make kill  # Kill existing proxy
make run   # Start fresh
```

### Certificate Errors
- Ensure CA certificate is installed on client
- Check certificate validity dates
- Verify MITM is enabled in config

### Performance Issues
- Increase `max_idle_conns` for more connections
- Check upstream proxy performance
- Monitor system resources

## Support

- Issues: [GitHub Issues](https://github.com/yourusername/smartproxy/issues)
- Documentation: See `docs/` directory
- FAQ: See `FAQ.md`
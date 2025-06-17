# SmartProxy

A high-performance HTTP/HTTPS proxy with intelligent routing, ad blocking, and minimal resource usage.

## Features

- 🚀 **High Performance**: Handles thousands of concurrent connections with connection pooling
- 🎯 **Intelligent Routing**: Direct connections for static files and CDNs
- 🚫 **Ad Blocking**: Block ads and tracking domains with O(1) lookup performance
- 🔒 **HTTPS Support**: Optional MITM for inspection or secure tunneling with full authentication
- 🌈 **Colored Logging**: Beautiful structured logs with slogcolor
- 📦 **Minimal Docker Images**: ~15MB production images using distroless/scratch
- 🔧 **Flexible Configuration**: YAML-based configuration with hot-reload support
- 🔐 **Smart Authentication**: Dynamic upstream proxy configuration via authentication credentials

## Quick Start

### Download Pre-built Binaries

Download the latest release for your platform from the [Releases page](https://github.com/hothuongtin/smartproxy/releases).

```bash
# Example for Linux AMD64
wget https://github.com/hothuongtin/smartproxy/releases/latest/download/smartproxy-linux-amd64.tar.gz
tar xzf smartproxy-linux-amd64.tar.gz
./smartproxy
```

### Build from Source

For detailed installation and setup instructions, see [Getting Started Guide](docs/en/getting-started.md).

```bash
# Quick install
make build
make run

# Or with Docker
docker-compose up -d
```

## Documentation

📚 **Complete documentation is available in the [`docs/`](docs/) directory:**

### English Documentation
- [Getting Started](docs/en/getting-started.md) - Installation and basic setup
- [Configuration Guide](docs/en/configuration.md) - All configuration options
- [Features](docs/en/features.md) - Core features explained
- [Authentication](docs/en/authentication.md) - Smart proxy authentication
- [Development](docs/en/development.md) - Development guide
- [Troubleshooting](docs/en/troubleshooting.md) - FAQ and debugging
- [Performance](docs/en/performance.md) - Performance optimization

### Vietnamese Documentation
- [Bắt đầu](docs/vi/getting-started_vi.md) - Hướng dẫn cài đặt
- [Cấu hình](docs/vi/configuration_vi.md) - Tùy chọn cấu hình
- [Tính năng](docs/vi/features_vi.md) - Giải thích các tính năng
- [Xác thực](docs/vi/authentication_vi.md) - Xác thực proxy thông minh
- [Phát triển](docs/vi/development_vi.md) - Hướng dẫn phát triển
- [Khắc phục sự cố](docs/vi/troubleshooting_vi.md) - FAQ và gỡ lỗi
- [Hiệu suất](docs/vi/performance_vi.md) - Tối ưu hiệu suất

## Key Features Overview

### 🔐 Smart Authentication
Configure upstream proxies dynamically via authentication:
```bash
# HTTP proxy
curl -x http://http:$(echo -n "proxy.com:8080" | base64)@localhost:8888 http://ipinfo.io

# SOCKS5 with auth
curl -x http://socks5:$(echo -n "socks.com:1080:user:pass" | base64)@localhost:8888 http://ipinfo.io
```

### 🚀 Performance
- 10,000+ concurrent connections
- 5,000+ requests/second
- Sub-millisecond routing decisions
- ~50MB memory usage

### 🎯 Intelligent Routing
- Static files → Direct connection
- CDN domains → Direct connection  
- Ad domains → Blocked
- Other requests → Upstream proxy

## Contributing

See [Development Guide](docs/en/development.md) for contribution guidelines.

## License

MIT License - see LICENSE file for details

## Support

- Issues: [GitHub Issues](https://github.com/hothuongtin/smartproxy/issues)
- Documentation: See [`docs/`](docs/) directory
- Community: Discussions and questions welcome!
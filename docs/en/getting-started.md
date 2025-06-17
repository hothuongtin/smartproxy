# Getting Started with SmartProxy

## Overview

SmartProxy is a high-performance HTTP/HTTPS proxy with intelligent routing, ad blocking, and minimal resource usage. This guide will help you get up and running quickly.

## System Requirements

- **Operating System**: Linux, macOS, or Windows
- **Go**: 1.21+ (for building from source)
- **Docker**: Optional, for containerized deployment
- **Memory**: Minimum 64MB, recommended 256MB+
- **CPU**: Any modern CPU

## Installation Methods

### Method 1: Using Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/hothuongtin/smartproxy/releases).

```bash
# Download and extract (example for Linux)
wget https://github.com/hothuongtin/smartproxy/releases/download/vX.X.X/smartproxy-linux-amd64.tar.gz
tar -xzf smartproxy-linux-amd64.tar.gz
chmod +x smartproxy

# Run
./smartproxy
```

### Method 2: Using Docker (Recommended)

```bash
# Using docker-compose (recommended)
cd docker && docker-compose up -d

# Or build and run manually
docker build -t smartproxy:latest .
docker run -d \
  --name smartproxy \
  -p 8888:8888 \
  -v $(pwd)/configs/config.yaml:/app/config.yaml:ro \
  smartproxy:latest
```

### Method 3: Building from Source

```bash
# Clone the repository
git clone https://github.com/hothuongtin/smartproxy.git
cd smartproxy

# Build using Make
make build

# Run
make run

# Or in one command for development
make dev
```

## Quick Start Guide

### Step 1: Configuration

Copy the example configuration file:

```bash
cp configs/config.example.yaml configs/config.yaml
```

Basic configuration (`configs/config.yaml`):

```yaml
server:
  http_port: 8888              # Proxy listen port
  https_mitm: false            # Enable HTTPS interception (requires CA cert)
  max_idle_conns: 10000        # Connection pool size
  max_idle_conns_per_host: 100 # Per-host connection limit

ad_blocking:
  enabled: true
  domains_file: "configs/ad_domains.yaml"

direct_patterns:
  static_files:
    - .js
    - .css
    - .jpg
    - .png
    - .gif
    - .ico
    - .woff
    - .woff2
  
  cdn_domains:
    - cloudflare.com
    - cdn.jsdelivr.net
    - cdnjs.cloudflare.com
    - unpkg.com

logging:
  level: info  # debug, info, warn, error
  colored: true
```

### Step 2: Configure Proxy Authentication

SmartProxy uses smart authentication to dynamically configure upstream proxies:

```bash
# Format
Username: <schema>  # http or socks5
Password: <base64-encoded-upstream>

# Example: HTTP proxy without auth
echo -n "proxy.example.com:8080" | base64
# Result: cHJveHkuZXhhbXBsZS5jb206ODA4MA==

# Example: SOCKS5 proxy with auth
echo -n "socks.example.com:1080:user:pass" | base64
# Result: c29ja3MuZXhhbXBsZS5jb206MTA4MDp1c2VyOnBhc3M=
```

### Step 3: Run SmartProxy

```bash
# Using Make
make run

# Or directly
./smartproxy

# With custom config path
SMARTPROXY_CONFIG=/path/to/config.yaml ./smartproxy
```

### Step 4: Configure Your Client

#### Browser Configuration

Configure your browser to use `http://localhost:8888` as the HTTP/HTTPS proxy.

Example with authentication:
```
Proxy: http://localhost:8888
Username: http
Password: cHJveHkuZXhhbXBsZS5jb206ODA4MA==
```

#### Command Line

```bash
# Using curl
curl -x http://http:cHJveHkuZXhhbXBsZS5jb206ODA4MA==@localhost:8888 http://ipinfo.io

# Using environment variables
export http_proxy=http://http:cHJveHkuZXhhbXBsZS5jb206ODA4MA==@localhost:8888
export https_proxy=http://http:cHJveHkuZXhhbXBsZS5jb206ODA4MA==@localhost:8888
curl http://ipinfo.io
```

## HTTPS Configuration

### Tunneling Mode (Default)

By default, SmartProxy tunnels HTTPS traffic without decryption:

- No certificate warnings
- End-to-end encryption maintained
- Zero configuration required

### MITM Mode (Advanced)

To inspect HTTPS traffic and enable features like ad blocking on HTTPS sites:

#### Step 1: Generate CA Certificate

```bash
make ca-cert
# Or manually
./scripts/generate_ca.sh
```

This creates:
- `certs/ca.crt` - CA certificate (install on clients)
- `certs/ca.key` - Private key (keep secure)

#### Step 2: Enable MITM in Configuration

```yaml
server:
  https_mitm: true
  ca_cert: "certs/ca.crt"
  ca_key: "certs/ca.key"
```

#### Step 3: Install CA Certificate on Clients

**macOS:**
1. Double-click `certs/ca.crt`
2. Add to System keychain
3. Trust for SSL

**Windows:**
1. Double-click `certs/ca.crt`
2. Install to "Trusted Root Certification Authorities"

**Linux:**
```bash
sudo cp certs/ca.crt /usr/local/share/ca-certificates/smartproxy.crt
sudo update-ca-certificates
```

**iOS:**
1. Email or AirDrop `ca.crt` to device
2. Install profile
3. Go to Settings > General > About > Certificate Trust Settings
4. Enable for SmartProxy CA

**Android:**
1. Copy `ca.crt` to device
2. Settings > Security > Install from storage
3. Choose "CA certificate"

## Docker Deployment

### Using Docker Compose (Recommended)

```yaml
version: '3.8'

services:
  smartproxy:
    image: smartproxy:latest
    container_name: smartproxy
    restart: unless-stopped
    ports:
      - "8888:8888"
    volumes:
      - ./configs/config.yaml:/app/config.yaml:ro
      - ./configs/ad_domains.yaml:/app/ad_domains.yaml:ro
      - ./certs:/app/certs:ro  # Only if using MITM
    environment:
      - SMARTPROXY_CONFIG=/app/config.yaml
    security_opt:
      - no-new-privileges:true
    read_only: true
    tmpfs:
      - /tmp
```

### Docker Image Options

**Distroless (Recommended):**
- Size: ~15MB
- Security: No shell, minimal attack surface
- Base: `gcr.io/distroless/static-debian12`

```bash
docker build -f docker/Dockerfile -t smartproxy:latest .
```

**Scratch (Minimal):**
- Size: ~12MB
- Security: Absolutely minimal
- Base: `scratch`

```bash
docker build -f docker/Dockerfile.scratch -t smartproxy:scratch .
```

## Verifying Installation

Test your proxy setup:

```bash
# Test HTTP
curl -x http://localhost:8888 http://httpbin.org/get

# Test HTTPS
curl -x http://localhost:8888 https://httpbin.org/get

# Test with authentication
curl -x http://http:cHJveHkuZXhhbXBsZS5jb206ODA4MA==@localhost:8888 http://ipinfo.io
```

## Common Issues

### Port Already in Use

```bash
# Kill existing proxy
make kill  # or pkill -f smartproxy

# Start fresh
make run
```

### Authentication Errors

Always use `http://` for the proxy URL, even when accessing HTTPS sites:

```bash
# ✅ CORRECT
curl -x http://http:PASSWORD@localhost:8888 https://example.com

# ❌ WRONG - Don't use https:// for proxy URL
curl -x https://http:PASSWORD@localhost:8888 https://example.com
```

### Certificate Errors in MITM Mode

- Ensure CA certificate is properly installed and trusted
- Check certificate validity dates
- Verify MITM is enabled in config

## Next Steps

- [Configuration Guide](configuration.md) - Detailed configuration options
- [Features](features.md) - Learn about intelligent routing and ad blocking
- [Authentication](authentication.md) - Advanced authentication setup
- [Troubleshooting](troubleshooting.md) - Common issues and solutions
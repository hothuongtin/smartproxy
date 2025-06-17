# SmartProxy Configuration Guide

## Overview

SmartProxy uses YAML configuration files to control its behavior. Configuration can be specified through:

1. **Configuration file**: Default `configs/config.yaml`
2. **Environment variable**: `SMARTPROXY_CONFIG=/path/to/config.yaml`
3. **Smart authentication**: Dynamic upstream configuration via proxy credentials

## Configuration File Format

The configuration file uses YAML format with the following main sections:

- `server` - Proxy server settings
- `ad_blocking` - Ad blocking configuration
- `direct_extensions` - File extensions for direct routing
- `direct_domains` - Domains for direct routing
- `logging` - Logging configuration

## Server Settings

```yaml
server:
  # HTTP/HTTPS proxy port
  http_port: 8888
  
  # HTTPS interception settings
  https_mitm: false    # Enable/disable HTTPS interception (MITM)
  ca_cert: ""          # Path to CA certificate file (auto-generated if empty)
  ca_key: ""           # Path to CA private key file (auto-generated if empty)
  
  # Performance settings for high concurrency
  max_idle_conns: 10000        # Maximum idle connections
  max_idle_conns_per_host: 100 # Maximum idle connections per host
  idle_conn_timeout: 90         # Idle connection timeout in seconds
  tls_handshake_timeout: 10     # TLS handshake timeout in seconds
  expect_continue_timeout: 1    # Expect continue timeout in seconds
  
  # Buffer settings
  read_buffer_size: 65536       # Read buffer size (64KB)
  write_buffer_size: 65536      # Write buffer size (64KB)
```

### Key Server Settings Explained

- **`http_port`**: The port SmartProxy listens on (default: 8888)
- **`https_mitm`**: When `true`, decrypts HTTPS traffic for inspection. Requires CA certificate.
- **`max_idle_conns`**: Total connection pool size. Higher values improve performance but use more memory.
- **`max_idle_conns_per_host`**: Per-host connection limit to prevent overwhelming single servers.

## Smart Authentication Mode

SmartProxy dynamically configures upstream proxies through authentication credentials:

```yaml
# No upstream configuration needed in config file!
# Upstream is configured per-connection via authentication

# Authentication format:
# Username: schema (http or socks5)
# Password: base64 encoded upstream details

# Examples:
# - Without upstream auth: base64("host:port")
#   Username: http
#   Password: bmEubHVuYXByb3h5LmNvbToxMjIzMw== (na.lunaproxy.com:12233)
#
# - With upstream auth: base64("host:port:username:password")
#   Username: socks5
#   Password: bmEubHVuYXByb3h5LmNvbToxMjIzMzp1c2VyOnBhc3M= (na.lunaproxy.com:12233:user:pass)
```

## Ad Blocking Configuration

```yaml
ad_blocking:
  enabled: true
  domains_file: "configs/ad_domains.yaml"
```

The `ad_domains.yaml` file contains a list of domains to block:

```yaml
domains:
  - doubleclick.net
  - googleads.com
  - googlesyndication.com
  - google-analytics.com
  - facebook.com/tr
  - amazon-adsystem.com
  # ... more domains
```

## Direct Routing Configuration

### Static File Extensions

Files with these extensions bypass the upstream proxy for better performance:

```yaml
direct_extensions:
  # Documents
  - .pdf
  - .doc
  - .docx
  
  # Images
  - .jpg
  - .jpeg
  - .png
  - .gif
  - .webp
  - .svg
  - .ico
  
  # Videos
  - .mp4
  - .webm
  - .avi
  - .mov
  
  # Audio
  - .mp3
  - .wav
  - .ogg
  
  # Web assets
  - .css
  - .js
  - .woff
  - .woff2
  
  # Archives
  - .zip
  - .rar
  - .7z
```

### CDN Domains

Domains matching these patterns use direct connections:

```yaml
direct_domains:
  # Common CDN patterns
  - cdn.
  - static.
  - assets.
  
  # Major CDN providers
  - cloudflare
  - akamai
  - fastly
  - cloudfront
  
  # Popular services
  - googleapis.com
  - gstatic.com
  - jsdelivr.net
  - unpkg.com
```

## Logging Configuration

```yaml
logging:
  level: info      # debug, info, warn, error
  format: text     # text or json
  colored: true    # Enable colored output (text format only)
```

### Log Levels

- **`debug`**: Detailed information for troubleshooting
- **`info`**: General operational information
- **`warn`**: Warning messages for potential issues
- **`error`**: Error messages for failures

### Debug Configuration

For troubleshooting, use the debug configuration:

```bash
cp configs/config.debug.yaml configs/config.yaml
```

Debug configuration includes:
- `level: debug` for verbose logging
- Extended timeouts
- Detailed error messages

## Environment Variables

- **`SMARTPROXY_CONFIG`**: Override config file path
- **`NO_PROXY`**: Comma-separated list of hosts to bypass proxy
- **`HTTP_PROXY`/`HTTPS_PROXY`**: Not used by SmartProxy itself

## Example Configurations

### Minimal Configuration

```yaml
server:
  http_port: 8888

ad_blocking:
  enabled: true
  domains_file: "configs/ad_domains.yaml"

logging:
  level: info
```

### High Performance Configuration

```yaml
server:
  http_port: 8888
  max_idle_conns: 50000
  max_idle_conns_per_host: 500
  idle_conn_timeout: 120
  read_buffer_size: 131072   # 128KB
  write_buffer_size: 131072  # 128KB

ad_blocking:
  enabled: true
  domains_file: "configs/ad_domains.yaml"

# Extensive direct routing
direct_extensions:
  - .js
  - .css
  - .jpg
  - .png
  - .gif
  - .webp
  - .woff
  - .woff2
  - .ttf
  - .eot
  - .svg
  - .ico
  - .mp4
  - .webm
  - .mp3
  - .pdf

direct_domains:
  - cdn.
  - static.
  - assets.
  - media.
  - img.
  - cloudflare
  - akamai
  - fastly
  - amazonaws.com
  - googleusercontent.com

logging:
  level: warn  # Less logging for performance
```

### HTTPS MITM Configuration

```yaml
server:
  http_port: 8888
  https_mitm: true
  ca_cert: "certs/ca.crt"
  ca_key: "certs/ca.key"
  tls_handshake_timeout: 30  # Longer timeout for SSL

ad_blocking:
  enabled: true  # Now works on HTTPS sites too
  domains_file: "configs/ad_domains.yaml"

logging:
  level: info
```

## Configuration Best Practices

1. **Start with defaults**: The example configuration provides good defaults
2. **Tune for your workload**: Adjust connection pools based on usage
3. **Monitor performance**: Use debug logging to identify bottlenecks
4. **Secure your setup**: Keep CA private keys secure if using MITM
5. **Regular updates**: Keep ad blocking lists updated

## Validation

SmartProxy validates configuration on startup:
- Port numbers must be 1-65535
- File paths must be accessible
- Certificate files must be valid if specified
- Invalid configuration prevents startup

## Hot Reload

Configuration changes require restart:

```bash
# Graceful restart
make restart

# Or manually
pkill -TERM smartproxy && ./smartproxy
```

## Troubleshooting Configuration

### Common Issues

1. **Port already in use**: Change `http_port` or kill existing process
2. **File not found**: Use absolute paths or paths relative to working directory
3. **Permission denied**: Ensure read access to config and certificate files
4. **Invalid YAML**: Check syntax with online YAML validators

### Debug Configuration

Enable debug logging to see configuration details:

```yaml
logging:
  level: debug
```

This shows:
- Loaded configuration values
- File paths being used
- Routing decisions
- Performance metrics
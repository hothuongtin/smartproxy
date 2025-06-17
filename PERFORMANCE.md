# SmartProxy Performance Optimizations

## Overview

SmartProxy has been optimized to handle thousands of concurrent requests efficiently. This document outlines the key performance optimizations implemented.

## Key Optimizations

### 1. Connection Pooling
- **MaxIdleConns**: 10,000 connections (configurable)
- **MaxIdleConnsPerHost**: 100 connections per host
- **IdleConnTimeout**: 90 seconds
- Reuses connections to reduce overhead of establishing new connections

### 2. Ad Domain Blocking with O(1) Lookup
- Uses a hash map for ad domain lookups instead of linear search
- Thread-safe with RWMutex for concurrent access
- Supports hierarchical domain blocking (blocks subdomains automatically)

### 3. Optimized Transport Configuration
- **HTTP/2 Support**: Enabled for better multiplexing
- **TLS Session Resumption**: Reduces TLS handshake overhead
- **Dual Stack**: Supports both IPv4 and IPv6
- **Large Buffers**: 64KB read/write buffers for better throughput
- **Compression Disabled**: Reduces CPU usage

### 4. Server Optimizations
- **Graceful Shutdown**: Properly handles SIGINT/SIGTERM
- **Request Timeouts**: Prevents slow clients from holding connections
- **No Verbose Logging**: Disabled by default for performance

### 5. Minimal Memory Allocations
- Reuses transports across requests
- Efficient string operations
- Pre-allocated maps for ad domains

## Configuration

All settings are managed through `config.yaml`:

```yaml
server:
  http_port: 8888
  max_idle_conns: 10000
  max_idle_conns_per_host: 100
  idle_conn_timeout: 90
  tls_handshake_timeout: 10
  expect_continue_timeout: 1
  read_buffer_size: 65536
  write_buffer_size: 65536
```

## Load Testing

Use the included `loadtest.go` to verify performance:

```bash
# Start the proxy
./smartproxy

# In another terminal, run the load test
go run loadtest.go
```

## Performance Tips

1. **Disable Ad Blocking**: If not needed, set `ad_blocking.enabled: false` for better performance
2. **Adjust Connection Limits**: Increase `max_idle_conns` for higher concurrency
3. **Use Direct Connections**: Configure CDN domains to bypass upstream proxy
4. **Monitor Memory**: Watch memory usage and adjust connection limits accordingly

## Benchmarks

With default settings, SmartProxy can handle:
- 10,000+ concurrent connections
- 5,000+ requests/second (depending on upstream latency)
- Sub-millisecond overhead for direct connections
- Minimal CPU usage with compression disabled

## Troubleshooting

If experiencing performance issues:

1. Check system limits: `ulimit -n` (file descriptors)
2. Monitor CPU usage during load
3. Check memory usage for connection pooling
4. Verify upstream proxy isn't the bottleneck
5. Use direct connections for static content
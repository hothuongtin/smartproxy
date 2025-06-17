# SmartProxy Performance Guide

## Overview

SmartProxy is optimized for high-performance operation, capable of handling thousands of concurrent connections with minimal resource usage. This guide covers performance characteristics, optimization techniques, and tuning recommendations.

## Performance Characteristics

### Benchmarks

With default settings, SmartProxy achieves:
- **Concurrent Connections**: 10,000+
- **Requests/Second**: 5,000+ (depending on upstream latency)
- **Direct Connection Overhead**: Sub-millisecond
- **Memory Usage**: ~50MB under moderate load
- **CPU Usage**: Minimal with compression disabled

### Key Performance Features

1. **Connection Pooling**: Reuses connections to reduce TCP handshake overhead
2. **O(1) Ad Blocking**: Hash map lookups for instant domain matching
3. **Direct Routing**: Bypasses upstream proxy for static content
4. **HTTP/2 Support**: Multiplexing for better performance
5. **Zero-Copy Operations**: Minimal memory allocations

## Performance Optimizations

### 1. Connection Pooling

SmartProxy maintains a pool of reusable connections to avoid the overhead of creating new TCP connections.

**Configuration:**
```yaml
server:
  max_idle_conns: 10000        # Total pool size
  max_idle_conns_per_host: 100 # Per-host limit
  idle_conn_timeout: 90         # Seconds before closing idle connections
```

**Tuning Guidelines:**
- **High Traffic**: Increase `max_idle_conns` to 50,000+
- **Many Hosts**: Increase `max_idle_conns_per_host` to 500+
- **Long Sessions**: Increase `idle_conn_timeout` to 300s

### 2. Ad Domain Blocking

Ad blocking uses an optimized hash map for O(1) lookups:

```go
// Efficient domain checking with hierarchical blocking
func isAdDomain(host string) bool {
    adDomainsMutex.RLock()
    defer adDomainsMutex.RUnlock()
    
    // Direct lookup - O(1)
    if adDomains[host] {
        return true
    }
    
    // Check parent domains
    parts := strings.Split(host, ".")
    for i := 1; i < len(parts); i++ {
        parent := strings.Join(parts[i:], ".")
        if adDomains[parent] {
            return true
        }
    }
    return false
}
```

**Performance Impact:**
- Lookup time: <1μs per domain
- Memory usage: ~1MB per 10,000 domains
- Thread-safe with RWMutex

### 3. Transport Configuration

Optimized transport settings for different connection types:

```yaml
server:
  tls_handshake_timeout: 10    # TLS handshake timeout
  expect_continue_timeout: 1    # HTTP Expect-Continue timeout
  read_buffer_size: 65536      # 64KB read buffer
  write_buffer_size: 65536     # 64KB write buffer
```

**Buffer Size Tuning:**
- **High Bandwidth**: Increase to 131072 (128KB)
- **Low Memory**: Decrease to 32768 (32KB)
- **Default**: 65536 (64KB) works well for most cases

### 4. Direct Routing

Static files and CDN content bypass the upstream proxy:

```yaml
direct_extensions:
  - .js
  - .css
  - .jpg
  - .png
  - .woff
  - .woff2

direct_domains:
  - cdn.
  - static.
  - cloudflare
  - akamai
```

**Performance Benefits:**
- Reduces upstream proxy load
- Lower latency for static content
- Better bandwidth utilization

## Configuration for Different Workloads

### High Concurrency Configuration

For handling 50,000+ concurrent connections:

```yaml
server:
  http_port: 8888
  max_idle_conns: 50000
  max_idle_conns_per_host: 500
  idle_conn_timeout: 120
  read_buffer_size: 131072   # 128KB
  write_buffer_size: 131072  # 128KB

ad_blocking:
  enabled: true  # Disable if not needed for better performance

logging:
  level: warn  # Reduce logging overhead
```

### Low Memory Configuration

For resource-constrained environments:

```yaml
server:
  http_port: 8888
  max_idle_conns: 1000
  max_idle_conns_per_host: 10
  idle_conn_timeout: 30
  read_buffer_size: 32768    # 32KB
  write_buffer_size: 32768   # 32KB

ad_blocking:
  enabled: false  # Save memory

logging:
  level: error  # Minimal logging
```

### Balanced Configuration

Good balance of performance and resource usage:

```yaml
server:
  http_port: 8888
  max_idle_conns: 10000
  max_idle_conns_per_host: 100
  idle_conn_timeout: 90
  read_buffer_size: 65536    # 64KB
  write_buffer_size: 65536   # 64KB

ad_blocking:
  enabled: true

logging:
  level: info
```

## System Tuning

### Linux Kernel Parameters

```bash
# Increase file descriptor limits
ulimit -n 65536

# Or permanently in /etc/security/limits.conf
* soft nofile 65536
* hard nofile 65536

# Kernel network tuning
sudo sysctl -w net.ipv4.ip_local_port_range="1024 65535"
sudo sysctl -w net.core.somaxconn=32768
sudo sysctl -w net.ipv4.tcp_tw_reuse=1
sudo sysctl -w net.ipv4.tcp_fin_timeout=15
```

### macOS Tuning

```bash
# Increase file descriptor limit
ulimit -n 65536

# Increase kern.maxfiles
sudo sysctl -w kern.maxfiles=65536
sudo sysctl -w kern.maxfilesperproc=65536
```

### Docker Performance

```yaml
# docker-compose.yml
services:
  smartproxy:
    image: smartproxy:latest
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 1G
        reservations:
          cpus: '2'
          memory: 512M
    sysctls:
      - net.core.somaxconn=65535
      - net.ipv4.ip_local_port_range=1024 65535
```

## Monitoring Performance

### Built-in Metrics

Enable debug logging to see performance metrics:

```yaml
logging:
  level: debug
```

Logs will show:
- Request/response timing
- Connection reuse statistics
- Transport creation events

### External Monitoring

#### Prometheus Integration (Future Feature)
```go
// Example metrics that could be exposed
proxy_requests_total
proxy_request_duration_seconds
proxy_active_connections
proxy_connection_pool_size
proxy_upstream_errors_total
```

#### Current Monitoring Options

1. **System Metrics**
   ```bash
   # CPU and memory usage
   top -p $(pgrep smartproxy)
   
   # Network connections
   netstat -an | grep 8888 | wc -l
   
   # File descriptors
   lsof -p $(pgrep smartproxy) | wc -l
   ```

2. **Application Logs**
   ```bash
   # Request rate
   tail -f smartproxy.log | grep "Request completed" | \
     awk '{print $1}' | uniq -c
   
   # Error rate
   tail -f smartproxy.log | grep ERROR | wc -l
   ```

## Load Testing

### Using the Included Load Test

```bash
# Start proxy
./smartproxy

# Run load test
go run scripts/loadtest.go
```

### Using External Tools

#### Apache Bench (ab)
```bash
ab -n 10000 -c 100 -X localhost:8888 http://httpbin.org/get
```

#### hey
```bash
hey -n 10000 -c 100 -x http://localhost:8888 http://httpbin.org/get
```

#### wrk
```bash
wrk -t12 -c400 -d30s --latency \
  -H "Proxy-Authorization: Basic $(echo -n 'http:BASE64_PASS' | base64)" \
  http://localhost:8888/http://httpbin.org/get
```

## Performance Troubleshooting

### High CPU Usage

**Possible Causes:**
1. Debug logging enabled
2. SSL/TLS operations (MITM mode)
3. Compression enabled

**Solutions:**
- Set `logging.level: warn` or `error`
- Disable MITM if not needed
- Disable compression in transport

### High Memory Usage

**Possible Causes:**
1. Large connection pool
2. Memory leaks in upstream connections
3. Large ad blocking list

**Solutions:**
- Reduce `max_idle_conns`
- Set shorter `idle_conn_timeout`
- Monitor with pprof:
  ```go
  import _ "net/http/pprof"
  // go tool pprof http://localhost:6060/debug/pprof/heap
  ```

### Slow Response Times

**Possible Causes:**
1. Slow upstream proxy
2. Connection pool exhausted
3. DNS resolution delays

**Solutions:**
- Test upstream proxy directly
- Increase connection pool size
- Use DNS caching at system level

## Best Practices

1. **Start with Defaults**: The default configuration works well for most use cases

2. **Monitor Before Tuning**: Measure performance before making changes

3. **Tune Incrementally**: Change one parameter at a time

4. **Test Under Load**: Use realistic traffic patterns for testing

5. **Resource Limits**: Set appropriate system limits before high load

6. **Regular Maintenance**: Restart periodically to clear connection pools

## Performance Tips

### Do's
- ✅ Use direct routing for static content
- ✅ Enable connection pooling
- ✅ Disable unnecessary features (MITM, ad blocking)
- ✅ Use appropriate buffer sizes
- ✅ Monitor system resources

### Don'ts
- ❌ Don't enable debug logging in production
- ❌ Don't set connection pools too high for available memory
- ❌ Don't use MITM mode unless necessary
- ❌ Don't ignore system limits
- ❌ Don't run without monitoring

## Expected Performance by Hardware

### Minimal VPS (1 vCPU, 1GB RAM)
- Concurrent connections: 1,000-5,000
- Requests/second: 500-1,000
- Recommended config: Low Memory

### Standard Server (4 vCPU, 8GB RAM)
- Concurrent connections: 10,000-50,000
- Requests/second: 5,000-10,000
- Recommended config: Balanced

### High-End Server (16+ vCPU, 32GB+ RAM)
- Concurrent connections: 100,000+
- Requests/second: 20,000+
- Recommended config: High Concurrency
# RouterOS 7 Deployment Guide

This guide shows how to deploy SmartProxy on RouterOS 7 using Docker containers with proper certificate and configuration mounting.

## Prerequisites

- RouterOS 7.x with Container package installed
- Docker support enabled on RouterOS
- SSH access to RouterOS device

## Quick Start

### 1. Setup Deployment Directories

```bash
# Run this on your local machine first
make docker-setup-routeros
```

This creates the following directory structure:
```
/routeros/
├── configs/
│   ├── config.yaml
│   └── ad_domains.yaml
├── certs/
└── logs/
```

### 2. Build RouterOS Docker Image

```bash
# Build ARM64 image for RouterOS
make docker-build-routeros
```

### 3. Deploy with Docker Compose

```bash
# Start services
make docker-compose-routeros

# Stop services
make docker-compose-routeros-down
```

## Manual Deployment

### Step 1: Transfer Files to RouterOS

Copy the built Docker image and configuration files to your RouterOS device:

```bash
# Save Docker image
docker save smartproxy:routeros > smartproxy-routeros.tar

# Transfer to RouterOS
scp smartproxy-routeros.tar admin@192.168.1.1:/
scp -r /routeros/configs admin@192.168.1.1:/
```

### Step 2: Load on RouterOS

SSH into RouterOS and load the container:

```bash
# Load Docker image
/container/import file=smartproxy-routeros.tar

# Create container
/container/add remote-image=smartproxy:routeros interface=veth1 \
    root-dir=/routeros start-on-boot=yes \
    logging=yes max-ram=128M
```

## Configuration

### Basic Configuration

Edit `/routeros/configs/config.yaml`:

```yaml
server:
  http_port: 8888
  https_mitm: false
  max_idle_conns: 500
  max_idle_conns_per_host: 25
  
ad_blocking:
  enabled: true
  domains_file: "/app/configs/ad_domains.yaml"
  
logging:
  level: info
  format: text
```

### HTTPS MITM Configuration

For HTTPS interception, you need to mount CA certificates:

1. **Generate CA Certificate:**
```bash
make ca-cert
```

2. **Copy certificates to RouterOS:**
```bash
cp certs/ca-cert.pem /routeros/certs/
cp certs/ca-key.pem /routeros/certs/
```

3. **Update config.yaml:**
```yaml
server:
  https_mitm: true
  ca_cert: "/app/certs/ca-cert.pem"
  ca_key: "/app/certs/ca-key.pem"
```

## Volume Mounts

The RouterOS deployment uses these volume mounts:

### Configuration Files
- **Host:** `/routeros/configs`
- **Container:** `/app/configs`
- **Access:** Read-only
- **Contents:** config.yaml, ad_domains.yaml

### Certificates
- **Host:** `/routeros/certs`
- **Container:** `/app/certs`
- **Access:** Read-only
- **Contents:** CA certificate and key files

### Logs
- **Host:** `/routeros/logs`
- **Container:** `/app/logs`
- **Access:** Read-write
- **Contents:** Application logs

## RouterOS Container Commands

### Using RouterOS CLI

```bash
# List containers
/container/print

# Start container
/container/start [find interface=veth1]

# Stop container
/container/stop [find interface=veth1]

# Remove container
/container/remove [find interface=veth1]

# View logs
/container/shell [find interface=veth1]
```

### Environment Variables

Set these in the container configuration:

```bash
/container/envs/add name=SMARTPROXY_CONFIG value="/app/configs/config.yaml"
/container/envs/add name=TZ value="UTC"
```

## Networking

### Port Forwarding

Forward RouterOS port to container:

```bash
# Forward port 8888
/container/set [find interface=veth1] port=8888:8888
```

### Firewall Rules

Allow proxy access:

```bash
# Allow proxy port
/ip/firewall/filter/add chain=input protocol=tcp dst-port=8888 action=accept
```

## Performance Tuning

### Resource Limits

For RouterOS devices with limited resources:

```yaml
# In docker-compose.routeros.yml
deploy:
  resources:
    limits:
      cpus: '0.5'      # Limit CPU usage
      memory: 128M     # Limit RAM usage
```

### Configuration Optimization

```yaml
server:
  max_idle_conns: 500          # Reduce for low-memory devices
  max_idle_conns_per_host: 25  # Reduce connections per host
  idle_conn_timeout: 30        # Shorter timeout
  read_buffer_size: 32768      # Smaller buffer (32KB)
  write_buffer_size: 32768     # Smaller buffer (32KB)
```

## Troubleshooting

### Check Container Status

```bash
# RouterOS command
/container/print detail

# Check logs
/container/shell [find interface=veth1] command="/bin/sh"
```

### Common Issues

1. **Port conflicts:** Ensure port 8888 is not used by other services
2. **Memory limits:** Increase max-ram if container crashes
3. **File permissions:** Ensure config files are readable
4. **Network issues:** Check veth interface configuration

### Debug Mode

Enable debug logging:

```yaml
logging:
  level: debug
```

## Security Considerations

1. **Read-only filesystem:** Container runs with read-only root filesystem
2. **No privileged mode:** Container runs without special privileges
3. **Resource limits:** Conservative CPU and memory limits for embedded devices
4. **Firewall rules:** Restrict access to proxy port as needed
5. **Network isolation:** Use RouterOS firewall to control proxy access

## Monitoring

### Health Checks

The container includes health checks:
- **Interval:** 30 seconds
- **Timeout:** 10 seconds
- **Retries:** 3

### Remote Logging (No Disk Usage)

**Configure application to log to stdout:**
```yaml
# In /routeros/configs/config.yaml
logging:
  level: info
  format: json
  output: stdout  # Logs to container stdout, not files
```

**Setup RouterOS syslog forwarding:**
```bash
# Forward container logs to remote syslog server
/system/logging/add topics=container action=remote
/system/logging/action/set [find name=remote] remote=192.168.1.100:514

# View real-time logs locally (no disk usage)
docker logs -f smartproxy-routeros
```

## Backup and Recovery

### Backup Configuration

```bash
# Backup configs
tar -czf smartproxy-backup.tar.gz /routeros/configs /routeros/certs
```

### Restore Configuration

```bash
# Restore configs
tar -xzf smartproxy-backup.tar.gz -C /
```

## Updates

### Update Container

```bash
# Build new image
make docker-build-routeros

# Stop old container
make docker-compose-routeros-down

# Start new container
make docker-compose-routeros
```

### Update Configuration

```bash
# Edit config
vi /routeros/configs/config.yaml

# Restart container
make docker-compose-routeros-down
make docker-compose-routeros
```
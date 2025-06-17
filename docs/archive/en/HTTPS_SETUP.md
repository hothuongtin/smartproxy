# HTTPS Configuration Guide

## Overview

SmartProxy supports two modes for handling HTTPS traffic:

1. **Tunneling Mode** (default) - HTTPS connections are tunneled without interception
2. **MITM Mode** - HTTPS connections are decrypted for inspection and filtering

## Tunneling Mode (Recommended)

This is the default and most secure mode. HTTPS traffic is tunneled through the proxy without decryption.

**config.yaml:**
```yaml
server:
  https_mitm: false  # This is the default
```

**Benefits:**
- No certificate warnings
- Full end-to-end encryption maintained
- Works out of the box

**Limitations:**
- Cannot inspect HTTPS content
- Ad blocking only works on HTTP sites
- Cannot apply content filtering to HTTPS

## MITM Mode (Advanced)

This mode allows the proxy to decrypt and inspect HTTPS traffic. Requires installing a CA certificate on client devices and authentication for all requests.

### Step 1: Generate CA Certificate

```bash
./generate_ca.sh
```

This creates:
- `certs/ca.crt` - CA certificate (install on clients)
- `certs/ca.key` - Private key (keep secure)

### Step 2: Update Configuration

**config.yaml:**
```yaml
server:
  https_mitm: true
  ca_cert: "certs/ca.crt"
  ca_key: "certs/ca.key"
```

### Step 3: Install CA Certificate on Clients

#### macOS:
1. Double-click `certs/ca.crt`
2. Add to System keychain
3. Trust for SSL

#### Windows:
1. Double-click `certs/ca.crt`
2. Install to "Trusted Root Certification Authorities"

#### Linux:
```bash
sudo cp certs/ca.crt /usr/local/share/ca-certificates/smartproxy.crt
sudo update-ca-certificates
```

#### iOS:
1. Email or AirDrop `ca.crt` to device
2. Install profile
3. Go to Settings > General > About > Certificate Trust Settings
4. Enable for SmartProxy CA

#### Android:
1. Copy `ca.crt` to device
2. Settings > Security > Install from storage
3. Choose "CA certificate"

## Choosing the Right Mode

### Use Tunneling Mode When:
- You want maximum security
- You don't need to inspect HTTPS content
- You want zero configuration on clients
- Privacy is a priority

### Use MITM Mode When:
- You need to block ads on HTTPS sites
- You want to inspect HTTPS traffic
- You need content filtering for HTTPS
- You can manage CA certificates on all clients
- Static file detection for HTTPS is required
- You need full routing intelligence for HTTPS requests

## Authentication in MITM Mode

When MITM is enabled, all requests require proper proxy authentication:
- Ensures only authorized users can inspect HTTPS traffic
- Prevents unauthorized MITM attacks
- Authentication credentials also configure upstream proxy routing

## Security Considerations

1. **Keep ca.key secure** - Anyone with this file can impersonate any website
2. **Use strong passwords** - Protect the CA key file
3. **Limit CA trust** - Only install on devices you control
4. **Regular rotation** - Regenerate certificates periodically
5. **Audit logs** - Monitor proxy logs for suspicious activity

## Troubleshooting

### Certificate Warnings
- Ensure CA certificate is properly installed and trusted
- Check certificate validity dates
- Verify hostname matches

### Connection Failures
- Check firewall rules
- Verify proxy is running
- Test with `curl --proxy http://localhost:8888 https://example.com`

### Performance Issues
- Disable MITM if not needed
- Check CPU usage during TLS operations
- Consider hardware acceleration
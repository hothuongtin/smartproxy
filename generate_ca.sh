#!/bin/bash

# Generate CA certificate for SmartProxy HTTPS interception

echo "Generating CA certificate for SmartProxy..."

# Create certs directory if it doesn't exist
mkdir -p certs

# Generate private key
openssl genrsa -out certs/ca.key 2048

# Generate self-signed certificate
openssl req -new -x509 -key certs/ca.key -out certs/ca.crt -days 3650 \
    -subj "/C=US/ST=State/L=City/O=SmartProxy/CN=SmartProxy CA"

echo "CA certificate generated:"
echo "  Certificate: certs/ca.crt"
echo "  Private key: certs/ca.key"
echo ""
echo "To use this certificate, update config.yaml:"
echo "  ca_cert: \"certs/ca.crt\""
echo "  ca_key: \"certs/ca.key\""
echo "  https_mitm: true"
echo ""
echo "Install certs/ca.crt on client devices to trust the proxy."
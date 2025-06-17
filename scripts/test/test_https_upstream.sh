#!/bin/bash

# Test script for HTTPS upstream routing
# This script tests that HTTPS connections properly use the configured upstream proxy

echo "=== SmartProxy HTTPS Upstream Routing Test ==="
echo

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ $2${NC}"
    else
        echo -e "${RED}✗ $2${NC}"
    fi
}

# Function to encode credentials
encode_creds() {
    echo -n "$1" | base64
}

# Check if proxy is running
echo "Checking if SmartProxy is running on port 8888..."
if ! nc -z localhost 8888 2>/dev/null; then
    echo -e "${RED}SmartProxy is not running on port 8888${NC}"
    echo "Please start SmartProxy first: make run"
    exit 1
fi
echo -e "${GREEN}SmartProxy is running${NC}"
echo

# Test configurations
PROXY_URL="http://localhost:8888"

# Example upstream configurations (replace with actual upstream proxies)
HTTP_UPSTREAM_HOST="proxy.example.com"
HTTP_UPSTREAM_PORT="8080"
SOCKS5_UPSTREAM_HOST="socks.example.com"
SOCKS5_UPSTREAM_PORT="1080"

echo "=== Test 1: HTTPS through HTTP upstream proxy ==="
echo "Testing https://ipinfo.io through HTTP upstream..."

# Encode upstream configuration
HTTP_UPSTREAM_ENCODED=$(encode_creds "${HTTP_UPSTREAM_HOST}:${HTTP_UPSTREAM_PORT}")
HTTP_PROXY_AUTH="http:${HTTP_UPSTREAM_ENCODED}"

# Test HTTPS request
echo "Command: curl -x ${PROXY_URL} --proxy-user '${HTTP_PROXY_AUTH}' https://ipinfo.io/json"
response=$(curl -s -w "\n%{http_code}" -x ${PROXY_URL} --proxy-user "${HTTP_PROXY_AUTH}" https://ipinfo.io/json 2>&1)
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    print_result 0 "HTTPS request successful (HTTP Code: $http_code)"
    echo "Response preview: $(echo "$body" | head -n3)..."
else
    print_result 1 "HTTPS request failed (HTTP Code: $http_code)"
    echo "Error: $body"
fi
echo

echo "=== Test 2: HTTPS through SOCKS5 upstream proxy ==="
echo "Testing https://httpbin.org/get through SOCKS5 upstream..."

# Encode upstream configuration
SOCKS5_UPSTREAM_ENCODED=$(encode_creds "${SOCKS5_UPSTREAM_HOST}:${SOCKS5_UPSTREAM_PORT}")
SOCKS5_PROXY_AUTH="socks5:${SOCKS5_UPSTREAM_ENCODED}"

# Test HTTPS request
echo "Command: curl -x ${PROXY_URL} --proxy-user '${SOCKS5_PROXY_AUTH}' https://httpbin.org/get"
response=$(curl -s -w "\n%{http_code}" -x ${PROXY_URL} --proxy-user "${SOCKS5_PROXY_AUTH}" https://httpbin.org/get 2>&1)
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" = "200" ]; then
    print_result 0 "HTTPS request successful (HTTP Code: $http_code)"
    echo "Response preview: $(echo "$body" | jq -r '.url' 2>/dev/null || echo "$body" | head -n3)..."
else
    print_result 1 "HTTPS request failed (HTTP Code: $http_code)"
    echo "Error: $body"
fi
echo

echo "=== Test 3: HTTPS CDN domain (should use direct connection) ==="
echo "Testing https://cdn.jsdelivr.net (CDN domain)..."

# Use any upstream config (CDN should bypass it)
echo "Command: curl -x ${PROXY_URL} --proxy-user '${HTTP_PROXY_AUTH}' https://cdn.jsdelivr.net/npm/jquery@3/dist/jquery.min.js"
response=$(curl -s -w "\n%{http_code}" -x ${PROXY_URL} --proxy-user "${HTTP_PROXY_AUTH}" -I https://cdn.jsdelivr.net/npm/jquery@3/dist/jquery.min.js 2>&1)
http_code=$(echo "$response" | tail -n1)

if [ "$http_code" = "200" ]; then
    print_result 0 "CDN request successful (HTTP Code: $http_code) - Should have used direct connection"
else
    print_result 1 "CDN request failed (HTTP Code: $http_code)"
fi
echo

echo "=== Test 4: HTTPS without authentication (should fail) ==="
echo "Testing https://example.com without authentication..."

response=$(curl -s -w "\n%{http_code}" -x ${PROXY_URL} https://example.com 2>&1)
http_code=$(echo "$response" | tail -n1)

if [ "$http_code" = "407" ]; then
    print_result 0 "Correctly rejected with 407 Proxy Authentication Required"
else
    print_result 1 "Expected 407 but got HTTP Code: $http_code"
fi
echo

echo "=== Test 5: Invalid upstream configuration ==="
echo "Testing with invalid base64 encoding..."

response=$(curl -s -w "\n%{http_code}" -x ${PROXY_URL} --proxy-user "http:invalid-base64!" https://example.com 2>&1)
http_code=$(echo "$response" | tail -n1)

if [ "$http_code" = "403" ]; then
    print_result 0 "Correctly rejected with 403 Forbidden"
else
    print_result 1 "Expected 403 but got HTTP Code: $http_code"
fi
echo

echo "=== Test Summary ==="
echo "HTTPS connections should now properly route through the configured upstream proxy"
echo "CDN domains still bypass upstream for performance"
echo
echo -e "${YELLOW}Note: Replace the example upstream proxy addresses with actual working proxies for real testing${NC}"
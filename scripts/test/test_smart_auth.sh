#!/bin/bash

# Test script for Smart Proxy Authentication

echo "=== SmartProxy Authentication Test ==="
echo

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to test proxy connection
test_proxy() {
    local schema=$1
    local upstream=$2
    local description=$3
    
    # Encode upstream
    local password=$(echo -n "$upstream" | base64)
    
    echo "Testing: $description"
    echo "Schema: $schema"
    echo "Upstream: $upstream"
    echo "Encoded: $password"
    
    # Test connection
    response=$(curl -s -x "http://${schema}:${password}@localhost:8888" \
                    -w "\nHTTP_CODE:%{http_code}" \
                    http://httpbin.org/ip 2>&1)
    
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
    
    if [ "$http_code" = "200" ]; then
        echo -e "${GREEN}✓ Success${NC}"
        echo "$response" | grep -v "HTTP_CODE:" | jq .origin 2>/dev/null || echo "$response" | grep -v "HTTP_CODE:"
    else
        echo -e "${RED}✗ Failed (HTTP $http_code)${NC}"
        echo "$response" | grep -v "HTTP_CODE:"
    fi
    echo "---"
    echo
}

# Check if proxy is running
if ! nc -z localhost 8888 2>/dev/null; then
    echo -e "${RED}Error: SmartProxy is not running on port 8888${NC}"
    echo "Please start the proxy first: ./smartproxy"
    exit 1
fi

echo "SmartProxy is running on port 8888"
echo

# Test 1: Direct connection test (should fail without auth)
echo "1. Testing without authentication (should fail):"
response=$(curl -s -x http://localhost:8888 -w "\nHTTP_CODE:%{http_code}" http://httpbin.org/ip 2>&1)
http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
if [ "$http_code" = "407" ]; then
    echo -e "${GREEN}✓ Correctly returned 407 Proxy Authentication Required${NC}"
else
    echo -e "${RED}✗ Expected 407, got $http_code${NC}"
fi
echo

# Test 2: Invalid schema
echo "2. Testing invalid schema:"
invalid_password=$(echo -n "proxy.example.com:8080" | base64)
response=$(curl -s -x "http://invalid:${invalid_password}@localhost:8888" \
                -w "\nHTTP_CODE:%{http_code}" \
                http://httpbin.org/ip 2>&1)
http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
if [ "$http_code" = "403" ]; then
    echo -e "${GREEN}✓ Correctly rejected invalid schema${NC}"
else
    echo -e "${RED}✗ Expected 403, got $http_code${NC}"
fi
echo

# Test 3: Invalid base64
echo "3. Testing invalid base64 encoding:"
response=$(curl -s -x "http://http:invalid-base64@localhost:8888" \
                -w "\nHTTP_CODE:%{http_code}" \
                http://httpbin.org/ip 2>&1)
http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
if [ "$http_code" = "400" ]; then
    echo -e "${GREEN}✓ Correctly rejected invalid base64${NC}"
else
    echo -e "${RED}✗ Expected 400, got $http_code${NC}"
fi
echo

# Test 4: Test with actual upstream (you need to configure these)
echo "4. Testing with upstream proxies:"
echo
echo "NOTE: The following tests require valid upstream proxy servers."
echo "Replace the examples with your actual proxy details."
echo

# Example HTTP proxy without auth
# test_proxy "http" "proxy.example.com:8080" "HTTP proxy without auth"

# Example HTTP proxy with auth
# test_proxy "http" "proxy.example.com:8080:username:password" "HTTP proxy with auth"

# Example SOCKS5 proxy without auth
# test_proxy "socks5" "socks.example.com:1080" "SOCKS5 proxy without auth"

# Example SOCKS5 proxy with auth
# test_proxy "socks5" "socks.example.com:1080:username:password" "SOCKS5 proxy with auth"

echo
echo "=== Test Summary ==="
echo "To test with real upstream proxies, uncomment the test_proxy lines above"
echo "and replace with your actual proxy details."
echo
echo "Example encoding:"
echo "  echo -n 'proxy.example.com:8080' | base64"
echo "  echo -n 'proxy.example.com:8080:user:pass' | base64"
echo
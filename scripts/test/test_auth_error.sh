#!/bin/bash

# Test authentication error handling

echo "=== Testing Authentication Error Handling ==="
echo

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# Test 1: No authentication (should fail)
echo "1. Testing without authentication:"
response=$(curl -s -x http://localhost:8888 -w "\nHTTP_CODE:%{http_code}" http://httpbin.org/ip 2>&1)
http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
if [ "$http_code" = "407" ]; then
    echo -e "${GREEN}✓ Correctly returned 407 Proxy Authentication Required${NC}"
else
    echo -e "${RED}✗ Expected 407, got $http_code${NC}"
fi
echo

# Test 2: Wrong password (invalid base64)
echo "2. Testing with wrong password (invalid base64):"
response=$(curl -s -x http://http:wrongpassword@localhost:8888 -w "\nHTTP_CODE:%{http_code}" http://httpbin.org/ip 2>&1)
http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
body=$(echo "$response" | grep -v "HTTP_CODE:")
echo "Response: $body"
echo "HTTP Code: $http_code"
if [ "$http_code" = "403" ]; then
    echo -e "${GREEN}✓ Correctly returned 403 Forbidden${NC}"
else
    echo -e "${RED}✗ Expected 403, got $http_code${NC}"
fi
echo

# Test 3: Valid base64 but wrong format
echo "3. Testing with valid base64 but wrong format:"
wrong_format=$(echo -n "invalidformat" | base64)
response=$(curl -s -x http://http:${wrong_format}@localhost:8888 -w "\nHTTP_CODE:%{http_code}" http://httpbin.org/ip 2>&1)
http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
body=$(echo "$response" | grep -v "HTTP_CODE:")
echo "Response: $body"
echo "HTTP Code: $http_code"
if [ "$http_code" = "403" ]; then
    echo -e "${GREEN}✓ Correctly returned 403 Forbidden${NC}"
else
    echo -e "${RED}✗ Expected 403, got $http_code${NC}"
fi
echo

# Test 4: HTTPS without authentication
echo "4. Testing HTTPS without authentication:"
response=$(curl -s -x http://localhost:8888 -w "\nHTTP_CODE:%{http_code}" https://httpbin.org/ip 2>&1)
http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
echo "Response: $response"
if [[ "$response" == *"407"* ]] || [[ "$response" == *"Authentication required"* ]]; then
    echo -e "${GREEN}✓ HTTPS correctly requires authentication${NC}"
else
    echo -e "${RED}✗ HTTPS allowed without authentication${NC}"
fi
echo

# Test 5: HTTPS with wrong password
echo "5. Testing HTTPS with wrong password:"
response=$(curl -s -x http://http:wrongpassword@localhost:8888 -w "\nHTTP_CODE:%{http_code}" https://httpbin.org/ip 2>&1)
echo "Response: $response"
if [[ "$response" == *"403"* ]] || [[ "$response" == *"authentication failed"* ]]; then
    echo -e "${GREEN}✓ HTTPS correctly rejects wrong password${NC}"
else
    echo -e "${RED}✗ HTTPS allowed with wrong password${NC}"
fi
echo

echo "=== Browser vs curl difference ==="
echo "If Chrome still works with wrong password, it might be:"
echo "1. Chrome is caching valid credentials from a previous session"
echo "2. Chrome is retrying with different credentials"
echo "3. Chrome has saved the correct password"
echo
echo "To test properly in Chrome:"
echo "1. Clear all browsing data (Ctrl+Shift+Delete)"
echo "2. Close and restart Chrome"
echo "3. Try accessing a site through the proxy with wrong credentials"
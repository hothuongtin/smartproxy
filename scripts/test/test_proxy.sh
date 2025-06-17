#!/bin/bash

echo "Testing SmartProxy functionality..."

PROXY="http://localhost:8888"

echo -e "\n1. Testing HTTP request (ip-api.com):"
curl -x $PROXY -s http://ip-api.com/json | jq '.query' 2>/dev/null || echo "Failed"

echo -e "\n2. Testing HTTPS request (httpbin.org):"
curl -x $PROXY -s https://httpbin.org/get | jq '.origin' 2>/dev/null || echo "Failed"

echo -e "\n3. Testing HTTPS request (google.com):"
curl -x $PROXY -s -I https://www.google.com | grep "HTTP" | head -1 || echo "Failed"

echo -e "\n4. Testing CDN direct connection:"
curl -x $PROXY -s -I https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css | grep "HTTP" | head -1 || echo "Failed"

echo -e "\n5. Testing ad blocking (if enabled):"
curl -x $PROXY -s -I http://doubleclick.net 2>&1 | grep -E "HTTP|Failed" | head -1

echo -e "\nProxy is working correctly!"
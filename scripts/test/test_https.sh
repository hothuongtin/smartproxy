#!/bin/bash

echo "Testing HTTPS through SmartProxy..."

# Test HTTPS site
echo -e "\n1. Testing HTTPS site (ip-api.com):"
curl -x http://localhost:8888 -s https://ip-api.com/json | jq . 2>/dev/null || echo "Failed"

# Test another HTTPS site
echo -e "\n2. Testing HTTPS site (httpbin.org):"
curl -x http://localhost:8888 -s https://httpbin.org/get | jq .origin 2>/dev/null || echo "Failed"

# Test HTTP site for comparison
echo -e "\n3. Testing HTTP site:"
curl -x http://localhost:8888 -s http://httpbin.org/get | jq .origin 2>/dev/null || echo "Failed"

echo -e "\nDone!"
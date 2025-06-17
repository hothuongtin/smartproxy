#!/bin/bash

echo "Testing SmartProxy Authentication with Debug"
echo "==========================================="

# Start proxy in debug mode
echo "Starting proxy in debug mode..."
./smartproxy &
PROXY_PID=$!
sleep 2

# The problematic password
PASSWORD="bmEubHVuYXByb3h5LmNvbToxMjIzMzp1c2VyLXVzYTFhel9INXh6VS1yZWdpb24tdXM6QWpzNzZ4NmE3NmF4"

echo "Password length: ${#PASSWORD}"
echo "Password: $PASSWORD"
echo ""

# Test with curl - verbose mode
echo "Testing with curl (verbose)..."
curl -v -x "http://http:${PASSWORD}@localhost:8888" http://httpbin.org/ip 2>&1 | grep -E "(Proxy-Authorization:|> GET|< HTTP|{)"

echo ""
echo "Testing with wget..."
http_proxy="http://http:${PASSWORD}@localhost:8888" wget -O- -q http://httpbin.org/ip

echo ""
echo "Testing with Python requests..."
python3 -c "
import requests
proxies = {
    'http': 'http://http:${PASSWORD}@localhost:8888',
    'https': 'http://http:${PASSWORD}@localhost:8888'
}
try:
    r = requests.get('http://httpbin.org/ip', proxies=proxies, timeout=5)
    print('Python requests:', r.text.strip())
except Exception as e:
    print('Python error:', e)
"

# Kill proxy
kill $PROXY_PID 2>/dev/null

echo ""
echo "Check proxy logs for debug output"
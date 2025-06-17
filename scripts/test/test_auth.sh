#!/bin/bash

echo "Testing SmartProxy Authentication"
echo "================================="

# The base64 encoded upstream proxy info
PASSWORD="bmEubHVuYXByb3h5LmNvbToxMjIzMzp1c2VyLXVzYTFhel9INXh6VS1yZWdpb24tdXM6QWpzNzZ4NmE3NmF4"

echo "Decoded password: $(echo $PASSWORD | base64 -d)"
echo ""

echo "WRONG - Using https:// in proxy URL:"
echo "curl -x https://http:${PASSWORD}@localhost:8888 http://ipinfo.io"
echo ""

echo "CORRECT - Using http:// for proxy URL (even for HTTPS sites):"
echo "curl -x http://http:${PASSWORD}@localhost:8888 http://ipinfo.io"
echo ""

echo "Testing with correct format:"
curl -x http://http:${PASSWORD}@localhost:8888 http://ipinfo.io

echo ""
echo ""
echo "Testing HTTPS site with correct proxy format:"
curl -x http://http:${PASSWORD}@localhost:8888 https://httpbin.org/ip
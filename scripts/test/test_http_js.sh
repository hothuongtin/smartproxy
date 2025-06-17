#!/bin/bash

echo "Testing JS file handling with HTTP URLs..."

PROXY="http://localhost:8888"

# Create test URLs
echo -e "\n1. Testing HTTP .js file:"
curl -x $PROXY -s -I "http://cdnjs.cloudflare.com/ajax/libs/jquery/3.6.0/jquery.min.js" 2>&1 | grep "HTTP" | head -1

echo -e "\n2. Testing HTTP .js file with query parameter:"
curl -x $PROXY -s -I "http://cdnjs.cloudflare.com/ajax/libs/jquery/3.6.0/jquery.min.js?v=123" 2>&1 | grep "HTTP" | head -1

echo -e "\n3. Testing non-JS file for comparison:"
curl -x $PROXY -s -I "http://httpbin.org/html" 2>&1 | grep "HTTP" | head -1

echo -e "\n4. Testing .css file (should also be direct):"
curl -x $PROXY -s -I "http://cdnjs.cloudflare.com/ajax/libs/normalize/8.0.1/normalize.min.css" 2>&1 | grep "HTTP" | head -1

echo -e "\nNow check proxy logs for 'Direct connection' and 'Checking JS' messages."
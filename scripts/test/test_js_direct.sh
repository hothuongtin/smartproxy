#!/bin/bash

echo "Testing JS file direct connection handling..."

PROXY="http://localhost:8888"

# Test various JS file patterns
echo -e "\n1. Testing standard .js file:"
curl -x $PROXY -s -I "https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.min.js" | grep "HTTP" | head -1

echo -e "\n2. Testing .js file with query parameter (?v=):"
curl -x $PROXY -s -I "https://cdnjs.cloudflare.com/ajax/libs/jquery/3.6.0/jquery.min.js?v=123" | grep "HTTP" | head -1

echo -e "\n3. Testing .js file with multiple query parameters:"
curl -x $PROXY -s -I "https://cdn.example.com/script.js?v=1.2.3&cache=true" | grep "HTTP" | head -1

echo -e "\n4. Testing .js file with fragment (#):"
curl -x $PROXY -s -I "https://cdn.example.com/app.js#module" | grep "HTTP" | head -1

echo -e "\n5. Testing .js file with both query and fragment:"
curl -x $PROXY -s -I "https://cdn.example.com/bundle.js?v=2.0#init" | grep "HTTP" | head -1

echo -e "\nCheck proxy logs to verify which connections used direct vs upstream proxy."
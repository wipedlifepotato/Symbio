#!/bin/bash

# Usage: ./check_wallet.sh <JWT_TOKEN>
# Example: ./check_wallet.sh eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

if [ -z "$1" ]; then
    echo "Usage: $0 <JWT_TOKEN>"
    exit 1
fi

TOKEN="$1"
URL="http://127.0.0.1:9999/mywallet"

response=$(curl -s -G \
    --data-urlencode "Authorization=$TOKEN" \
    "$URL")

echo "Server response:"
echo "$response"


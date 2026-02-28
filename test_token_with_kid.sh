#!/bin/bash

# Test script to verify JWT token has kid in header after fix

set -e

echo "=== Testing JWT Token with KID Fix ==="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Configuration
AUTHORIZER_URL="http://localhost:4000"
STACKFORGE_URL="http://localhost:4040"

# Login credentials
LOGIN_DATA='{"application":"STACKFORGE","email":"maftuhzaidi93@gmail.com","password":"Admin123"}'

echo -e "${YELLOW}Step 1: Login to Authorizer Service${NC}"
echo "URL: $AUTHORIZER_URL/authorizer/v1/auth/login"
echo ""

# Login and get token
RESPONSE=$(curl -s -X POST "$AUTHORIZER_URL/authorizer/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "$LOGIN_DATA")

# Check if login was successful
if echo "$RESPONSE" | grep -q "error"; then
    echo -e "${RED}✗ Login failed${NC}"
    echo "Response: $RESPONSE"
    exit 1
fi

# Extract token from response
TOKEN=$(echo "$RESPONSE" | jq -r '.data.access_token.token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo -e "${RED}✗ Failed to extract token${NC}"
    echo "Response: $RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ Login successful${NC}"
echo "Token (first 50 chars): ${TOKEN:0:50}..."
echo ""

echo -e "${YELLOW}Step 2: Decode JWT Header to Verify KID${NC}"
# JWT header is the first part before the first dot
HEADER=$(echo "$TOKEN" | cut -d. -f1)

# Decode base64 (add padding if needed)
HEADER_PADDED="$HEADER"
while [ $((${#HEADER_PADDED} % 4)) -ne 0 ]; do
    HEADER_PADDED="${HEADER_PADDED}="
done

DECODED_HEADER=$(echo "$HEADER_PADDED" | base64 -d 2>/dev/null || echo "$HEADER_PADDED" | base64 -D 2>/dev/null)

echo "Decoded Header:"
echo "$DECODED_HEADER" | jq '.' 2>/dev/null || echo "$DECODED_HEADER"
echo ""

# Check if kid exists in header
if echo "$DECODED_HEADER" | grep -q "kid"; then
    KID=$(echo "$DECODED_HEADER" | jq -r '.kid' 2>/dev/null)
    echo -e "${GREEN}✓ KID found in token header: $KID${NC}"
else
    echo -e "${RED}✗ KID not found in token header${NC}"
    exit 1
fi
echo ""

echo -e "${YELLOW}Step 3: Verify JWKS Endpoint Returns Same KID${NC}"
JWKS_RESPONSE=$(curl -s "$AUTHORIZER_URL/.well-known/jwks.json")
JWKS_KID=$(echo "$JWKS_RESPONSE" | jq -r '.keys[0].kid')

echo "JWKS KID: $JWKS_KID"

if [ "$KID" = "$JWKS_KID" ]; then
    echo -e "${GREEN}✓ Token KID matches JWKS KID${NC}"
else
    echo -e "${RED}✗ Token KID does not match JWKS KID${NC}"
    echo "Token KID: $KID"
    echo "JWKS KID: $JWKS_KID"
    exit 1
fi
echo ""

echo -e "${YELLOW}Step 4: Test Token with Stackforge Service${NC}"
echo "Testing: GET $STACKFORGE_URL/api/todos"
echo ""

STACKFORGE_RESPONSE=$(curl -s -w "\n%{http_code}" \
  -H "Authorization: Bearer $TOKEN" \
  "$STACKFORGE_URL/api/todos")

HTTP_CODE=$(echo "$STACKFORGE_RESPONSE" | tail -n1)
BODY=$(echo "$STACKFORGE_RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ Successfully accessed protected endpoint${NC}"
    echo "HTTP Status: $HTTP_CODE"
    echo "Response:"
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
else
    echo -e "${RED}✗ Failed to access protected endpoint${NC}"
    echo "HTTP Status: $HTTP_CODE"
    echo "Response:"
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
    exit 1
fi
echo ""

echo -e "${GREEN}=== All Tests Passed! ===${NC}"
echo ""
echo "Summary:"
echo "✓ Login successful"
echo "✓ Token contains kid in header"
echo "✓ Token kid matches JWKS kid"
echo "✓ Token validated successfully by Stackforge service"
echo ""

#!/bin/bash

# SSO Authentication Integration - Manual Test Script
# This script tests the complete SSO flow and RBAC enforcement

set -e

echo "=== SSO Authentication Integration - Manual Test ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
STACKFORGE_URL="http://localhost:4040"
AUTHORIZER_URL="http://authorizer.localdev.me"

echo -e "${YELLOW}Prerequisites:${NC}"
echo "1. Authorizer service should be running on port 4000"
echo "2. Stackforge service should be running on port 4040"
echo "3. You should have a valid user account in the Authorizer service"
echo ""
read -p "Press Enter to continue or Ctrl+C to abort..."
echo ""

# Test 1: Health check (public endpoint)
echo -e "${YELLOW}Test 1: Health Check (Public Endpoint)${NC}"
echo "Testing: GET $STACKFORGE_URL/health"
RESPONSE=$(curl -s -w "\n%{http_code}" "$STACKFORGE_URL/health")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ Health check passed${NC}"
    echo "Response: $BODY"
else
    echo -e "${RED}✗ Health check failed (HTTP $HTTP_CODE)${NC}"
    echo "Response: $BODY"
fi
echo ""

# Test 2: Access protected endpoint without token
echo -e "${YELLOW}Test 2: Access Protected Endpoint Without Token${NC}"
echo "Testing: GET $STACKFORGE_URL/api/todos"
RESPONSE=$(curl -s -w "\n%{http_code}" "$STACKFORGE_URL/api/todos")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "401" ]; then
    echo -e "${GREEN}✓ Correctly rejected unauthenticated request${NC}"
    echo "Response: $BODY"
else
    echo -e "${RED}✗ Expected 401, got HTTP $HTTP_CODE${NC}"
    echo "Response: $BODY"
fi
echo ""

# Test 3: SSO Login Flow
echo -e "${YELLOW}Test 3: SSO Login Flow${NC}"
echo "To test the complete SSO flow:"
echo "1. Open your browser and navigate to: $STACKFORGE_URL/auth/login"
echo "2. You should be redirected to the Authorizer login page"
echo "3. Log in with your credentials"
echo "4. After successful login, you should be redirected back to Stackforge"
echo "5. Check that the jwt_user_token cookie is set in your browser"
echo ""
echo "Manual verification steps:"
echo "  - Check browser DevTools > Application > Cookies"
echo "  - Verify 'jwt_user_token' cookie exists"
echo "  - Verify cookie has HttpOnly and Secure flags (in production)"
echo ""
read -p "Press Enter after completing the login flow..."
echo ""

# Test 4: Access protected endpoint with token (requires manual token input)
echo -e "${YELLOW}Test 4: Access Protected Endpoint With Token${NC}"
echo "Please provide a valid JWT token from your browser cookie:"
read -p "Token: " JWT_TOKEN
echo ""

if [ -n "$JWT_TOKEN" ]; then
    echo "Testing: GET $STACKFORGE_URL/api/todos with Authorization header"
    RESPONSE=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $JWT_TOKEN" "$STACKFORGE_URL/api/todos")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)
    
    if [ "$HTTP_CODE" = "200" ]; then
        echo -e "${GREEN}✓ Successfully accessed protected endpoint${NC}"
        echo "Response: $BODY"
    else
        echo -e "${RED}✗ Failed to access protected endpoint (HTTP $HTTP_CODE)${NC}"
        echo "Response: $BODY"
    fi
    echo ""
    
    # Test 5: Test with expired token
    echo -e "${YELLOW}Test 5: Test With Expired Token${NC}"
    echo "To test expired token handling, wait for the token to expire or use an expired token"
    echo "Skipping for now..."
    echo ""
    
    # Test 6: RBAC - Test role-based access
    echo -e "${YELLOW}Test 6: RBAC - Role-Based Access Control${NC}"
    echo "Testing endpoints with role requirements..."
    echo "Note: This requires endpoints with RequireRole middleware configured"
    echo "Check the main.go file for examples of protected endpoints"
    echo ""
    
    # Test 7: RBAC - Test permission-based access
    echo -e "${YELLOW}Test 7: RBAC - Permission-Based Access Control${NC}"
    echo "Testing endpoints with permission requirements..."
    echo "Note: This requires endpoints with RequirePermission middleware configured"
    echo "Check the main.go file for examples of protected endpoints"
    echo ""
else
    echo -e "${YELLOW}Skipping token-based tests (no token provided)${NC}"
    echo ""
fi

# Test 8: Logout flow
echo -e "${YELLOW}Test 8: Logout Flow${NC}"
echo "To test the logout flow:"
echo "1. Navigate to: $STACKFORGE_URL/auth/logout"
echo "2. You should be redirected to the Authorizer logout page"
echo "3. Check that the jwt_user_token cookie is cleared"
echo "4. Try accessing a protected endpoint - should get 401"
echo ""
read -p "Press Enter after testing logout..."
echo ""

# Test 9: JWKS caching verification
echo -e "${YELLOW}Test 9: JWKS Caching Verification${NC}"
echo "To verify JWKS caching:"
echo "1. Check the application logs for JWKS fetch operations"
echo "2. Make multiple authenticated requests"
echo "3. Verify that JWKS is fetched only once and then cached"
echo "4. Look for log messages like:"
echo "   - 'JWKS fetch started'"
echo "   - 'JWKS keys cached successfully'"
echo "   - 'JWKS cache hit'"
echo ""
echo "Check your application logs now..."
echo ""

# Summary
echo -e "${YELLOW}=== Test Summary ===${NC}"
echo ""
echo "Manual verification checklist:"
echo "□ Health check endpoint is accessible without authentication"
echo "□ Protected endpoints require authentication"
echo "□ SSO login redirects to Authorizer service"
echo "□ Successful login sets HttpOnly cookie"
echo "□ Valid token allows access to protected endpoints"
echo "□ Invalid/expired tokens are rejected with 401"
echo "□ RBAC middleware enforces role requirements"
echo "□ RBAC middleware enforces permission requirements"
echo "□ Logout clears authentication cookie"
echo "□ JWKS keys are cached and reused"
echo ""
echo -e "${GREEN}Manual testing complete!${NC}"
echo ""
echo "For automated integration tests, see:"
echo "  - internal/auth/service/*_test.go"
echo "  - internal/middleware/*_test.go"
echo ""

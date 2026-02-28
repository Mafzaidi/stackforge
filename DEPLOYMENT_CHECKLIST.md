# SSO Authentication Deployment Checklist

This checklist ensures proper deployment of the SSO authentication integration for Stackforge Service in production environments.

## Required Environment Variables

### Authentication Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `AUTHORIZER_BASE_URL` | **Yes** | `http://authorizer.localdev.me` | Base URL of the Authorizer Service (must be HTTPS in production) |
| `APP_CODE` | No | `STACKFORGE` | Application identifier for JWT audience validation |
| `APP_CALLBACK_URL` | **Yes** | `http://localhost:4040/auth/callback` | Full callback URL for OAuth2-style flow (must be HTTPS in production) |
| `JWKS_CACHE_DURATION` | No | `3600` | JWKS public key cache duration in seconds (1 hour default) |

### Example Production Configuration

```bash
# Authorizer Service
AUTHORIZER_BASE_URL=https://auth.yourdomain.com

# Application Identity
APP_CODE=STACKFORGE
APP_CALLBACK_URL=https://stackforge.yourdomain.com/auth/callback

# JWKS Caching (optional)
JWKS_CACHE_DURATION=3600
```

## Security Requirements

### HTTPS Configuration

**CRITICAL**: Production deployments **MUST** use HTTPS for the following reasons:

1. **Secure Cookies**: Authentication cookies are marked with the `Secure` flag, which requires HTTPS
   - Without HTTPS, browsers will reject the authentication cookie
   - Users will be unable to maintain authenticated sessions

2. **Token Security**: JWT tokens contain sensitive user information
   - Tokens transmitted over HTTP can be intercepted
   - Man-in-the-middle attacks can steal authentication credentials

3. **OAuth2 Compliance**: The SSO callback flow follows OAuth2 best practices
   - OAuth2 specification requires HTTPS for production deployments
   - Authorizer Service may reject HTTP callback URLs

**Action Items**:
- [ ] Configure TLS/SSL certificates for your domain
- [ ] Ensure `APP_CALLBACK_URL` uses `https://` scheme
- [ ] Ensure `AUTHORIZER_BASE_URL` uses `https://` scheme
- [ ] Verify load balancer/reverse proxy terminates TLS correctly
- [ ] Test cookie setting works correctly over HTTPS

### Cookie Configuration

The authentication system uses cookies with the following security attributes:

- **HttpOnly**: `true` - Prevents JavaScript access (XSS protection)
- **Secure**: `true` - Requires HTTPS in production
- **SameSite**: `Lax` - CSRF protection while allowing OAuth2 redirects
- **Path**: `/` - Available to all application routes
- **Max-Age**: Set to token expiration time

## Service Dependencies

### JWKS Endpoint Availability

The application requires continuous access to the Authorizer Service's JWKS endpoint for JWT validation.

**Endpoint**: `{AUTHORIZER_BASE_URL}/.well-known/jwks.json`

**Requirements**:
- [ ] JWKS endpoint must be accessible from production environment
- [ ] Network firewall rules allow outbound HTTPS to Authorizer Service
- [ ] DNS resolution works for Authorizer Service domain
- [ ] JWKS endpoint returns valid JSON with RSA public keys

**Failure Impact**:
- If JWKS endpoint is unavailable, cached keys will be used (default 1 hour)
- After cache expiration, all authentication requests will fail with 500 errors
- Users will be unable to log in or access protected endpoints

**Testing**:
```bash
# Verify JWKS endpoint is accessible
curl https://auth.yourdomain.com/.well-known/jwks.json

# Expected response format:
# {
#   "keys": [
#     {
#       "kty": "RSA",
#       "use": "sig",
#       "alg": "RS256",
#       "kid": "...",
#       "n": "...",
#       "e": "AQAB"
#     }
#   ]
# }
```

### Network Connectivity

- [ ] Verify outbound HTTPS connectivity to Authorizer Service
- [ ] Configure appropriate timeout values for JWKS requests (default: 10s)
- [ ] Ensure retry logic is enabled (3 retries with exponential backoff)

## Monitoring and Observability

### Key Metrics to Monitor

1. **Authentication Success Rate**
   - Metric: `auth_attempts_total{status="success|failure"}`
   - Alert: Success rate < 95% over 5 minutes
   - Action: Check Authorizer Service availability and JWKS endpoint

2. **JWKS Fetch Latency**
   - Metric: `jwks_fetch_duration_seconds`
   - Alert: P95 latency > 2 seconds
   - Action: Investigate network connectivity or Authorizer Service performance

3. **JWKS Fetch Failures**
   - Metric: `jwks_fetch_errors_total`
   - Alert: Any failures in 5-minute window
   - Action: Check Authorizer Service health and network connectivity

4. **Token Validation Errors**
   - Metric: `token_validation_errors_total{reason="expired|invalid_signature|invalid_issuer|invalid_audience"}`
   - Alert: Spike in validation errors (> 10% of requests)
   - Action: Investigate token issuance or clock synchronization issues

5. **RBAC Authorization Failures**
   - Metric: `rbac_authorization_failures_total{reason="missing_role|missing_permission"}`
   - Alert: Unusual spike in authorization failures
   - Action: Review user permissions and role assignments

### Logging Recommendations

**Required Log Events**:
- Authentication attempts (success/failure) with user ID and IP address
- RBAC authorization decisions (allow/deny) with required role/permission
- JWKS fetch operations (success/failure) with latency
- Token validation errors with error type and request ID

**Log Format**:
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "event": "auth_success",
  "user_id": "user_123",
  "username": "john",
  "ip_address": "192.168.1.100",
  "request_id": "req_abc123",
  "endpoint": "/api/todos"
}
```

**Security Considerations**:
- **NEVER** log full JWT tokens
- Only log token ID (last 4 characters) for debugging
- Redact sensitive claims (email, personal information) in logs
- Use structured logging for easy parsing and analysis

### Health Checks

- [ ] Configure health check endpoint: `GET /health`
- [ ] Health check should remain **public** (no authentication required)
- [ ] Health check should verify:
  - Application is running
  - Database connectivity (if applicable)
  - JWKS endpoint accessibility (optional, may cause false positives)

### Alerting Rules

**Critical Alerts** (Page on-call):
- Authentication success rate < 90% for 5 minutes
- JWKS endpoint unreachable for 3 consecutive attempts
- Application unable to start due to configuration errors

**Warning Alerts** (Notify team):
- Authentication success rate < 95% for 10 minutes
- JWKS fetch latency P95 > 2 seconds for 10 minutes
- Spike in token validation errors (> 5% of requests)

## Pre-Deployment Checklist

### Configuration Validation

- [ ] All required environment variables are set
- [ ] `AUTHORIZER_BASE_URL` uses HTTPS scheme
- [ ] `APP_CALLBACK_URL` uses HTTPS scheme and matches deployed domain
- [ ] `APP_CODE` matches the application identifier in Authorizer Service
- [ ] JWKS endpoint is accessible from production environment

### Security Validation

- [ ] TLS/SSL certificates are valid and not expired
- [ ] HTTPS is enforced (HTTP redirects to HTTPS)
- [ ] Secure cookies are enabled in production
- [ ] CORS configuration allows Authorizer Service domain (if applicable)
- [ ] Rate limiting is configured for auth endpoints

### Testing Validation

- [ ] All unit tests pass (`go test ./...`)
- [ ] All property-based tests pass
- [ ] Integration tests pass in staging environment
- [ ] Manual SSO login flow tested in staging
- [ ] Token expiration and refresh tested
- [ ] RBAC enforcement tested for protected endpoints

### Monitoring Validation

- [ ] Logging is configured and working
- [ ] Metrics are being collected
- [ ] Alerts are configured and tested
- [ ] Dashboards are created for auth metrics
- [ ] On-call team has access to logs and metrics

## Post-Deployment Validation

### Smoke Tests

Run these tests immediately after deployment:

1. **Health Check**
   ```bash
   curl https://stackforge.yourdomain.com/health
   # Expected: 200 OK
   ```

2. **Public Endpoint Access**
   ```bash
   curl https://stackforge.yourdomain.com/health
   # Expected: 200 OK (no authentication required)
   ```

3. **Protected Endpoint Without Auth**
   ```bash
   curl https://stackforge.yourdomain.com/api/todos
   # Expected: 302 Redirect to Authorizer login page
   ```

4. **SSO Login Flow**
   - Open browser to `https://stackforge.yourdomain.com/api/todos`
   - Verify redirect to Authorizer Service login page
   - Complete authentication
   - Verify redirect back to application
   - Verify authentication cookie is set
   - Verify access to protected endpoint works

5. **JWKS Endpoint Connectivity**
   ```bash
   curl https://auth.yourdomain.com/.well-known/jwks.json
   # Expected: 200 OK with valid JWKS response
   ```

### Monitoring Validation

- [ ] Check logs for authentication events
- [ ] Verify metrics are being reported
- [ ] Confirm no error spikes in dashboards
- [ ] Test alert notifications (if possible)

## Rollback Plan

If issues are detected after deployment:

1. **Immediate Rollback**
   - Revert to previous application version
   - Verify previous version is working correctly
   - Investigate issues in staging environment

2. **Partial Rollback** (if only auth is broken)
   - Remove auth middleware from protected routes temporarily
   - Keep application running without authentication
   - Fix issues and redeploy

3. **Configuration Rollback**
   - Revert environment variable changes
   - Restart application with previous configuration
   - Verify functionality is restored

## Troubleshooting Guide

### Issue: Users cannot log in (cookie not set)

**Symptoms**: Users redirected to login repeatedly, cookie not visible in browser

**Possible Causes**:
- Application not using HTTPS
- `APP_CALLBACK_URL` doesn't match actual domain
- Browser blocking third-party cookies

**Resolution**:
1. Verify HTTPS is enabled: `curl -I https://stackforge.yourdomain.com`
2. Check `APP_CALLBACK_URL` matches deployed URL
3. Verify Secure cookie flag is appropriate for environment
4. Check browser console for cookie errors

### Issue: Authentication fails with "Authentication service unavailable"

**Symptoms**: 500 errors on protected endpoints, logs show JWKS fetch failures

**Possible Causes**:
- JWKS endpoint is down
- Network connectivity issues
- Firewall blocking outbound HTTPS

**Resolution**:
1. Test JWKS endpoint: `curl https://auth.yourdomain.com/.well-known/jwks.json`
2. Check network connectivity from production environment
3. Verify firewall rules allow outbound HTTPS
4. Check Authorizer Service health

### Issue: Token validation fails with "Invalid token signature"

**Symptoms**: 401 errors, logs show signature verification failures

**Possible Causes**:
- Clock skew between services
- Key rotation in progress
- Wrong JWKS endpoint configured

**Resolution**:
1. Verify `AUTHORIZER_BASE_URL` is correct
2. Check system clock synchronization (NTP)
3. Force JWKS cache refresh (restart application)
4. Verify Authorizer Service is issuing valid tokens

### Issue: RBAC authorization fails unexpectedly

**Symptoms**: 403 errors for users who should have access

**Possible Causes**:
- User roles not configured in Authorizer Service
- `APP_CODE` mismatch between services
- Authorization claims missing from token

**Resolution**:
1. Verify user has correct roles in Authorizer Service
2. Check `APP_CODE` matches between services
3. Decode JWT token and verify authorization claims
4. Check RBAC middleware configuration

## Additional Resources

- [SSO Authentication Requirements](/.kiro/specs/sso-authentication-integration/requirements.md)
- [SSO Authentication Design](/.kiro/specs/sso-authentication-integration/design.md)
- [README - Authentication Setup](README.md#authentication-setup)
- Authorizer Service Documentation: Contact your platform team

## Support Contacts

- **Platform Team**: [Contact information]
- **Security Team**: [Contact information]
- **On-Call Engineer**: [Pager/contact information]

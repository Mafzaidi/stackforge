package handler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/delivery/http/middleware"
	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/service"
	"github.com/mafzaidi/stackforge/internal/pkg/response"
	"github.com/mafzaidi/stackforge/internal/usecase/auth"
)

const (
	authCookieName = "jwt_user_token"
	cookiePath     = "/"
)

// AuthHandler handles authentication HTTP requests
// It delegates business logic to use cases and focuses on HTTP concerns
type AuthHandler struct {
	loginUC    auth.LoginUseCase
	callbackUC auth.CallbackUseCase
	logoutUC   auth.LogoutUseCase
	log        service.Logger
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(
	loginUC auth.LoginUseCase,
	callbackUC auth.CallbackUseCase,
	logoutUC auth.LogoutUseCase,
	log service.Logger,
) *AuthHandler {
	return &AuthHandler{
		loginUC:    loginUC,
		callbackUC: callbackUC,
		logoutUC:   logoutUC,
		log:        log,
	}
}

// Login redirects to Authorizer login page
// GET /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	h.log.Info("SSO login initiated", service.Fields{
		"request_id": requestID,
		"ip":         c.ClientIP(),
	})

	// Delegate to use case
	loginURL, err := h.loginUC.BuildLoginURL()
	if err != nil {
		h.log.Error("SSO login failed: invalid authorizer URL", service.Fields{
			"request_id": requestID,
			"ip":         c.ClientIP(),
			"error":      err.Error(),
		})

		response.InternalServerError(c, "Authentication service configuration error")
		return
	}

	h.log.Info("SSO login redirect", service.Fields{
		"request_id":   requestID,
		"ip":           c.ClientIP(),
		"redirect_url": loginURL,
	})

	// Redirect to Authorizer
	c.Redirect(302, loginURL)
}

// Callback handles OAuth2-style callback from Authorizer
// GET /auth/callback?token=<jwt>
func (h *AuthHandler) Callback(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	// Extract token from query parameter
	token := c.Query("token")
	if token == "" {
		h.log.Warn("SSO callback failed: missing token", service.Fields{
			"request_id": requestID,
			"ip":         c.ClientIP(),
		})

		response.BadRequest(c, "Missing authentication token")
		return
	}

	// Delegate to use case for token validation
	claims, err := h.callbackUC.ValidateToken(token)
	if err != nil {
		h.log.Warn("SSO callback failed: invalid token", service.Fields{
			"request_id": requestID,
			"ip":         c.ClientIP(),
			"error":      err.Error(),
			"token":      truncateToken(token),
		})

		// Use generic message to avoid exposing internal details
		response.Unauthorized(c, "Invalid authentication token")
		return
	}

	// Token is valid, set the authentication cookie
	h.setAuthCookie(c, token, claims.ExpiresAt)

	h.log.Info("SSO callback successful", service.Fields{
		"request_id": requestID,
		"ip":         c.ClientIP(),
		"user_id":    claims.Subject,
		"username":   claims.Username,
		"email":      claims.Email,
		"token":      truncateToken(token),
	})

	// Redirect to dashboard
	c.Redirect(302, "/dashboard")
}

// Logout clears cookie and redirects to Authorizer logout
// GET /auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	// Try to get user info from context for logging (if available)
	userID := "unknown"
	username := "unknown"
	if claims, err := getClaimsFromContext(c); err == nil {
		userID = claims.Subject
		username = claims.Username
	}

	h.log.Info("SSO logout initiated", service.Fields{
		"request_id": requestID,
		"ip":         c.ClientIP(),
		"user_id":    userID,
		"username":   username,
	})

	// Clear the authentication cookie
	h.clearAuthCookie(c)

	// Delegate to use case for logout URL
	logoutURL, err := h.logoutUC.BuildLogoutURL()
	if err != nil {
		h.log.Error("SSO logout failed: invalid authorizer URL", service.Fields{
			"request_id": requestID,
			"ip":         c.ClientIP(),
			"user_id":    userID,
			"username":   username,
			"error":      err.Error(),
		})

		response.InternalServerError(c, "Authentication service configuration error")
		return
	}

	h.log.Info("SSO logout successful", service.Fields{
		"request_id":   requestID,
		"ip":           c.ClientIP(),
		"user_id":      userID,
		"username":     username,
		"redirect_url": logoutURL,
	})

	// Redirect to Authorizer logout
	c.Redirect(302, logoutURL)
}

// setAuthCookie sets HttpOnly + Secure cookie with JWT
func (h *AuthHandler) setAuthCookie(c *gin.Context, token string, expiresAt int64) {
	// Calculate max age from token expiration
	maxAge := int(time.Until(time.Unix(expiresAt, 0)).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}

	// Determine if we should use Secure flag (true for HTTPS)
	secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"

	c.SetCookie(
		authCookieName,
		token,
		maxAge,
		cookiePath,
		"",     // domain (empty = current domain)
		secure, // secure (HTTPS only)
		true,   // httpOnly
	)
}

// clearAuthCookie removes authentication cookie
func (h *AuthHandler) clearAuthCookie(c *gin.Context) {
	c.SetCookie(
		authCookieName,
		"",
		-1, // negative max age deletes the cookie
		cookiePath,
		"",    // domain
		false, // secure
		true,  // httpOnly
	)
}

// getClaimsFromContext is a helper to get claims without panicking
func getClaimsFromContext(c *gin.Context) (*entity.Claims, error) {
	claimsInterface, exists := c.Get("auth_claims")
	if !exists {
		return nil, fmt.Errorf("claims not found in context")
	}

	claims, ok := claimsInterface.(*entity.Claims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type in context")
	}

	return claims, nil
}

// truncateToken returns the last 4 characters of a token for safe logging
// Returns empty string if token is too short
func truncateToken(token string) string {
	if len(token) < 4 {
		return ""
	}
	return "..." + token[len(token)-4:]
}

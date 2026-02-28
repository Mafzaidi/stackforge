package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/domain/service"
	"github.com/mafzaidi/stackforge/internal/pkg/response"
)

// JWTValidator defines the interface for JWT token validation
type JWTValidator interface {
	ValidateToken(tokenString string) (*entity.Claims, error)
}

// AuthMiddleware validates JWT tokens and populates user context
// It extracts tokens from Authorization header or cookie, validates them,
// and stores the claims in gin.Context for downstream handlers
func AuthMiddleware(jwtService JWTValidator, log service.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := GetRequestID(c)

		// Extract token from Authorization header or cookie
		token := extractToken(c)
		if token == "" {
			log.Warn("Authentication failed: missing token", service.Fields{
				"request_id": requestID,
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
				"ip":         c.ClientIP(),
			})

			response.Unauthorized(c, "Missing authentication token")
			c.Abort()
			return
		}

		// Validate token using JWT service
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			// Map specific errors to appropriate messages
			message := mapValidationError(err)

			log.Warn("Authentication failed: token validation error", service.Fields{
				"request_id": requestID,
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
				"ip":         c.ClientIP(),
				"error":      message,
				"token":      truncateToken(token),
			})

			response.Unauthorized(c, message)
			c.Abort()
			return
		}

		// Store claims in gin.Context
		SetClaims(c, claims)

		log.Info("Authentication successful", service.Fields{
			"request_id": requestID,
			"path":       c.Request.URL.Path,
			"method":     c.Request.Method,
			"user_id":    claims.Subject,
			"username":   claims.Username,
			"email":      claims.Email,
			"token":      truncateToken(token),
		})

		// Call next handler
		c.Next()
	}
}

// extractToken gets token from "Authorization: Bearer <token>" header or "auth_token" cookie
// Returns empty string if token is not found in either location
func extractToken(c *gin.Context) string {
	// Try Authorization header first
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		// Expected format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return strings.TrimSpace(parts[1])
		}
	}

	// Try cookie as fallback
	token, err := c.Cookie("auth_token")
	if err == nil && token != "" {
		return token
	}

	return ""
}

// mapValidationError maps JWT validation errors to user-friendly messages
// This ensures consistent error messages as specified in requirements 10.1-10.5
// and prevents exposure of internal implementation details
func mapValidationError(err error) string {
	errMsg := err.Error()
	fmt.Println(errMsg)
	// Check for specific error messages from JWT service
	// These messages are safe to expose to clients
	switch {
	case strings.Contains(errMsg, "Token expired"):
		return "Token expired"
	case strings.Contains(errMsg, "Invalid token issuer"):
		return "Invalid token issuer"
	case strings.Contains(errMsg, "Invalid token audience"):
		return "Invalid token audience"
	case strings.Contains(errMsg, "failed to get public key"):
		// Don't expose internal JWKS details
		return "Invalid token signature"
	case strings.Contains(errMsg, "unexpected signing method"):
		return "Invalid token signature"
	case strings.Contains(errMsg, "missing or invalid kid"):
		return "Invalid token signature"
	case strings.Contains(errMsg, "failed to fetch JWKS"):
		// Don't expose internal service details
		return "Authentication service unavailable"
	case strings.Contains(errMsg, "JWKS endpoint"):
		// Don't expose internal service details
		return "Authentication service unavailable"
	default:
		// Generic message for any other validation errors
		return "Invalid token signature"
	}
}

// truncateToken returns the last 4 characters of a token for safe logging
// Returns empty string if token is too short
func truncateToken(token string) string {
	if len(token) < 4 {
		return ""
	}
	return "..." + token[len(token)-4:]
}

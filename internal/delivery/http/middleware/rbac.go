package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/domain/service"
	"github.com/mafzaidi/stackforge/internal/pkg/response"
)

// RequireRole creates middleware that enforces role requirement
// The middleware checks if the authenticated user has the specified role
// in either app-specific or GLOBAL authorization
func RequireRole(cfg service.Config, log service.Logger, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := GetRequestID(c)

		// Get claims from context (set by AuthMiddleware)
		claims, err := GetClaims(c)
		if err != nil {
			log.Warn("RBAC authorization failed: missing authentication", service.Fields{
				"request_id":    requestID,
				"path":          c.Request.URL.Path,
				"method":        c.Request.Method,
				"required_role": role,
			})

			response.Unauthorized(c, "Missing authentication")
			c.Abort()
			return
		}

		// Check if user has the required role
		if !claims.HasRole(cfg.GetAppCode(), role) {
			log.Warn("RBAC authorization failed: missing required role", service.Fields{
				"request_id":    requestID,
				"path":          c.Request.URL.Path,
				"method":        c.Request.Method,
				"user_id":       claims.Subject,
				"username":      claims.Username,
				"required_role": role,
			})

			response.Forbidden(c, "Insufficient permissions: missing required role")
			c.Abort()
			return
		}

		log.Info("RBAC authorization successful: role check passed", service.Fields{
			"request_id":    requestID,
			"path":          c.Request.URL.Path,
			"method":        c.Request.Method,
			"user_id":       claims.Subject,
			"username":      claims.Username,
			"required_role": role,
		})

		c.Next()
	}
}

// RequirePermission creates middleware that enforces permission requirement
// The middleware checks if the authenticated user has the specified permission
// or wildcard permission (*) in either app-specific or GLOBAL authorization
func RequirePermission(cfg service.Config, log service.Logger, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := GetRequestID(c)

		// Get claims from context (set by AuthMiddleware)
		claims, err := GetClaims(c)
		if err != nil {
			log.Warn("RBAC authorization failed: missing authentication", service.Fields{
				"request_id":          requestID,
				"path":                c.Request.URL.Path,
				"method":              c.Request.Method,
				"required_permission": permission,
			})

			response.Unauthorized(c, "Missing authentication")
			c.Abort()
			return
		}

		// Check if user has the required permission
		if !claims.HasPermission(cfg.GetAppCode(), permission) {
			log.Warn("RBAC authorization failed: missing required permission", service.Fields{
				"request_id":          requestID,
				"path":                c.Request.URL.Path,
				"method":              c.Request.Method,
				"user_id":             claims.Subject,
				"username":            claims.Username,
				"required_permission": permission,
			})

			response.Forbidden(c, "Insufficient permissions: missing required permission")
			c.Abort()
			return
		}

		log.Info("RBAC authorization successful: permission check passed", service.Fields{
			"request_id":          requestID,
			"path":                c.Request.URL.Path,
			"method":              c.Request.Method,
			"user_id":             claims.Subject,
			"username":            claims.Username,
			"required_permission": permission,
		})

		c.Next()
	}
}

// RequireAnyRole creates middleware that accepts any of the specified roles
// The middleware checks if the authenticated user has at least one of the
// specified roles in either app-specific or GLOBAL authorization
func RequireAnyRole(cfg service.Config, log service.Logger, roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := GetRequestID(c)

		// Get claims from context (set by AuthMiddleware)
		claims, err := GetClaims(c)
		if err != nil {
			log.Warn("RBAC authorization failed: missing authentication", service.Fields{
				"request_id":     requestID,
				"path":           c.Request.URL.Path,
				"method":         c.Request.Method,
				"required_roles": roles,
			})

			response.Unauthorized(c, "Missing authentication")
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		hasAnyRole := false
		for _, role := range roles {
			if claims.HasRole(cfg.GetAppCode(), role) {
				hasAnyRole = true
				break
			}
		}

		if !hasAnyRole {
			log.Warn("RBAC authorization failed: missing any required role", service.Fields{
				"request_id":     requestID,
				"path":           c.Request.URL.Path,
				"method":         c.Request.Method,
				"user_id":        claims.Subject,
				"username":       claims.Username,
				"required_roles": roles,
			})

			response.Forbidden(c, "Insufficient permissions: missing required role")
			c.Abort()
			return
		}

		log.Info("RBAC authorization successful: any role check passed", service.Fields{
			"request_id":     requestID,
			"path":           c.Request.URL.Path,
			"method":         c.Request.Method,
			"user_id":        claims.Subject,
			"username":       claims.Username,
			"required_roles": roles,
		})

		c.Next()
	}
}

// RequireAnyPermission creates middleware that accepts any of the specified permissions
// The middleware checks if the authenticated user has at least one of the
// specified permissions or wildcard permission (*) in either app-specific or GLOBAL authorization
func RequireAnyPermission(cfg service.Config, log service.Logger, permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := GetRequestID(c)

		// Get claims from context (set by AuthMiddleware)
		claims, err := GetClaims(c)
		if err != nil {
			log.Warn("RBAC authorization failed: missing authentication", service.Fields{
				"request_id":           requestID,
				"path":                 c.Request.URL.Path,
				"method":               c.Request.Method,
				"required_permissions": permissions,
			})

			response.Unauthorized(c, "Missing authentication")
			c.Abort()
			return
		}

		// Check if user has any of the required permissions
		hasAnyPermission := false
		for _, permission := range permissions {
			if claims.HasPermission(cfg.GetAppCode(), permission) {
				hasAnyPermission = true
				break
			}
		}

		if !hasAnyPermission {
			log.Warn("RBAC authorization failed: missing any required permission", service.Fields{
				"request_id":           requestID,
				"path":                 c.Request.URL.Path,
				"method":               c.Request.Method,
				"user_id":              claims.Subject,
				"username":             claims.Username,
				"required_permissions": permissions,
			})

			response.Forbidden(c, "Insufficient permissions: missing required permission")
			c.Abort()
			return
		}

		log.Info("RBAC authorization successful: any permission check passed", service.Fields{
			"request_id":           requestID,
			"path":                 c.Request.URL.Path,
			"method":               c.Request.Method,
			"user_id":              claims.Subject,
			"username":             claims.Username,
			"required_permissions": permissions,
		})

		c.Next()
	}
}

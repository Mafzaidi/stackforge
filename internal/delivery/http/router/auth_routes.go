package router

import (
	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/delivery/http/handler"
)

// RegisterAuthRoutes registers public authentication routes
// These routes do not require authentication middleware
//
// Routes:
//   - GET /auth/login    - Redirects to SSO login page
//   - GET /auth/callback - Handles SSO callback with JWT token
//   - GET /auth/logout   - Clears auth cookie and redirects to SSO logout
func RegisterAuthRoutes(router *gin.Engine, authHandler *handler.AuthHandler) {
	auth := router.Group("/auth")
	{
		auth.GET("/login", authHandler.Login)
		auth.GET("/callback", authHandler.Callback)
		auth.GET("/logout", authHandler.Logout)
	}
}

// RegisterHealthRoutes registers public health check routes
// These routes do not require authentication middleware
//
// Routes:
//   - GET /health - Health check endpoint for container orchestration
func RegisterHealthRoutes(router *gin.Engine, healthHandler *handler.HealthHandler) {
	router.GET("/health", healthHandler.HealthCheck)
}

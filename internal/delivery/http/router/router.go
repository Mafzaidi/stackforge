package router

import (
	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/delivery/http/handler"
	"github.com/mafzaidi/stackforge/internal/delivery/http/middleware"
)

// RouterConfig holds all dependencies needed for route registration
type RouterConfig struct {
	AuthHandler       *handler.AuthHandler
	TodoHandler       *handler.TodoHandler
	CredentialHandler *handler.CredentialHandler
	HealthHandler     *handler.HealthHandler
	AuthMiddleware    gin.HandlerFunc
}

// Setup creates and configures the main router with all routes
// It applies global middleware and registers all route groups
func Setup(cfg *RouterConfig) *gin.Engine {
	router := gin.Default()

	// Apply global middleware
	// Request ID middleware must be first to ensure all requests have IDs
	router.Use(middleware.RequestIDMiddleware())

	// Register route groups
	RegisterAuthRoutes(router, cfg.AuthHandler)
	RegisterHealthRoutes(router, cfg.HealthHandler)
	RegisterAPIRoutes(router, cfg.TodoHandler, cfg.CredentialHandler, cfg.AuthMiddleware)

	return router
}

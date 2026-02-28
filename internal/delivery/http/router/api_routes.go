package router

import (
	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/delivery/http/handler"
)

// RegisterAPIRoutes registers protected API routes
// All routes in this group require authentication via AuthMiddleware
//
// The AuthMiddleware validates JWT tokens and populates user claims in the request context.
// Additional middleware (RBAC, permissions) can be added to specific routes as needed.
//
// Routes:
//   - GET /api/todos       - List all todos (requires authentication)
//   - GET /api/credentials - List all credentials (requires authentication)
//
// Example: Adding role-based access control to a route:
//   api.GET("/todos", middleware.RequireRole(cfg, logger, "admin"), todoHandler.List)
//
// Example: Adding permission-based access control to a route:
//   api.GET("/todos", middleware.RequirePermission(cfg, logger, "todo.read"), todoHandler.List)
func RegisterAPIRoutes(
	router *gin.Engine,
	todoHandler *handler.TodoHandler,
	credentialHandler *handler.CredentialHandler,
	authMiddleware gin.HandlerFunc,
) {
	// Protected API routes - all /api/* routes require authentication
	api := router.Group("/api")
	api.Use(authMiddleware)
	{
		// Todo endpoints
		api.GET("/todos", todoHandler.List)

		// Credential endpoints
		api.GET("/credentials", credentialHandler.List)
	}
}

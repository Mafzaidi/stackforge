package router

import (
	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/delivery/http/handler"
)

// RegisterAPIRoutes registers protected API routes under /stackforge/v1
// All routes in this group require authentication via AuthMiddleware
//
// The AuthMiddleware validates JWT tokens and populates user claims in the request context.
// Additional middleware (RBAC, permissions) can be added to specific routes as needed.
//
// Routes:
//   - GET /stackforge/v1/todos       - List all todos (requires authentication)
//   - GET /stackforge/v1/credentials - List all credentials (requires authentication)
//
// Example: Adding role-based access control to a route:
//
//	api.GET("/todos", middleware.RequireRole(cfg, logger, "admin"), todoHandler.List)
//
// Example: Adding permission-based access control to a route:
//
//	api.GET("/todos", middleware.RequirePermission(cfg, logger, "todo.read"), todoHandler.List)
func RegisterAPIRoutes(
	router *gin.Engine,
	masterDataHandler *handler.MasterDataHandler,
	UserProfilesHandler *handler.UserProfilesHandler,
	todoHandler *handler.TodoHandler,
	credentialHandler *handler.CredentialHandler,
	menuHandler *handler.MenuHandler,
	vaultHandler *handler.VaultHandler,
	authMiddleware gin.HandlerFunc,
) {
	// Protected API routes - all /stackforge/v1/* routes require authentication
	api := router.Group("/stackforge/v1")
	api.Use(authMiddleware)
	{
		// Menu endpoints
		api.GET("/menus", menuHandler.List)

		// Master data endpoints
		api.GET("/master-data", masterDataHandler.List)

		// Todo endpoints
		api.GET("/todos", todoHandler.List)

		// Credential endpoints
		api.GET("/credentials", credentialHandler.List)
		api.POST("/credentials", credentialHandler.Create)

		// User profile endpoints
		api.POST("/user-profiles", UserProfilesHandler.Create)
		api.GET("/user-profiles", UserProfilesHandler.GetByUserID)
		api.POST("/user-profiles/master-password", UserProfilesHandler.SetupMasterPassword)

		// Vault endpoints
		api.POST("/vaults", vaultHandler.Create)
		api.GET("/vaults", vaultHandler.List)

	}
}

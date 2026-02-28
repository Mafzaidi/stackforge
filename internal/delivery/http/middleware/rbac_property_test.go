package middleware

import (
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/infrastructure/config"
	"github.com/mafzaidi/stackforge/internal/infrastructure/logger"
)

// **Validates: Requirements 4.2, 4.3, 4.4, 4.5**
// Property 4: Role-Based Authorization
// For any user claims and required role, if the user has that role in either
// the app-specific authorization or GLOBAL authorization, the RBAC_Middleware
// should allow the request to proceed; otherwise it should return 403 Forbidden.
func TestProperty_RoleBasedAuthorization(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	cfg := &config.Config{AppCode: "STACKFORGE"}
	log := logger.New()

	properties.Property("User with required role in app-specific auth is allowed", prop.ForAll(
		func(requiredRole string, otherRoles []string, authorizations []rbacTestAuthorization) bool {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Build claims with required role in app-specific authorization
			claims := &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
			}

			// Add app-specific authorization with required role
			appRoles := append([]string{requiredRole}, otherRoles...)
			claims.Authorization = append(claims.Authorization, entity.Authorization{
				App:   "STACKFORGE",
				Roles: appRoles,
			})

			// Add other authorizations
			for _, auth := range authorizations {
				claims.Authorization = append(claims.Authorization, entity.Authorization{
					App:   auth.App,
					Roles: auth.Roles,
				})
			}

			SetClaims(c, claims)

			// Create and execute middleware
			middleware := RequireRole(cfg, log, requiredRole)
			middleware(c)

			// Should allow request (not aborted)
			if c.IsAborted() {
				t.Logf("Expected request to proceed, but it was aborted. Status: %d", w.Code)
				return false
			}

			return true
		},
		genRoleName(),
		genRoleSlice(),
		genRBACAuthorizationSlice(),
	))

	properties.Property("User with required role in GLOBAL auth is allowed", prop.ForAll(
		func(requiredRole string, otherRoles []string, authorizations []rbacTestAuthorization) bool {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Build claims with required role in GLOBAL authorization
			claims := &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
			}

			// Add GLOBAL authorization with required role
			globalRoles := append([]string{requiredRole}, otherRoles...)
			claims.Authorization = append(claims.Authorization, entity.Authorization{
				App:   "GLOBAL",
				Roles: globalRoles,
			})

			// Add other authorizations
			for _, auth := range authorizations {
				claims.Authorization = append(claims.Authorization, entity.Authorization{
					App:   auth.App,
					Roles: auth.Roles,
				})
			}

			SetClaims(c, claims)

			// Create and execute middleware
			middleware := RequireRole(cfg, log, requiredRole)
			middleware(c)

			// Should allow request (not aborted)
			if c.IsAborted() {
				t.Logf("Expected request to proceed, but it was aborted. Status: %d", w.Code)
				return false
			}

			return true
		},
		genRoleName(),
		genRoleSlice(),
		genRBACAuthorizationSlice(),
	))

	properties.Property("User without required role is forbidden", prop.ForAll(
		func(requiredRole string, userRoles []string, authorizations []rbacTestAuthorization) bool {
			// Ensure userRoles doesn't contain requiredRole
			filteredRoles := make([]string, 0)
			for _, role := range userRoles {
				if role != requiredRole {
					filteredRoles = append(filteredRoles, role)
				}
			}

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Build claims without required role
			claims := &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
			}

			// Add authorizations without required role
			for _, auth := range authorizations {
				// Filter out required role from authorization
				authRoles := make([]string, 0)
				for _, role := range auth.Roles {
					if role != requiredRole {
						authRoles = append(authRoles, role)
					}
				}
				claims.Authorization = append(claims.Authorization, entity.Authorization{
					App:   auth.App,
					Roles: authRoles,
				})
			}

			// Add app-specific authorization without required role
			if len(filteredRoles) > 0 {
				claims.Authorization = append(claims.Authorization, entity.Authorization{
					App:   "STACKFORGE",
					Roles: filteredRoles,
				})
			}

			SetClaims(c, claims)

			// Create and execute middleware
			middleware := RequireRole(cfg, log, requiredRole)
			middleware(c)

			// Should forbid request (aborted with 403)
			if !c.IsAborted() {
				t.Logf("Expected request to be aborted, but it proceeded")
				return false
			}

			if w.Code != 403 {
				t.Logf("Expected 403 Forbidden, got %d", w.Code)
				return false
			}

			return true
		},
		genRoleName(),
		genRoleSlice(),
		genRBACAuthorizationSlice(),
	))

	properties.Property("User with role in different app only is forbidden", prop.ForAll(
		func(requiredRole string, otherRoles []string) bool {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Build claims with required role only in OTHER-APP
			claims := &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
				Authorization: []entity.Authorization{
					{
						App:   "OTHER-APP",
						Roles: append([]string{requiredRole}, otherRoles...),
					},
				},
			}

			SetClaims(c, claims)

			// Create and execute middleware
			middleware := RequireRole(cfg, log, requiredRole)
			middleware(c)

			// Should forbid request (aborted with 403)
			if !c.IsAborted() {
				t.Logf("Expected request to be aborted, but it proceeded")
				return false
			}

			if w.Code != 403 {
				t.Logf("Expected 403 Forbidden, got %d", w.Code)
				return false
			}

			return true
		},
		genRoleName(),
		genRoleSlice(),
	))

	properties.Property("Request without claims is unauthorized", prop.ForAll(
		func(requiredRole string) bool {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Don't set claims in context

			// Create and execute middleware
			middleware := RequireRole(cfg, log, requiredRole)
			middleware(c)

			// Should return 401 Unauthorized
			if !c.IsAborted() {
				t.Logf("Expected request to be aborted, but it proceeded")
				return false
			}

			if w.Code != 401 {
				t.Logf("Expected 401 Unauthorized, got %d", w.Code)
				return false
			}

			return true
		},
		genRoleName(),
	))

	properties.Property("User with empty authorization array is forbidden", prop.ForAll(
		func(requiredRole string) bool {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Build claims with empty authorization
			claims := &entity.Claims{
				Subject:       "user123",
				Username:      "testuser",
				Authorization: []entity.Authorization{},
			}

			SetClaims(c, claims)

			// Create and execute middleware
			middleware := RequireRole(cfg, log, requiredRole)
			middleware(c)

			// Should forbid request (aborted with 403)
			if !c.IsAborted() {
				t.Logf("Expected request to be aborted, but it proceeded")
				return false
			}

			if w.Code != 403 {
				t.Logf("Expected 403 Forbidden, got %d", w.Code)
				return false
			}

			return true
		},
		genRoleName(),
	))

	properties.TestingRun(t)
}

// rbacTestAuthorization is a helper struct for generating authorization data
type rbacTestAuthorization struct {
	App   string
	Roles []string
}

// genRoleName generates realistic role names
func genRoleName() gopter.Gen {
	return gen.OneConstOf(
		"admin",
		"user",
		"moderator",
		"viewer",
		"editor",
		"manager",
		"developer",
		"analyst",
	)
}

// genRoleSlice generates slices of role names
func genRoleSlice() gopter.Gen {
	return gen.SliceOfN(3, genRoleName())
}

// genRBACAuthorizationSlice generates slices of authorization structures
func genRBACAuthorizationSlice() gopter.Gen {
	return gen.SliceOfN(2, gen.Struct(reflect.TypeOf(rbacTestAuthorization{}), map[string]gopter.Gen{
		"App": gen.OneConstOf(
			"OTHER-APP",
			"ANOTHER-SERVICE",
			"THIRD-APP",
		),
		"Roles": genRoleSlice(),
	}))
}

// **Validates: Requirements 5.2, 5.3, 5.4, 5.5, 5.6**
// Property 5: Permission-Based Authorization
// For any user claims and required permission, if the user has that exact permission
// or the wildcard permission "*" in either the app-specific authorization or GLOBAL
// authorization, the RBAC_Middleware should allow the request to proceed; otherwise
// it should return 403 Forbidden.
func TestProperty_PermissionBasedAuthorization(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	cfg := &config.Config{AppCode: "STACKFORGE"}
	log := logger.New()

	properties.Property("User with exact required permission in app-specific auth is allowed", prop.ForAll(
		func(requiredPermission string, otherPermissions []string, authorizations []permissionTestAuthorization) bool {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Build claims with required permission in app-specific authorization
			claims := &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
			}

			// Add app-specific authorization with required permission
			appPermissions := append([]string{requiredPermission}, otherPermissions...)
			claims.Authorization = append(claims.Authorization, entity.Authorization{
				App:         "STACKFORGE",
				Permissions: appPermissions,
			})

			// Add other authorizations
			for _, auth := range authorizations {
				claims.Authorization = append(claims.Authorization, entity.Authorization{
					App:         auth.App,
					Permissions: auth.Permissions,
				})
			}

			SetClaims(c, claims)

			// Create and execute middleware
			middleware := RequirePermission(cfg, log, requiredPermission)
			middleware(c)

			// Should allow request (not aborted)
			if c.IsAborted() {
				t.Logf("Expected request to proceed, but it was aborted. Status: %d", w.Code)
				return false
			}

			return true
		},
		genPermissionName(),
		genPermissionSlice(),
		genPermissionAuthorizationSlice(),
	))

	properties.Property("User with exact required permission in GLOBAL auth is allowed", prop.ForAll(
		func(requiredPermission string, otherPermissions []string, authorizations []permissionTestAuthorization) bool {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Build claims with required permission in GLOBAL authorization
			claims := &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
			}

			// Add GLOBAL authorization with required permission
			globalPermissions := append([]string{requiredPermission}, otherPermissions...)
			claims.Authorization = append(claims.Authorization, entity.Authorization{
				App:         "GLOBAL",
				Permissions: globalPermissions,
			})

			// Add other authorizations
			for _, auth := range authorizations {
				claims.Authorization = append(claims.Authorization, entity.Authorization{
					App:         auth.App,
					Permissions: auth.Permissions,
				})
			}

			SetClaims(c, claims)

			// Create and execute middleware
			middleware := RequirePermission(cfg, log, requiredPermission)
			middleware(c)

			// Should allow request (not aborted)
			if c.IsAborted() {
				t.Logf("Expected request to proceed, but it was aborted. Status: %d", w.Code)
				return false
			}

			return true
		},
		genPermissionName(),
		genPermissionSlice(),
		genPermissionAuthorizationSlice(),
	))

	properties.Property("User with wildcard permission in app-specific auth is allowed", prop.ForAll(
		func(requiredPermission string, otherPermissions []string, authorizations []permissionTestAuthorization) bool {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Build claims with wildcard permission in app-specific authorization
			claims := &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
			}

			// Add app-specific authorization with wildcard permission
			appPermissions := append([]string{"*"}, otherPermissions...)
			claims.Authorization = append(claims.Authorization, entity.Authorization{
				App:         "STACKFORGE",
				Permissions: appPermissions,
			})

			// Add other authorizations
			for _, auth := range authorizations {
				claims.Authorization = append(claims.Authorization, entity.Authorization{
					App:         auth.App,
					Permissions: auth.Permissions,
				})
			}

			SetClaims(c, claims)

			// Create and execute middleware
			middleware := RequirePermission(cfg, log, requiredPermission)
			middleware(c)

			// Should allow request (not aborted)
			if c.IsAborted() {
				t.Logf("Expected request to proceed with wildcard permission, but it was aborted. Status: %d", w.Code)
				return false
			}

			return true
		},
		genPermissionName(),
		genPermissionSlice(),
		genPermissionAuthorizationSlice(),
	))

	properties.Property("User with wildcard permission in GLOBAL auth is allowed", prop.ForAll(
		func(requiredPermission string, otherPermissions []string, authorizations []permissionTestAuthorization) bool {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Build claims with wildcard permission in GLOBAL authorization
			claims := &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
			}

			// Add GLOBAL authorization with wildcard permission
			globalPermissions := append([]string{"*"}, otherPermissions...)
			claims.Authorization = append(claims.Authorization, entity.Authorization{
				App:         "GLOBAL",
				Permissions: globalPermissions,
			})

			// Add other authorizations
			for _, auth := range authorizations {
				claims.Authorization = append(claims.Authorization, entity.Authorization{
					App:         auth.App,
					Permissions: auth.Permissions,
				})
			}

			SetClaims(c, claims)

			// Create and execute middleware
			middleware := RequirePermission(cfg, log, requiredPermission)
			middleware(c)

			// Should allow request (not aborted)
			if c.IsAborted() {
				t.Logf("Expected request to proceed with wildcard permission, but it was aborted. Status: %d", w.Code)
				return false
			}

			return true
		},
		genPermissionName(),
		genPermissionSlice(),
		genPermissionAuthorizationSlice(),
	))

	properties.Property("User without required permission is forbidden", prop.ForAll(
		func(requiredPermission string, userPermissions []string, authorizations []permissionTestAuthorization) bool {
			// Ensure userPermissions doesn't contain requiredPermission or wildcard
			filteredPermissions := make([]string, 0)
			for _, perm := range userPermissions {
				if perm != requiredPermission && perm != "*" {
					filteredPermissions = append(filteredPermissions, perm)
				}
			}

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Build claims without required permission
			claims := &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
			}

			// Add authorizations without required permission or wildcard
			for _, auth := range authorizations {
				// Filter out required permission and wildcard from authorization
				authPermissions := make([]string, 0)
				for _, perm := range auth.Permissions {
					if perm != requiredPermission && perm != "*" {
						authPermissions = append(authPermissions, perm)
					}
				}
				claims.Authorization = append(claims.Authorization, entity.Authorization{
					App:         auth.App,
					Permissions: authPermissions,
				})
			}

			// Add app-specific authorization without required permission
			if len(filteredPermissions) > 0 {
				claims.Authorization = append(claims.Authorization, entity.Authorization{
					App:         "STACKFORGE",
					Permissions: filteredPermissions,
				})
			}

			SetClaims(c, claims)

			// Create and execute middleware
			middleware := RequirePermission(cfg, log, requiredPermission)
			middleware(c)

			// Should forbid request (aborted with 403)
			if !c.IsAborted() {
				t.Logf("Expected request to be aborted, but it proceeded")
				return false
			}

			if w.Code != 403 {
				t.Logf("Expected 403 Forbidden, got %d", w.Code)
				return false
			}

			return true
		},
		genPermissionName(),
		genPermissionSlice(),
		genPermissionAuthorizationSlice(),
	))

	properties.Property("User with permission in different app only is forbidden", prop.ForAll(
		func(requiredPermission string, otherPermissions []string) bool {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Build claims with required permission only in OTHER-APP
			claims := &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
				Authorization: []entity.Authorization{
					{
						App:         "OTHER-APP",
						Permissions: append([]string{requiredPermission}, otherPermissions...),
					},
				},
			}

			SetClaims(c, claims)

			// Create and execute middleware
			middleware := RequirePermission(cfg, log, requiredPermission)
			middleware(c)

			// Should forbid request (aborted with 403)
			if !c.IsAborted() {
				t.Logf("Expected request to be aborted, but it proceeded")
				return false
			}

			if w.Code != 403 {
				t.Logf("Expected 403 Forbidden, got %d", w.Code)
				return false
			}

			return true
		},
		genPermissionName(),
		genPermissionSlice(),
	))

	properties.Property("Request without claims is unauthorized", prop.ForAll(
		func(requiredPermission string) bool {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Don't set claims in context

			// Create and execute middleware
			middleware := RequirePermission(cfg, log, requiredPermission)
			middleware(c)

			// Should return 401 Unauthorized
			if !c.IsAborted() {
				t.Logf("Expected request to be aborted, but it proceeded")
				return false
			}

			if w.Code != 401 {
				t.Logf("Expected 401 Unauthorized, got %d", w.Code)
				return false
			}

			return true
		},
		genPermissionName(),
	))

	properties.Property("User with empty authorization array is forbidden", prop.ForAll(
		func(requiredPermission string) bool {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Build claims with empty authorization
			claims := &entity.Claims{
				Subject:       "user123",
				Username:      "testuser",
				Authorization: []entity.Authorization{},
			}

			SetClaims(c, claims)

			// Create and execute middleware
			middleware := RequirePermission(cfg, log, requiredPermission)
			middleware(c)

			// Should forbid request (aborted with 403)
			if !c.IsAborted() {
				t.Logf("Expected request to be aborted, but it proceeded")
				return false
			}

			if w.Code != 403 {
				t.Logf("Expected 403 Forbidden, got %d", w.Code)
				return false
			}

			return true
		},
		genPermissionName(),
	))

	properties.TestingRun(t)
}

// permissionTestAuthorization is a helper struct for generating authorization data with permissions
type permissionTestAuthorization struct {
	App         string
	Permissions []string
}

// genPermissionName generates realistic permission names
func genPermissionName() gopter.Gen {
	return gen.OneConstOf(
		"todo.read",
		"todo.write",
		"todo.delete",
		"user.read",
		"user.write",
		"admin.access",
		"report.generate",
		"settings.modify",
	)
}

// genPermissionSlice generates slices of permission names
func genPermissionSlice() gopter.Gen {
	return gen.SliceOfN(3, genPermissionName())
}

// genPermissionAuthorizationSlice generates slices of authorization structures with permissions
func genPermissionAuthorizationSlice() gopter.Gen {
	return gen.SliceOfN(2, gen.Struct(reflect.TypeOf(permissionTestAuthorization{}), map[string]gopter.Gen{
		"App": gen.OneConstOf(
			"OTHER-APP",
			"ANOTHER-SERVICE",
			"THIRD-APP",
		),
		"Permissions": genPermissionSlice(),
	}))
}

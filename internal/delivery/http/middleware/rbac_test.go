package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/infrastructure/config"
	"github.com/mafzaidi/stackforge/internal/infrastructure/logger"
	"github.com/mafzaidi/stackforge/internal/pkg/response"
	"github.com/stretchr/testify/assert"
)

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{AppCode: "STACKFORGE"}
	log := logger.New()

	tests := []struct {
		name           string
		requiredRole   string
		claims         *entity.Claims
		expectedStatus int
		expectedError  string
	}{
		{
			name:         "user has required role in app-specific auth",
			requiredRole: "admin",
			claims: &entity.Claims{
				Authorization: []entity.Authorization{
					{
						App:   "STACKFORGE",
						Roles: []string{"admin", "user"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "user has required role in GLOBAL auth",
			requiredRole: "admin",
			claims: &entity.Claims{
				Authorization: []entity.Authorization{
					{
						App:   "GLOBAL",
						Roles: []string{"admin"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "user has role in both app and GLOBAL",
			requiredRole: "user",
			claims: &entity.Claims{
				Authorization: []entity.Authorization{
					{
						App:   "STACKFORGE",
						Roles: []string{"user"},
					},
					{
						App:   "GLOBAL",
						Roles: []string{"admin"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "user does not have required role",
			requiredRole: "admin",
			claims: &entity.Claims{
				Authorization: []entity.Authorization{
					{
						App:   "STACKFORGE",
						Roles: []string{"user"},
					},
				},
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions: missing required role",
		},
		{
			name:         "user has role in different app only",
			requiredRole: "admin",
			claims: &entity.Claims{
				Authorization: []entity.Authorization{
					{
						App:   "OTHER-APP",
						Roles: []string{"admin"},
					},
				},
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions: missing required role",
		},
		{
			name:           "no claims in context",
			requiredRole:   "admin",
			claims:         nil,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Missing authentication",
		},
		{
			name:         "empty authorization array",
			requiredRole: "admin",
			claims: &entity.Claims{
				Authorization: []entity.Authorization{},
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions: missing required role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Set claims in context if provided
			if tt.claims != nil {
				SetClaims(c, tt.claims)
			}

			// Create middleware and execute
			middleware := RequireRole(cfg, log, tt.requiredRole)
			middleware(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRequirePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{AppCode: "STACKFORGE"}
	log := logger.New()

	tests := []struct {
		name               string
		requiredPermission string
		claims             *entity.Claims
		expectedStatus     int
		expectedError      string
	}{
		{
			name:               "user has exact permission in app-specific auth",
			requiredPermission: "todo.write",
			claims: &entity.Claims{
				Authorization: []entity.Authorization{
					{
						App:         "STACKFORGE",
						Permissions: []string{"todo.read", "todo.write"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:               "user has wildcard permission in app-specific auth",
			requiredPermission: "todo.write",
			claims: &entity.Claims{
				Authorization: []entity.Authorization{
					{
						App:         "STACKFORGE",
						Permissions: []string{"*"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:               "user has permission in GLOBAL auth",
			requiredPermission: "todo.write",
			claims: &entity.Claims{
				Authorization: []entity.Authorization{
					{
						App:         "GLOBAL",
						Permissions: []string{"todo.write"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:               "user has wildcard in GLOBAL auth",
			requiredPermission: "todo.write",
			claims: &entity.Claims{
				Authorization: []entity.Authorization{
					{
						App:         "GLOBAL",
						Permissions: []string{"*"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:               "user does not have required permission",
			requiredPermission: "todo.delete",
			claims: &entity.Claims{
				Authorization: []entity.Authorization{
					{
						App:         "STACKFORGE",
						Permissions: []string{"todo.read", "todo.write"},
					},
				},
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions: missing required permission",
		},
		{
			name:               "user has permission in different app only",
			requiredPermission: "todo.write",
			claims: &entity.Claims{
				Authorization: []entity.Authorization{
					{
						App:         "OTHER-APP",
						Permissions: []string{"todo.write"},
					},
				},
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions: missing required permission",
		},
		{
			name:               "no claims in context",
			requiredPermission: "todo.write",
			claims:             nil,
			expectedStatus:     http.StatusUnauthorized,
			expectedError:      "Missing authentication",
		},
		{
			name:               "empty authorization array",
			requiredPermission: "todo.write",
			claims: &entity.Claims{
				Authorization: []entity.Authorization{},
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions: missing required permission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Set claims in context if provided
			if tt.claims != nil {
				SetClaims(c, tt.claims)
			}

			// Create middleware and execute
			middleware := RequirePermission(cfg, log, tt.requiredPermission)
			middleware(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRequireAnyRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{AppCode: "STACKFORGE"}
	log := logger.New()

	tests := []struct {
		name           string
		requiredRoles  []string
		claims         *entity.Claims
		expectedStatus int
		expectedError  string
	}{
		{
			name:          "user has one of the required roles",
			requiredRoles: []string{"admin", "moderator"},
			claims: &entity.Claims{
				Authorization: []entity.Authorization{
					{
						App:   "STACKFORGE",
						Roles: []string{"moderator"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "user has multiple required roles",
			requiredRoles: []string{"admin", "moderator"},
			claims: &entity.Claims{
				Authorization: []entity.Authorization{
					{
						App:   "STACKFORGE",
						Roles: []string{"admin", "moderator", "user"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "user has role in GLOBAL auth",
			requiredRoles: []string{"admin", "moderator"},
			claims: &entity.Claims{
				Authorization: []entity.Authorization{
					{
						App:   "GLOBAL",
						Roles: []string{"admin"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "user does not have any required role",
			requiredRoles: []string{"admin", "moderator"},
			claims: &entity.Claims{
				Authorization: []entity.Authorization{
					{
						App:   "STACKFORGE",
						Roles: []string{"user"},
					},
				},
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions: missing required role",
		},
		{
			name:           "no claims in context",
			requiredRoles:  []string{"admin", "moderator"},
			claims:         nil,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Missing authentication",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Set claims in context if provided
			if tt.claims != nil {
				SetClaims(c, tt.claims)
			}

			// Create middleware and execute
			middleware := RequireAnyRole(cfg, log, tt.requiredRoles...)
			middleware(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestRequireAnyPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{AppCode: "STACKFORGE"}
	log := logger.New()

	tests := []struct {
		name                string
		requiredPermissions []string
		claims              *entity.Claims
		expectedStatus      int
		expectedError       string
	}{
		{
			name:                "user has one of the required permissions",
			requiredPermissions: []string{"todo.write", "todo.delete"},
			claims: &entity.Claims{
				Authorization: []entity.Authorization{
					{
						App:         "STACKFORGE",
						Permissions: []string{"todo.write"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:                "user has multiple required permissions",
			requiredPermissions: []string{"todo.write", "todo.delete"},
			claims: &entity.Claims{
				Authorization: []entity.Authorization{
					{
						App:         "STACKFORGE",
						Permissions: []string{"todo.read", "todo.write", "todo.delete"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:                "user has wildcard permission",
			requiredPermissions: []string{"todo.write", "todo.delete"},
			claims: &entity.Claims{
				Authorization: []entity.Authorization{
					{
						App:         "STACKFORGE",
						Permissions: []string{"*"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:                "user has permission in GLOBAL auth",
			requiredPermissions: []string{"todo.write", "todo.delete"},
			claims: &entity.Claims{
				Authorization: []entity.Authorization{
					{
						App:         "GLOBAL",
						Permissions: []string{"todo.delete"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:                "user does not have any required permission",
			requiredPermissions: []string{"todo.write", "todo.delete"},
			claims: &entity.Claims{
				Authorization: []entity.Authorization{
					{
						App:         "STACKFORGE",
						Permissions: []string{"todo.read"},
					},
				},
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions: missing required permission",
		},
		{
			name:                "no claims in context",
			requiredPermissions: []string{"todo.write", "todo.delete"},
			claims:              nil,
			expectedStatus:      http.StatusUnauthorized,
			expectedError:       "Missing authentication",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Set claims in context if provided
			if tt.claims != nil {
				SetClaims(c, tt.claims)
			}

			// Create middleware and execute
			middleware := RequireAnyPermission(cfg, log, tt.requiredPermissions...)
			middleware(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestRBACErrorResponseFormat verifies that RBAC middleware returns
// properly formatted error responses according to the ErrorResponse structure
func TestRBACErrorResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{AppCode: "STACKFORGE"}
	log := logger.New()

	tests := []struct {
		name               string
		middleware         gin.HandlerFunc
		claims             *entity.Claims
		expectedStatus     int
		expectedError      string
		expectedMessage    string
		expectedStatusCode int
	}{
		{
			name:               "RequireRole - missing claims returns 401 with proper format",
			middleware:         RequireRole(cfg, log, "admin"),
			claims:             nil,
			expectedStatus:     http.StatusUnauthorized,
			expectedError:      "Unauthorized",
			expectedMessage:    "Missing authentication",
			expectedStatusCode: 401,
		},
		{
			name:       "RequireRole - missing role returns 403 with proper format",
			middleware: RequireRole(cfg, log, "admin"),
			claims: &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
				Authorization: []entity.Authorization{
					{
						App:   "STACKFORGE",
						Roles: []string{"user"},
					},
				},
			},
			expectedStatus:     http.StatusForbidden,
			expectedError:      "Forbidden",
			expectedMessage:    "Insufficient permissions: missing required role",
			expectedStatusCode: 403,
		},
		{
			name:               "RequirePermission - missing claims returns 401 with proper format",
			middleware:         RequirePermission(cfg, log, "todo.write"),
			claims:             nil,
			expectedStatus:     http.StatusUnauthorized,
			expectedError:      "Unauthorized",
			expectedMessage:    "Missing authentication",
			expectedStatusCode: 401,
		},
		{
			name:       "RequirePermission - missing permission returns 403 with proper format",
			middleware: RequirePermission(cfg, log, "todo.delete"),
			claims: &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
				Authorization: []entity.Authorization{
					{
						App:         "STACKFORGE",
						Permissions: []string{"todo.read", "todo.write"},
					},
				},
			},
			expectedStatus:     http.StatusForbidden,
			expectedError:      "Forbidden",
			expectedMessage:    "Insufficient permissions: missing required permission",
			expectedStatusCode: 403,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Set request ID for error response
			c.Set(response.RequestIDKey, "test-request-id")

			// Set claims in context if provided
			if tt.claims != nil {
				SetClaims(c, tt.claims)
			}

			// Execute middleware
			tt.middleware(c)

			// Verify status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse and verify error response format
			var errorResp response.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &errorResp)
			assert.NoError(t, err, "Response should be valid JSON")

			// Verify error response fields
			assert.Equal(t, tt.expectedError, errorResp.Error, "Error type should match")
			assert.Equal(t, tt.expectedMessage, errorResp.Message, "Error message should match")
			assert.Equal(t, tt.expectedStatusCode, errorResp.Code, "Status code in response should match")
			assert.Equal(t, "test-request-id", errorResp.RequestID, "Request ID should be included")
		})
	}
}

// TestRBACWithAppSpecificRoles verifies role checking specifically with app-specific roles
func TestRBACWithAppSpecificRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{AppCode: "STACKFORGE"}
	log := logger.New()

	tests := []struct {
		name           string
		requiredRole   string
		claims         *entity.Claims
		expectedStatus int
	}{
		{
			name:         "user has role in app-specific authorization",
			requiredRole: "developer",
			claims: &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
				Authorization: []entity.Authorization{
					{
						App:   "STACKFORGE",
						Roles: []string{"developer", "tester"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "user has role only in different app",
			requiredRole: "developer",
			claims: &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
				Authorization: []entity.Authorization{
					{
						App:   "OTHER-SERVICE",
						Roles: []string{"developer"},
					},
				},
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:         "user has multiple roles in app-specific authorization",
			requiredRole: "admin",
			claims: &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
				Authorization: []entity.Authorization{
					{
						App:   "STACKFORGE",
						Roles: []string{"user", "moderator", "admin"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			SetClaims(c, tt.claims)

			middleware := RequireRole(cfg, log, tt.requiredRole)
			middleware(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestRBACWithGlobalRoles verifies role checking specifically with GLOBAL roles
func TestRBACWithGlobalRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{AppCode: "STACKFORGE"}
	log := logger.New()

	tests := []struct {
		name           string
		requiredRole   string
		claims         *entity.Claims
		expectedStatus int
	}{
		{
			name:         "user has role in GLOBAL authorization",
			requiredRole: "superadmin",
			claims: &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
				Authorization: []entity.Authorization{
					{
						App:   "GLOBAL",
						Roles: []string{"superadmin"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "user has role in both GLOBAL and app-specific",
			requiredRole: "admin",
			claims: &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
				Authorization: []entity.Authorization{
					{
						App:   "GLOBAL",
						Roles: []string{"admin"},
					},
					{
						App:   "STACKFORGE",
						Roles: []string{"user"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "GLOBAL role takes precedence",
			requiredRole: "superadmin",
			claims: &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
				Authorization: []entity.Authorization{
					{
						App:   "STACKFORGE",
						Roles: []string{"user"},
					},
					{
						App:   "GLOBAL",
						Roles: []string{"superadmin"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			SetClaims(c, tt.claims)

			middleware := RequireRole(cfg, log, tt.requiredRole)
			middleware(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestRBACPermissionExactMatch verifies permission checking with exact match
func TestRBACPermissionExactMatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{AppCode: "STACKFORGE"}
	log := logger.New()

	tests := []struct {
		name               string
		requiredPermission string
		claims             *entity.Claims
		expectedStatus     int
	}{
		{
			name:               "exact permission match in app-specific",
			requiredPermission: "todo.write",
			claims: &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
				Authorization: []entity.Authorization{
					{
						App:         "STACKFORGE",
						Permissions: []string{"todo.read", "todo.write", "todo.delete"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:               "exact permission match in GLOBAL",
			requiredPermission: "system.config",
			claims: &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
				Authorization: []entity.Authorization{
					{
						App:         "GLOBAL",
						Permissions: []string{"system.config"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:               "permission not found - exact match required",
			requiredPermission: "todo.admin",
			claims: &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
				Authorization: []entity.Authorization{
					{
						App:         "STACKFORGE",
						Permissions: []string{"todo.read", "todo.write"},
					},
				},
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:               "partial match is not sufficient",
			requiredPermission: "todo.write.all",
			claims: &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
				Authorization: []entity.Authorization{
					{
						App:         "STACKFORGE",
						Permissions: []string{"todo.write"},
					},
				},
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			SetClaims(c, tt.claims)

			middleware := RequirePermission(cfg, log, tt.requiredPermission)
			middleware(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestRBACPermissionWildcard verifies permission checking with wildcard
func TestRBACPermissionWildcard(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{AppCode: "STACKFORGE"}
	log := logger.New()

	tests := []struct {
		name               string
		requiredPermission string
		claims             *entity.Claims
		expectedStatus     int
	}{
		{
			name:               "wildcard permission in app-specific grants access",
			requiredPermission: "todo.write",
			claims: &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
				Authorization: []entity.Authorization{
					{
						App:         "STACKFORGE",
						Permissions: []string{"*"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:               "wildcard permission in GLOBAL grants access",
			requiredPermission: "any.permission",
			claims: &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
				Authorization: []entity.Authorization{
					{
						App:         "GLOBAL",
						Permissions: []string{"*"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:               "wildcard with other permissions",
			requiredPermission: "system.admin",
			claims: &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
				Authorization: []entity.Authorization{
					{
						App:         "STACKFORGE",
						Permissions: []string{"todo.read", "*", "todo.write"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:               "wildcard in GLOBAL overrides app-specific restrictions",
			requiredPermission: "restricted.action",
			claims: &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
				Authorization: []entity.Authorization{
					{
						App:         "STACKFORGE",
						Permissions: []string{"todo.read"},
					},
					{
						App:         "GLOBAL",
						Permissions: []string{"*"},
					},
				},
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			SetClaims(c, tt.claims)

			middleware := RequirePermission(cfg, log, tt.requiredPermission)
			middleware(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestRBACMissingClaimsErrorHandling verifies error handling when claims are missing
func TestRBACMissingClaimsErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{AppCode: "STACKFORGE"}
	log := logger.New()

	tests := []struct {
		name           string
		middleware     gin.HandlerFunc
		setupContext   func(*gin.Context)
		expectedStatus int
		expectedError  string
	}{
		{
			name:       "RequireRole with no claims in context",
			middleware: RequireRole(cfg, log, "admin"),
			setupContext: func(c *gin.Context) {
				// Don't set any claims
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Missing authentication",
		},
		{
			name:       "RequirePermission with no claims in context",
			middleware: RequirePermission(cfg, log, "todo.write"),
			setupContext: func(c *gin.Context) {
				// Don't set any claims
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Missing authentication",
		},
		{
			name:       "RequireAnyRole with no claims in context",
			middleware: RequireAnyRole(cfg, log, "admin", "moderator"),
			setupContext: func(c *gin.Context) {
				// Don't set any claims
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Missing authentication",
		},
		{
			name:       "RequireAnyPermission with no claims in context",
			middleware: RequireAnyPermission(cfg, log, "todo.write", "todo.delete"),
			setupContext: func(c *gin.Context) {
				// Don't set any claims
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Missing authentication",
		},
		{
			name:       "RequireRole with empty authorization array",
			middleware: RequireRole(cfg, log, "admin"),
			setupContext: func(c *gin.Context) {
				SetClaims(c, &entity.Claims{
					Subject:       "user123",
					Username:      "testuser",
					Authorization: []entity.Authorization{},
				})
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions: missing required role",
		},
		{
			name:       "RequirePermission with empty authorization array",
			middleware: RequirePermission(cfg, log, "todo.write"),
			setupContext: func(c *gin.Context) {
				SetClaims(c, &entity.Claims{
					Subject:       "user123",
					Username:      "testuser",
					Authorization: []entity.Authorization{},
				})
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions: missing required permission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			// Setup context
			tt.setupContext(c)

			// Execute middleware
			tt.middleware(c)

			// Verify status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse and verify error message
			var errorResp response.ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &errorResp)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedError, errorResp.Message)
		})
	}
}

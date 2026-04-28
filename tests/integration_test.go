package tests

import (
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/delivery/http/handler"
	"github.com/mafzaidi/stackforge/internal/delivery/http/middleware"
	"github.com/mafzaidi/stackforge/internal/infrastructure/auth"
	"github.com/mafzaidi/stackforge/internal/infrastructure/config"
	"github.com/mafzaidi/stackforge/internal/infrastructure/logger"
	authUseCase "github.com/mafzaidi/stackforge/internal/usecase/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Requirements: 1.1, 1.3, 1.4, 2.1, 2.8, 4.4, 4.5, 5.5, 5.6, 9.1

// setupTestRouter creates a test router with auth components
func setupTestRouter(t *testing.T, privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Create test config
	cfg := &config.Config{
		AppCode: "STACKFORGE",
	}

	// Create mock JWKS server
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Convert public key to JWK format
		jwks := map[string]interface{}{
			"keys": []map[string]interface{}{
				{
					"kty": "RSA",
					"use": "sig",
					"alg": "RS256",
					"kid": "test-kid",
					"n":   encodePublicKeyN(publicKey),
					"e":   encodePublicKeyE(publicKey),
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	}))
	t.Cleanup(jwksServer.Close)

	// Initialize auth components
	appLogger := logger.New()
	jwksClient := auth.NewJWKSClient(jwksServer.URL, 1*time.Hour, appLogger)
	jwtService := auth.NewJWTService(jwksClient, "STACKFORGE", "authorizer")

	// Create router
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	// Create mock authorizer server for SSO redirects
	authorizerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock authorizer endpoints
		if r.URL.Path == "/auth/login" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Login page"))
		} else if r.URL.Path == "/auth/logout" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Logout page"))
		}
	}))
	t.Cleanup(authorizerServer.Close)

	// Setup auth use cases
	loginUC := authUseCase.NewLoginUseCase(authorizerServer.URL, "http://localhost:4040/auth/callback")
	callbackUC := authUseCase.NewCallbackUseCase(jwtService)
	logoutUC := authUseCase.NewLogoutUseCase(authorizerServer.URL, authorizerServer.URL+"/auth/login")

	// Setup auth handler
	authHandler := handler.NewAuthHandler(loginUC, callbackUC, logoutUC, appLogger)

	// Register auth routes
	router.GET("/auth/login", authHandler.Login)
	router.GET("/auth/callback", authHandler.Callback)
	router.GET("/auth/logout", authHandler.Logout)

	// Public endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Protected endpoints
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware(jwtService, appLogger))
	{
		api.GET("/todos", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"todos": []string{"task1", "task2"}})
		})

		// RBAC protected endpoint - requires admin role
		api.GET("/admin",
			middleware.RequireRole(cfg, appLogger, "admin"),
			func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
			})

		// Permission protected endpoint - requires todo.write permission
		api.GET("/write",
			middleware.RequirePermission(cfg, appLogger, "todo.write"),
			func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "write access granted"})
			})
	}

	return router
}

// TestFullSSOLoginFlow tests the complete SSO authentication flow
// Requirements: 1.1, 1.3, 1.4
func TestFullSSOLoginFlow(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	router := setupTestRouter(t, privateKey, publicKey)

	// Step 1: Access protected endpoint without token - should redirect to login
	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	location := w.Header().Get("Location")
	assert.Contains(t, location, "/auth/login")
	assert.Contains(t, location, "redirect_uri=")

	// Step 2: Simulate callback with valid token
	expiresAt := time.Now().Add(1 * time.Hour).Unix()
	claims := map[string]interface{}{
		"iss":      "authorizer",
		"sub":      "user123",
		"aud":      []string{"STACKFORGE"},
		"exp":      expiresAt,
		"iat":      time.Now().Unix(),
		"username": "testuser",
		"email":    "test@example.com",
		"authorization": []map[string]interface{}{
			{
				"app":         "STACKFORGE",
				"roles":       []string{"user"},
				"permissions": []string{"todo.read"},
			},
		},
	}

	tokenString, err := createTestToken(privateKey, "test-kid", claims)
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodGet, "/auth/callback?token="+tokenString, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/dashboard", w.Header().Get("Location"))

	// Verify cookie was set
	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, "jwt_user_token", cookies[0].Name)
	assert.Equal(t, tokenString, cookies[0].Value)
	assert.True(t, cookies[0].HttpOnly)

	// Step 3: Access protected endpoint with cookie
	req = httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	req.AddCookie(cookies[0])
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "todos")
}

// TestAccessProtectedEndpointWithoutToken tests accessing protected endpoint without authentication
// Requirements: 2.1, 2.8
func TestAccessProtectedEndpointWithoutToken(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	router := setupTestRouter(t, privateKey, publicKey)

	req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Missing authentication token")
}

// TestAccessProtectedEndpointWithValidToken tests accessing protected endpoint with valid token
// Requirements: 2.1, 2.8
func TestAccessProtectedEndpointWithValidToken(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	router := setupTestRouter(t, privateKey, publicKey)

	// Create valid token
	expiresAt := time.Now().Add(1 * time.Hour).Unix()
	claims := map[string]interface{}{
		"iss":      "authorizer",
		"sub":      "user123",
		"aud":      []string{"STACKFORGE"},
		"exp":      expiresAt,
		"iat":      time.Now().Unix(),
		"username": "testuser",
		"email":    "test@example.com",
		"authorization": []map[string]interface{}{
			{
				"app":         "STACKFORGE",
				"roles":       []string{"user"},
				"permissions": []string{"todo.read"},
			},
		},
	}

	tokenString, err := createTestToken(privateKey, "test-kid", claims)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "todos")
}

// TestAccessProtectedEndpointWithExpiredToken tests accessing protected endpoint with expired token
// Requirements: 2.1, 2.8
func TestAccessProtectedEndpointWithExpiredToken(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	router := setupTestRouter(t, privateKey, publicKey)

	// Create expired token
	expiresAt := time.Now().Add(-1 * time.Hour).Unix()
	claims := map[string]interface{}{
		"iss":      "authorizer",
		"sub":      "user123",
		"aud":      []string{"STACKFORGE"},
		"exp":      expiresAt,
		"iat":      time.Now().Add(-2 * time.Hour).Unix(),
		"username": "testuser",
		"email":    "test@example.com",
		"authorization": []map[string]interface{}{
			{
				"app":         "STACKFORGE",
				"roles":       []string{"user"},
				"permissions": []string{"todo.read"},
			},
		},
	}

	tokenString, err := createTestToken(privateKey, "test-kid", claims)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Token expired")
}

// TestRBACEnforcementOnProtectedEndpoints tests role-based access control
// Requirements: 4.4, 4.5
func TestRBACEnforcementOnProtectedEndpoints(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	router := setupTestRouter(t, privateKey, publicKey)

	tests := []struct {
		name           string
		roles          []string
		endpoint       string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "user with admin role can access admin endpoint",
			roles:          []string{"admin"},
			endpoint:       "/api/admin",
			expectedStatus: http.StatusOK,
			expectedBody:   "admin access granted",
		},
		{
			name:           "user without admin role cannot access admin endpoint",
			roles:          []string{"user"},
			endpoint:       "/api/admin",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Insufficient permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expiresAt := time.Now().Add(1 * time.Hour).Unix()
			claims := map[string]interface{}{
				"iss":      "authorizer",
				"sub":      "user123",
				"aud":      []string{"STACKFORGE"},
				"exp":      expiresAt,
				"iat":      time.Now().Unix(),
				"username": "testuser",
				"email":    "test@example.com",
				"authorization": []map[string]interface{}{
					{
						"app":         "STACKFORGE",
						"roles":       tt.roles,
						"permissions": []string{},
					},
				},
			}

			tokenString, err := createTestToken(privateKey, "test-kid", claims)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, tt.endpoint, nil)
			req.Header.Set("Authorization", "Bearer "+tokenString)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}

// TestPermissionEnforcementOnProtectedEndpoints tests permission-based access control
// Requirements: 5.5, 5.6
func TestPermissionEnforcementOnProtectedEndpoints(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	router := setupTestRouter(t, privateKey, publicKey)

	tests := []struct {
		name           string
		permissions    []string
		endpoint       string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "user with todo.write permission can access write endpoint",
			permissions:    []string{"todo.write"},
			endpoint:       "/api/write",
			expectedStatus: http.StatusOK,
			expectedBody:   "write access granted",
		},
		{
			name:           "user without todo.write permission cannot access write endpoint",
			permissions:    []string{"todo.read"},
			endpoint:       "/api/write",
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Insufficient permissions",
		},
		{
			name:           "user with wildcard permission can access write endpoint",
			permissions:    []string{"*"},
			endpoint:       "/api/write",
			expectedStatus: http.StatusOK,
			expectedBody:   "write access granted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expiresAt := time.Now().Add(1 * time.Hour).Unix()
			claims := map[string]interface{}{
				"iss":      "authorizer",
				"sub":      "user123",
				"aud":      []string{"STACKFORGE"},
				"exp":      expiresAt,
				"iat":      time.Now().Unix(),
				"username": "testuser",
				"email":    "test@example.com",
				"authorization": []map[string]interface{}{
					{
						"app":         "STACKFORGE",
						"roles":       []string{},
						"permissions": tt.permissions,
					},
				},
			}

			tokenString, err := createTestToken(privateKey, "test-kid", claims)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, tt.endpoint, nil)
			req.Header.Set("Authorization", "Bearer "+tokenString)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}

// TestPublicEndpointAccessWithoutToken tests that public endpoints are accessible without authentication
// Requirements: 9.1
func TestPublicEndpointAccessWithoutToken(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	router := setupTestRouter(t, privateKey, publicKey)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
}

// TestGlobalRoleAccess tests that GLOBAL roles grant access to app-specific endpoints
// Requirements: 4.4, 4.5
func TestGlobalRoleAccess(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	router := setupTestRouter(t, privateKey, publicKey)

	// Create token with GLOBAL admin role
	expiresAt := time.Now().Add(1 * time.Hour).Unix()
	claims := map[string]interface{}{
		"iss":      "authorizer",
		"sub":      "user123",
		"aud":      []string{"STACKFORGE"},
		"exp":      expiresAt,
		"iat":      time.Now().Unix(),
		"username": "globaladmin",
		"email":    "admin@example.com",
		"authorization": []map[string]interface{}{
			{
				"app":         "GLOBAL",
				"roles":       []string{"admin"},
				"permissions": []string{"*"},
			},
		},
	}

	tokenString, err := createTestToken(privateKey, "test-kid", claims)
	require.NoError(t, err)

	// Test access to app-specific admin endpoint with GLOBAL role
	req := httptest.NewRequest(http.MethodGet, "/api/admin", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "admin access granted")
}

// TestTokenInCookieVsHeader tests that tokens can be provided via cookie or header
// Requirements: 2.1
func TestTokenInCookieVsHeader(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	router := setupTestRouter(t, privateKey, publicKey)

	expiresAt := time.Now().Add(1 * time.Hour).Unix()
	claims := map[string]interface{}{
		"iss":      "authorizer",
		"sub":      "user123",
		"aud":      []string{"STACKFORGE"},
		"exp":      expiresAt,
		"iat":      time.Now().Unix(),
		"username": "testuser",
		"email":    "test@example.com",
		"authorization": []map[string]interface{}{
			{
				"app":         "STACKFORGE",
				"roles":       []string{"user"},
				"permissions": []string{"todo.read"},
			},
		},
	}

	tokenString, err := createTestToken(privateKey, "test-kid", claims)
	require.NoError(t, err)

	// Test with header
	t.Run("token in header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Test with cookie
	t.Run("token in cookie", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
		req.AddCookie(&http.Cookie{
			Name:  "jwt_user_token",
			Value: tokenString,
		})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestInvalidTokenScenarios tests various invalid token scenarios
// Requirements: 2.1, 2.8
func TestInvalidTokenScenarios(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	router := setupTestRouter(t, privateKey, publicKey)

	tests := []struct {
		name           string
		setupToken     func() string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "invalid issuer",
			setupToken: func() string {
				claims := map[string]interface{}{
					"iss": "wrong-issuer",
					"sub": "user123",
					"aud": []string{"STACKFORGE"},
					"exp": time.Now().Add(1 * time.Hour).Unix(),
					"iat": time.Now().Unix(),
				}
				token, _ := createTestToken(privateKey, "test-kid", claims)
				return token
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid token issuer",
		},
		{
			name: "invalid audience",
			setupToken: func() string {
				claims := map[string]interface{}{
					"iss": "authorizer",
					"sub": "user123",
					"aud": []string{"WRONG-APP"},
					"exp": time.Now().Add(1 * time.Hour).Unix(),
					"iat": time.Now().Unix(),
				}
				token, _ := createTestToken(privateKey, "test-kid", claims)
				return token
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid token audience",
		},
		{
			name: "malformed token",
			setupToken: func() string {
				return "not.a.valid.jwt.token"
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid token signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupToken()

			req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedError)
		})
	}
}

// TestLogoutFlow tests the complete logout flow
// Requirements: 1.1, 1.3
func TestLogoutFlow(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	router := setupTestRouter(t, privateKey, publicKey)

	req := httptest.NewRequest(http.MethodGet, "/auth/logout", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)

	// Verify redirect to authorizer logout
	location := w.Header().Get("Location")
	assert.Contains(t, location, "/auth/logout")

	// Verify cookie was cleared
	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, "jwt_user_token", cookies[0].Name)
	assert.Equal(t, "", cookies[0].Value)
	assert.Equal(t, -1, cookies[0].MaxAge)
}

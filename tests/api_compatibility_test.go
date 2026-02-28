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
	"github.com/mafzaidi/stackforge/internal/infrastructure/persistence/memory"
	authUseCase "github.com/mafzaidi/stackforge/internal/usecase/auth"
	todoUseCase "github.com/mafzaidi/stackforge/internal/usecase/todo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIBackwardCompatibility_HealthEndpoint tests the health check endpoint
// Requirements: 10.3 - API backward compatibility
func TestAPIBackwardCompatibility_HealthEndpoint(t *testing.T) {
	router := setupCompatibilityTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "Service is healthy", response["message"])
}

// TestAPIBackwardCompatibility_LoginEndpoint tests the SSO login endpoint
// Requirements: 10.3 - API backward compatibility
func TestAPIBackwardCompatibility_LoginEndpoint(t *testing.T) {
	router := setupCompatibilityTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)

	location := w.Header().Get("Location")
	assert.Contains(t, location, "/auth/login")
	assert.Contains(t, location, "redirect_uri=")
}

// TestAPIBackwardCompatibility_CallbackEndpoint tests the SSO callback endpoint
// Requirements: 10.3 - API backward compatibility
func TestAPIBackwardCompatibility_CallbackEndpoint(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	router := setupCompatibilityTestRouterWithKeys(t, privateKey, publicKey)

	t.Run("missing token returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/callback", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["message"], "Missing authentication token")
		assert.NotEmpty(t, response["request_id"])
	})

	t.Run("valid token redirects to dashboard", func(t *testing.T) {
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

		req := httptest.NewRequest(http.MethodGet, "/auth/callback?token="+tokenString, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "/dashboard", w.Header().Get("Location"))

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, "auth_token", cookies[0].Name)
		assert.True(t, cookies[0].HttpOnly)
	})

	t.Run("expired token returns 401", func(t *testing.T) {
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

		req := httptest.NewRequest(http.MethodGet, "/auth/callback?token="+tokenString, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["message"], "Invalid authentication token")
		assert.NotEmpty(t, response["request_id"])
	})
}

// TestAPIBackwardCompatibility_LogoutEndpoint tests the SSO logout endpoint
// Requirements: 10.3 - API backward compatibility
func TestAPIBackwardCompatibility_LogoutEndpoint(t *testing.T) {
	router := setupCompatibilityTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/auth/logout", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)

	location := w.Header().Get("Location")
	assert.Contains(t, location, "/auth/logout")

	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, "auth_token", cookies[0].Name)
	assert.Equal(t, "", cookies[0].Value)
	assert.Equal(t, -1, cookies[0].MaxAge)
}

// TestAPIBackwardCompatibility_TodosEndpoint tests the todos list endpoint
// Requirements: 10.3 - API backward compatibility
func TestAPIBackwardCompatibility_TodosEndpoint(t *testing.T) {
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	router := setupCompatibilityTestRouterWithKeys(t, privateKey, publicKey)

	t.Run("no authentication returns 401", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["message"], "Missing authentication token")
		assert.NotEmpty(t, response["request_id"])
	})

	t.Run("valid token returns todos", func(t *testing.T) {
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

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "items")
	})

	t.Run("token in cookie works", func(t *testing.T) {
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
		req.AddCookie(&http.Cookie{
			Name:  "auth_token",
			Value: tokenString,
		})
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("expired token returns 401", func(t *testing.T) {
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

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["message"], "Token expired")
	})
}

// TestAPIBackwardCompatibility_ErrorResponseFormat tests error response structure
// Requirements: 10.3 - API backward compatibility
func TestAPIBackwardCompatibility_ErrorResponseFormat(t *testing.T) {
	router := setupCompatibilityTestRouter(t)

	tests := []struct {
		name           string
		endpoint       string
		method         string
		expectedStatus int
		errorContains  string
	}{
		{
			name:           "missing token on protected endpoint",
			endpoint:       "/api/todos",
			method:         http.MethodGet,
			expectedStatus: http.StatusUnauthorized,
			errorContains:  "Missing authentication token",
		},
		{
			name:           "missing token on callback",
			endpoint:       "/auth/callback",
			method:         http.MethodGet,
			expectedStatus: http.StatusBadRequest,
			errorContains:  "Missing authentication token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.endpoint, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			// Verify error response structure
			assert.Contains(t, response, "error")
			assert.Contains(t, response, "message")
			assert.Contains(t, response, "request_id")
			assert.Contains(t, response["message"], tt.errorContains)
			assert.NotEmpty(t, response["request_id"])
		})
	}
}

// TestAPIBackwardCompatibility_RequestIDHeader tests request ID tracking
// Requirements: 10.3 - API backward compatibility
func TestAPIBackwardCompatibility_RequestIDHeader(t *testing.T) {
	router := setupCompatibilityTestRouter(t)

	t.Run("generates request ID if not provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		requestID := w.Header().Get("X-Request-ID")
		assert.NotEmpty(t, requestID)
	})

	t.Run("uses provided request ID", func(t *testing.T) {
		customID := "custom-request-id-12345"
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		req.Header.Set("X-Request-ID", customID)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		requestID := w.Header().Get("X-Request-ID")
		assert.Equal(t, customID, requestID)
	})
}

// setupCompatibilityTestRouter creates a test router without JWT validation
func setupCompatibilityTestRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	// Create mock authorizer server
	authorizerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/login" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Login page"))
		} else if r.URL.Path == "/auth/logout" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Logout page"))
		}
	}))
	t.Cleanup(authorizerServer.Close)

	// Setup auth use cases (without JWT validation for basic tests)
	appLogger := logger.New()
	loginUC := authUseCase.NewLoginUseCase(authorizerServer.URL, "http://localhost:4040/auth/callback")
	callbackUC := authUseCase.NewCallbackUseCase(nil) // No JWT service for basic tests
	logoutUC := authUseCase.NewLogoutUseCase(authorizerServer.URL, authorizerServer.URL+"/auth/login")

	// Setup auth handler
	authHandler := handler.NewAuthHandler(loginUC, callbackUC, logoutUC, appLogger)

	// Register auth routes
	router.GET("/auth/login", authHandler.Login)
	router.GET("/auth/callback", authHandler.Callback)
	router.GET("/auth/logout", authHandler.Logout)

	// Health check endpoint
	healthHandler := handler.NewHealthHandler(nil)
	router.GET("/health", healthHandler.HealthCheck)

	// Protected API routes (will fail without auth)
	api := router.Group("/api")
	api.Use(func(c *gin.Context) {
		// Simple auth check for compatibility tests
		token := c.GetHeader("Authorization")
		cookie, _ := c.Cookie("auth_token")
		if token == "" && cookie == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Unauthorized",
				"message":    "Missing authentication token",
				"code":       http.StatusUnauthorized,
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}
		c.Next()
	})
	{
		repo := memory.NewTodoRepository()
		listUC := todoUseCase.NewListUseCase(repo)
		todoHandler := handler.NewTodoHandler(listUC)
		api.GET("/todos", todoHandler.List)
	}

	return router
}

// setupCompatibilityTestRouterWithKeys creates a test router with full JWT validation
func setupCompatibilityTestRouterWithKeys(t *testing.T, privateKey interface{}, publicKey interface{}) *gin.Engine {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		AppCode: "STACKFORGE",
	}

	// Create mock JWKS server
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwks := map[string]interface{}{
			"keys": []map[string]interface{}{
				{
					"kty": "RSA",
					"use": "sig",
					"alg": "RS256",
					"kid": "test-kid",
					"n":   encodePublicKeyN(publicKey.(*rsa.PublicKey)),
					"e":   encodePublicKeyE(publicKey.(*rsa.PublicKey)),
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

	// Create mock authorizer server
	authorizerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	// Health check endpoint
	healthHandler := handler.NewHealthHandler(nil)
	router.GET("/health", healthHandler.HealthCheck)

	// Protected API routes
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware(jwtService, appLogger))
	{
		repo := memory.NewTodoRepository()
		listUC := todoUseCase.NewListUseCase(repo)
		todoHandler := handler.NewTodoHandler(listUC)
		api.GET("/todos", todoHandler.List)

		// RBAC protected endpoints
		api.GET("/admin",
			middleware.RequireRole(cfg, appLogger, "admin"),
			func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
			})

		api.GET("/write",
			middleware.RequirePermission(cfg, appLogger, "todo.write"),
			func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "write access granted"})
			})
	}

	return router
}

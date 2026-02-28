package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/mafzaidi/stackforge/internal/infrastructure/auth"
	"github.com/mafzaidi/stackforge/internal/infrastructure/logger"
	"github.com/stretchr/testify/require"
)

// **Validates: Requirements 2.1**
// Property 8: Token Extraction from Multiple Sources
// For any valid JWT token, if it's provided in either the Authorization header
// (as "Bearer <token>") or in the auth_token cookie, the Auth_Middleware should
// successfully extract and validate it.
func TestProperty_TokenExtractionFromMultipleSources(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generate a single key pair for all tests
	privateKey, publicKey, err := generateTestKeyPair()
	require.NoError(t, err)

	properties.Property("Token extraction from Authorization header", prop.ForAll(
		func(subject string, username string, email string) bool {
			gin.SetMode(gin.TestMode)

			// Create valid token
			now := time.Now().Unix()
			claims := map[string]interface{}{
				"iss":      "authorizer",
				"sub":      subject,
				"aud":      []string{"STACKFORGE"},
				"exp":      now + 3600,
				"iat":      now,
				"username": username,
				"email":    email,
			}

			tokenString, err := createTestToken(privateKey, "test-kid", claims)
			if err != nil {
				t.Logf("Failed to create token: %v", err)
				return false
			}

			// Create mock JWKS client
			mockClient := &mockJWKSClient{
				keys: map[string]*rsa.PublicKey{
					"test-kid": publicKey,
				},
			}

			// Create JWT service
			jwtService := auth.NewJWTService(mockClient, "STACKFORGE", "authorizer")
			log := logger.New()

			// Create test router with auth middleware
			router := gin.New()
			router.Use(AuthMiddleware(jwtService, log))
			router.GET("/test", func(c *gin.Context) {
				claims, err := GetClaims(c)
				if err != nil {
					t.Logf("Failed to get claims: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get claims"})
					return
				}
				c.JSON(http.StatusOK, gin.H{"user": claims.Username})
			})

			// Test with Authorization header
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer "+tokenString)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Verify successful authentication
			if w.Code != http.StatusOK {
				t.Logf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
				return false
			}

			return true
		},
		genNonEmptyString(),
		genOptionalString(),
		genOptionalEmail(),
	))

	properties.Property("Token extraction from cookie", prop.ForAll(
		func(subject string, username string, email string) bool {
			gin.SetMode(gin.TestMode)

			// Create valid token
			now := time.Now().Unix()
			claims := map[string]interface{}{
				"iss":      "authorizer",
				"sub":      subject,
				"aud":      []string{"STACKFORGE"},
				"exp":      now + 3600,
				"iat":      now,
				"username": username,
				"email":    email,
			}

			tokenString, err := createTestToken(privateKey, "test-kid", claims)
			if err != nil {
				t.Logf("Failed to create token: %v", err)
				return false
			}

			// Create mock JWKS client
			mockClient := &mockJWKSClient{
				keys: map[string]*rsa.PublicKey{
					"test-kid": publicKey,
				},
			}

			// Create JWT service
			jwtService := auth.NewJWTService(mockClient, "STACKFORGE", "authorizer")
			log := logger.New()

			// Create test router with auth middleware
			router := gin.New()
			router.Use(AuthMiddleware(jwtService, log))
			router.GET("/test", func(c *gin.Context) {
				claims, err := GetClaims(c)
				if err != nil {
					t.Logf("Failed to get claims: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get claims"})
					return
				}
				c.JSON(http.StatusOK, gin.H{"user": claims.Username})
			})

			// Test with cookie
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.AddCookie(&http.Cookie{
				Name:  "auth_token",
				Value: tokenString,
			})
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Verify successful authentication
			if w.Code != http.StatusOK {
				t.Logf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
				return false
			}

			return true
		},
		genNonEmptyString(),
		genOptionalString(),
		genOptionalEmail(),
	))

	properties.Property("Token extraction prioritizes header over cookie", prop.ForAll(
		func(subject string, username string, email string) bool {
			gin.SetMode(gin.TestMode)

			// Create valid token for header
			now := time.Now().Unix()
			headerClaims := map[string]interface{}{
				"iss":      "authorizer",
				"sub":      subject,
				"aud":      []string{"STACKFORGE"},
				"exp":      now + 3600,
				"iat":      now,
				"username": username,
				"email":    email,
			}

			headerToken, err := createTestToken(privateKey, "test-kid", headerClaims)
			if err != nil {
				t.Logf("Failed to create header token: %v", err)
				return false
			}

			// Create different token for cookie (with different username)
			cookieUsername := "cookie-" + username
			cookieClaims := map[string]interface{}{
				"iss":      "authorizer",
				"sub":      subject,
				"aud":      []string{"STACKFORGE"},
				"exp":      now + 3600,
				"iat":      now,
				"username": cookieUsername,
				"email":    email,
			}

			cookieToken, err := createTestToken(privateKey, "test-kid", cookieClaims)
			if err != nil {
				t.Logf("Failed to create cookie token: %v", err)
				return false
			}

			// Create mock JWKS client
			mockClient := &mockJWKSClient{
				keys: map[string]*rsa.PublicKey{
					"test-kid": publicKey,
				},
			}

			// Create JWT service
			jwtService := auth.NewJWTService(mockClient, "STACKFORGE", "authorizer")
			log := logger.New()

			// Create test router with auth middleware
			router := gin.New()
			router.Use(AuthMiddleware(jwtService, log))
			router.GET("/test", func(c *gin.Context) {
				claims, err := GetClaims(c)
				if err != nil {
					t.Logf("Failed to get claims: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get claims"})
					return
				}
				c.JSON(http.StatusOK, gin.H{"user": claims.Username})
			})

			// Test with both header and cookie
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer "+headerToken)
			req.AddCookie(&http.Cookie{
				Name:  "auth_token",
				Value: cookieToken,
			})
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Verify successful authentication
			if w.Code != http.StatusOK {
				t.Logf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
				return false
			}

			// Verify that header token was used (not cookie token)
			// The username should match the header token's username
			if !contains(w.Body.String(), username) {
				t.Logf("Expected username '%s' from header token, but got: %s", username, w.Body.String())
				return false
			}

			// Verify that cookie token's username is NOT present
			if username != "" && cookieUsername != username && contains(w.Body.String(), cookieUsername) {
				t.Logf("Cookie token was used instead of header token")
				return false
			}

			return true
		},
		genNonEmptyString(),
		genNonEmptyString(), // Use non-empty username to verify priority
		genOptionalEmail(),
	))

	properties.Property("Token extraction fails when no token provided", prop.ForAll(
		func() bool {
			gin.SetMode(gin.TestMode)

			// Create mock JWKS client
			mockClient := &mockJWKSClient{
				keys: map[string]*rsa.PublicKey{
					"test-kid": publicKey,
				},
			}

			// Create JWT service
			jwtService := auth.NewJWTService(mockClient, "STACKFORGE", "authorizer")
			log := logger.New()

			// Create test router with auth middleware
			router := gin.New()
			router.Use(AuthMiddleware(jwtService, log))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			// Test without any token
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Verify authentication failure
			if w.Code != http.StatusUnauthorized {
				t.Logf("Expected status 401, got %d", w.Code)
				return false
			}

			if !contains(w.Body.String(), "Missing authentication token") {
				t.Logf("Expected 'Missing authentication token' error, got: %s", w.Body.String())
				return false
			}

			return true
		},
	))

	properties.Property("Token extraction with various Bearer formats", prop.ForAll(
		func(subject string, bearerCase string) bool {
			gin.SetMode(gin.TestMode)

			// Create valid token
			now := time.Now().Unix()
			claims := map[string]interface{}{
				"iss": "authorizer",
				"sub": subject,
				"aud": []string{"STACKFORGE"},
				"exp": now + 3600,
				"iat": now,
			}

			tokenString, err := createTestToken(privateKey, "test-kid", claims)
			if err != nil {
				t.Logf("Failed to create token: %v", err)
				return false
			}

			// Create mock JWKS client
			mockClient := &mockJWKSClient{
				keys: map[string]*rsa.PublicKey{
					"test-kid": publicKey,
				},
			}

			// Create JWT service
			jwtService := auth.NewJWTService(mockClient, "STACKFORGE", "authorizer")
			log := logger.New()

			// Create test router with auth middleware
			router := gin.New()
			router.Use(AuthMiddleware(jwtService, log))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			// Test with different Bearer case formats
			var authHeader string
			switch bearerCase {
			case "Bearer":
				authHeader = "Bearer " + tokenString
			case "bearer":
				authHeader = "bearer " + tokenString
			case "BEARER":
				authHeader = "BEARER " + tokenString
			default:
				authHeader = "Bearer " + tokenString
			}

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", authHeader)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Verify successful authentication (case-insensitive Bearer)
			if w.Code != http.StatusOK {
				t.Logf("Expected status 200 for '%s', got %d. Body: %s", authHeader, w.Code, w.Body.String())
				return false
			}

			return true
		},
		genNonEmptyString(),
		gen.OneConstOf("Bearer", "bearer", "BEARER"),
	))

	properties.TestingRun(t)
}

// mockJWKSClient is a mock implementation of JWKSClient for testing
type mockJWKSClient struct {
	keys map[string]*rsa.PublicKey
	err  error
}

func (m *mockJWKSClient) GetKey(kid string) (*rsa.PublicKey, error) {
	if m.err != nil {
		return nil, m.err
	}
	key, exists := m.keys[kid]
	if !exists {
		return nil, errors.New("key not found")
	}
	return key, nil
}

func (m *mockJWKSClient) RefreshKeys() error {
	return m.err
}

// generateTestKeyPair generates an RSA key pair for testing
func generateTestKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

// createTestToken creates a JWT token for testing
func createTestToken(privateKey *rsa.PrivateKey, kid string, claims map[string]interface{}) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims(claims))
	token.Header["kid"] = kid
	return token.SignedString(privateKey)
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/infrastructure/logger"
	"github.com/stretchr/testify/assert"
)

// mockJWTService is a mock implementation of JWTService for testing
type mockJWTService struct {
	validateFunc func(token string) (*entity.Claims, error)
}

func (m *mockJWTService) ValidateToken(token string) (*entity.Claims, error) {
	if m.validateFunc != nil {
		return m.validateFunc(token)
	}
	return nil, errors.New("not implemented")
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock JWT service that returns valid claims
	mockService := &mockJWTService{
		validateFunc: func(token string) (*entity.Claims, error) {
			return &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
				Email:    "test@example.com",
			}, nil
		},
	}

	log := logger.New()

	// Create test router with auth middleware
	router := gin.New()
	router.Use(AuthMiddleware(mockService, log))
	router.GET("/test", func(c *gin.Context) {
		claims, err := GetClaims(c)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		c.JSON(http.StatusOK, gin.H{"user": claims.Username})
	})

	// Test with Authorization header
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "testuser")
}

func TestAuthMiddleware_ValidTokenFromCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock JWT service that returns valid claims
	mockService := &mockJWTService{
		validateFunc: func(token string) (*entity.Claims, error) {
			return &entity.Claims{
				Subject:  "user123",
				Username: "testuser",
				Email:    "test@example.com",
			}, nil
		},
	}

	log := logger.New()

	// Create test router with auth middleware
	router := gin.New()
	router.Use(AuthMiddleware(mockService, log))
	router.GET("/test", func(c *gin.Context) {
		claims, err := GetClaims(c)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		c.JSON(http.StatusOK, gin.H{"user": claims.Username})
	})

	// Test with cookie
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "jwt_user_token",
		Value: "valid-token",
	})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "testuser")
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockJWTService{}
	log := logger.New()

	router := gin.New()
	router.Use(AuthMiddleware(mockService, log))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Missing authentication token")
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockJWTService{
		validateFunc: func(token string) (*entity.Claims, error) {
			return nil, errors.New("Token expired")
		},
	}
	log := logger.New()

	router := gin.New()
	router.Use(AuthMiddleware(mockService, log))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer expired-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Token expired")
}

func TestAuthMiddleware_InvalidIssuer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockJWTService{
		validateFunc: func(token string) (*entity.Claims, error) {
			return nil, errors.New("Invalid token issuer")
		},
	}
	log := logger.New()

	router := gin.New()
	router.Use(AuthMiddleware(mockService, log))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-issuer-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid token issuer")
}

func TestAuthMiddleware_InvalidAudience(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockJWTService{
		validateFunc: func(token string) (*entity.Claims, error) {
			return nil, errors.New("Invalid token audience")
		},
	}
	log := logger.New()

	router := gin.New()
	router.Use(AuthMiddleware(mockService, log))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-audience-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid token audience")
}

func TestAuthMiddleware_InvalidSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockJWTService{
		validateFunc: func(token string) (*entity.Claims, error) {
			return nil, errors.New("failed to get public key: unknown key ID")
		},
	}
	log := logger.New()

	router := gin.New()
	router.Use(AuthMiddleware(mockService, log))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-signature-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid token signature")
}

func TestExtractToken_FromAuthorizationHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "valid bearer token",
			header:   "Bearer valid-token-123",
			expected: "valid-token-123",
		},
		{
			name:     "bearer with extra spaces",
			header:   "Bearer  token-with-spaces  ",
			expected: "token-with-spaces",
		},
		{
			name:     "lowercase bearer",
			header:   "bearer lowercase-token",
			expected: "lowercase-token",
		},
		{
			name:     "invalid format - no bearer",
			header:   "valid-token-123",
			expected: "",
		},
		{
			name:     "invalid format - empty token",
			header:   "Bearer ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			c.Request.Header.Set("Authorization", tt.header)

			token := extractToken(c)
			assert.Equal(t, tt.expected, token)
		})
	}
}

func TestExtractToken_FromCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.AddCookie(&http.Cookie{
		Name:  "jwt_user_token",
		Value: "cookie-token-123",
	})

	token := extractToken(c)
	assert.Equal(t, "cookie-token-123", token)
}

func TestExtractToken_HeaderPriorityOverCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Authorization", "Bearer header-token")
	c.Request.AddCookie(&http.Cookie{
		Name:  "jwt_user_token",
		Value: "cookie-token",
	})

	token := extractToken(c)
	assert.Equal(t, "header-token", token)
}

func TestExtractToken_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	token := extractToken(c)
	assert.Equal(t, "", token)
}

func TestMapValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "expired token",
			err:      errors.New("Token expired"),
			expected: "Token expired",
		},
		{
			name:     "invalid issuer",
			err:      errors.New("Invalid token issuer"),
			expected: "Invalid token issuer",
		},
		{
			name:     "invalid audience",
			err:      errors.New("Invalid token audience"),
			expected: "Invalid token audience",
		},
		{
			name:     "failed to get public key",
			err:      errors.New("failed to get public key: unknown kid"),
			expected: "Invalid token signature",
		},
		{
			name:     "unexpected signing method",
			err:      errors.New("unexpected signing method: HS256"),
			expected: "Invalid token signature",
		},
		{
			name:     "missing kid",
			err:      errors.New("missing or invalid kid in token header"),
			expected: "Invalid token signature",
		},
		{
			name:     "generic error",
			err:      errors.New("some other error"),
			expected: "Invalid token signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapValidationError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

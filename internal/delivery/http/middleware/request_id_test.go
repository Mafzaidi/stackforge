package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestIDMiddleware_GeneratesID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		requestID := GetRequestID(c)
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify request ID was generated and added to response header
	requestID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID)

	// Verify it's a valid UUID format (36 characters with dashes)
	assert.Len(t, requestID, 36)
	assert.Contains(t, requestID, "-")
}

func TestRequestIDMiddleware_UsesExistingID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	existingID := "existing-request-id-123"

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		requestID := GetRequestID(c)
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", existingID)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the existing request ID was used
	requestID := w.Header().Get("X-Request-ID")
	assert.Equal(t, existingID, requestID)
}

func TestRequestIDMiddleware_StoresInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var capturedRequestID string

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		capturedRequestID = GetRequestID(c)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, capturedRequestID)

	// Verify the request ID in response header matches what was stored in context
	responseRequestID := w.Header().Get("X-Request-ID")
	assert.Equal(t, responseRequestID, capturedRequestID)
}

func TestGetRequestID_ReturnsEmptyWhenNotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	requestID := GetRequestID(c)
	assert.Empty(t, requestID)
}

func TestGetRequestID_ReturnsStoredID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	expectedID := "test-request-id-456"

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Set(RequestIDKey, expectedID)

	requestID := GetRequestID(c)
	assert.Equal(t, expectedID, requestID)
}

func TestGetRequestID_HandlesInvalidType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	// Store a non-string value
	c.Set(RequestIDKey, 12345)

	requestID := GetRequestID(c)
	assert.Empty(t, requestID)
}

func TestRequestIDMiddleware_MultipleRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		requestID := GetRequestID(c)
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	// Make multiple requests and verify each gets a unique ID
	requestIDs := make(map[string]bool)

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		router.ServeHTTP(w, req)

		requestID := w.Header().Get("X-Request-ID")
		assert.NotEmpty(t, requestID)

		// Verify this ID hasn't been seen before
		assert.False(t, requestIDs[requestID], "Request ID should be unique")
		requestIDs[requestID] = true
	}

	// Verify we got 5 unique IDs
	assert.Len(t, requestIDs, 5)
}

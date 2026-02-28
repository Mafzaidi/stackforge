package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRespondWithError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		statusCode int
		errorType  string
		message    string
		requestID  string
	}{
		{
			name:       "unauthorized error with request ID",
			statusCode: http.StatusUnauthorized,
			errorType:  "Unauthorized",
			message:    "Missing authentication token",
			requestID:  "test-request-123",
		},
		{
			name:       "forbidden error with request ID",
			statusCode: http.StatusForbidden,
			errorType:  "Forbidden",
			message:    "Insufficient permissions",
			requestID:  "test-request-456",
		},
		{
			name:       "error without request ID",
			statusCode: http.StatusBadRequest,
			errorType:  "Bad Request",
			message:    "Invalid input",
			requestID:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

			// Set request ID if provided
			if tt.requestID != "" {
				c.Set(RequestIDKey, tt.requestID)
			}

			RespondWithError(c, tt.statusCode, tt.errorType, tt.message)

			assert.Equal(t, tt.statusCode, w.Code)

			var response ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.errorType, response.Error)
			assert.Equal(t, tt.message, response.Message)
			assert.Equal(t, tt.statusCode, response.Code)
			assert.Equal(t, tt.requestID, response.RequestID)
		})
	}
}

func TestUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Set(RequestIDKey, "test-request-789")

	Unauthorized(c, "Token expired")

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Unauthorized", response.Error)
	assert.Equal(t, "Token expired", response.Message)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Equal(t, "test-request-789", response.RequestID)
}

func TestForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Set(RequestIDKey, "test-request-abc")

	Forbidden(c, "Insufficient permissions: missing required role")

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Forbidden", response.Error)
	assert.Equal(t, "Insufficient permissions: missing required role", response.Message)
	assert.Equal(t, http.StatusForbidden, response.Code)
	assert.Equal(t, "test-request-abc", response.RequestID)
}

func TestBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Set(RequestIDKey, "test-request-def")

	BadRequest(c, "Missing required parameter")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Bad Request", response.Error)
	assert.Equal(t, "Missing required parameter", response.Message)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "test-request-def", response.RequestID)
}

func TestInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Set(RequestIDKey, "test-request-ghi")

	InternalServerError(c, "Authentication service configuration error")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Internal Server Error", response.Error)
	assert.Equal(t, "Authentication service configuration error", response.Message)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "test-request-ghi", response.RequestID)
}

func TestServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Set(RequestIDKey, "test-request-jkl")

	ServiceUnavailable(c, "Authentication service unavailable")

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Service Unavailable", response.Error)
	assert.Equal(t, "Authentication service unavailable", response.Message)
	assert.Equal(t, http.StatusServiceUnavailable, response.Code)
	assert.Equal(t, "test-request-jkl", response.RequestID)
}

func TestErrorResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Set(RequestIDKey, "test-request-mno")

	Unauthorized(c, "Test message")

	// Verify JSON structure
	var rawJSON map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &rawJSON)
	require.NoError(t, err)

	// Verify all required fields are present
	assert.Contains(t, rawJSON, "error")
	assert.Contains(t, rawJSON, "message")
	assert.Contains(t, rawJSON, "code")
	assert.Contains(t, rawJSON, "request_id")

	// Verify field types
	assert.IsType(t, "", rawJSON["error"])
	assert.IsType(t, "", rawJSON["message"])
	assert.IsType(t, float64(0), rawJSON["code"]) // JSON numbers are float64
	assert.IsType(t, "", rawJSON["request_id"])
}

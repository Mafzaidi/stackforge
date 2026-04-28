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

func setupTestContext() (*httptest.ResponseRecorder, *gin.Context) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
	return w, c
}

func TestRespondError(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		statusMessage string
		errDetail     string
	}{
		{"unauthorized", http.StatusUnauthorized, "Unauthorized", "invalid token"},
		{"forbidden", http.StatusForbidden, "Forbidden", "missing role"},
		{"bad request", http.StatusBadRequest, "Bad Request", "invalid input"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := setupTestContext()

			writeError(c, tt.statusCode, tt.statusMessage, tt.errDetail)

			assert.Equal(t, tt.statusCode, w.Code)

			var resp Response
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)

			assert.Equal(t, tt.statusCode, resp.Status.Code)
			assert.Equal(t, tt.statusMessage, resp.Status.Message)
			assert.Equal(t, "error occurred", resp.Error)
			assert.Equal(t, tt.errDetail, resp.Error)
		})
	}
}

func TestUnauthorized(t *testing.T) {
	w, c := setupTestContext()

	Unauthorized(c, "Token expired")

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.Status.Code)
	assert.Equal(t, "Unauthorized", resp.Status.Message)
	assert.Equal(t, "Token expired", resp.Error)
}

func TestForbidden(t *testing.T) {
	w, c := setupTestContext()

	Forbidden(c, "Insufficient permissions")

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, http.StatusForbidden, resp.Status.Code)
	assert.Equal(t, "Forbidden", resp.Status.Message)
	assert.Equal(t, "Insufficient permissions", resp.Error)
}

func TestBadRequest(t *testing.T) {
	w, c := setupTestContext()

	BadRequest(c, "Missing required parameter")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.Status.Code)
	assert.Equal(t, "Bad Request", resp.Status.Message)
	assert.Equal(t, "Missing required parameter", resp.Error)
}

func TestInternalServerError(t *testing.T) {
	w, c := setupTestContext()

	InternalServerError(c, "something went wrong")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, resp.Status.Code)
	assert.Equal(t, "Internal Server Error", resp.Status.Message)
	assert.Equal(t, "something went wrong", resp.Error)
}

func TestServiceUnavailable(t *testing.T) {
	w, c := setupTestContext()

	ServiceUnavailable(c, "database unreachable")

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, http.StatusServiceUnavailable, resp.Status.Code)
	assert.Equal(t, "Service Unavailable", resp.Status.Message)
	assert.Equal(t, "database unreachable", resp.Error)
}

func TestSuccess(t *testing.T) {
	w, c := setupTestContext()

	data := gin.H{"id": "123", "name": "test"}
	Success(c, "Item retrieved successfully", data)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.Status.Code)
	assert.Equal(t, "Item retrieved successfully", resp.Status.Message)
	assert.NotNil(t, resp.Data)
}

func TestCreated(t *testing.T) {
	w, c := setupTestContext()

	data := gin.H{"id": "456"}
	Created(c, "Item created successfully", data)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, resp.Status.Code)
	assert.Equal(t, "Item created successfully", resp.Status.Message)
	assert.NotNil(t, resp.Data)
}

func TestSuccessWithNilData(t *testing.T) {
	w, c := setupTestContext()

	Success(c, "Operation completed", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.Status.Code)
	assert.Equal(t, "Operation completed", resp.Status.Message)
	assert.Nil(t, resp.Data)
}

func TestErrorResponseJSONStructure(t *testing.T) {
	w, c := setupTestContext()

	Unauthorized(c, "invalid token")

	var raw map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &raw)
	require.NoError(t, err)

	// Verify top-level keys
	assert.Contains(t, raw, "status")
	assert.Contains(t, raw, "message")
	assert.Contains(t, raw, "error")

	// Verify status block
	status, ok := raw["status"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, status, "code")
	assert.Contains(t, status, "message")
}

func TestSuccessResponseJSONStructure(t *testing.T) {
	w, c := setupTestContext()

	Success(c, "ok", gin.H{"items": []string{"a", "b"}})

	var raw map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &raw)
	require.NoError(t, err)

	// Verify top-level keys
	assert.Contains(t, raw, "status")
	assert.Contains(t, raw, "data")

	// Verify status block
	status, ok := raw["status"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, status, "code")
	assert.Contains(t, status, "message")
}

package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const RequestIDKey = "request_id"

// ErrorResponse represents a consistent error response structure
// All error responses in the application should use this format
type ErrorResponse struct {
	Error     string `json:"error"`      // Error type (e.g., "Unauthorized", "Forbidden")
	Message   string `json:"message"`    // User-friendly error message
	Code      int    `json:"code"`       // HTTP status code
	RequestID string `json:"request_id"` // Unique request ID for tracking
}

// getRequestID retrieves the request ID from the gin.Context
// Returns empty string if no request ID is found
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// RespondWithError sends a consistent error response with the given status code and message
// It automatically includes the request ID from the context
func RespondWithError(c *gin.Context, statusCode int, errorType string, message string) {
	requestID := getRequestID(c)
	
	c.JSON(statusCode, ErrorResponse{
		Error:     errorType,
		Message:   message,
		Code:      statusCode,
		RequestID: requestID,
	})
}

// Unauthorized sends a 401 Unauthorized error response
func Unauthorized(c *gin.Context, message string) {
	RespondWithError(c, http.StatusUnauthorized, "Unauthorized", message)
}

// Forbidden sends a 403 Forbidden error response
func Forbidden(c *gin.Context, message string) {
	RespondWithError(c, http.StatusForbidden, "Forbidden", message)
}

// BadRequest sends a 400 Bad Request error response
func BadRequest(c *gin.Context, message string) {
	RespondWithError(c, http.StatusBadRequest, "Bad Request", message)
}

// InternalServerError sends a 500 Internal Server Error response
// This should be used for unexpected errors that don't expose internal details
func InternalServerError(c *gin.Context, message string) {
	RespondWithError(c, http.StatusInternalServerError, "Internal Server Error", message)
}

// ServiceUnavailable sends a 503 Service Unavailable error response
// This should be used when external dependencies are unavailable
func ServiceUnavailable(c *gin.Context, message string) {
	RespondWithError(c, http.StatusServiceUnavailable, "Service Unavailable", message)
}

// Success sends a 200 OK response with the given data
func Success(c *gin.Context, data interface{}) {
	requestID := getRequestID(c)
	
	response := gin.H{
		"request_id": requestID,
	}
	
	// Merge data into response
	if dataMap, ok := data.(gin.H); ok {
		for k, v := range dataMap {
			response[k] = v
		}
	} else {
		response["data"] = data
	}
	
	c.JSON(http.StatusOK, response)
}

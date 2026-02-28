package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "request_id"

// RequestIDMiddleware generates a unique request ID for each request
// and stores it in the gin.Context for use in logging and error responses.
// If the request already has an X-Request-ID header (from a proxy/load balancer),
// it will use that value instead of generating a new one.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get request ID from header (if set by a proxy/load balancer)
		requestID := c.GetHeader("X-Request-ID")
		
		// Generate a new UUID if no request ID was provided
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store request ID in context for use by other middleware and handlers
		c.Set(RequestIDKey, requestID)

		// Add request ID to response headers for client-side tracking
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// GetRequestID retrieves the request ID from the gin.Context
// Returns empty string if no request ID is found
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

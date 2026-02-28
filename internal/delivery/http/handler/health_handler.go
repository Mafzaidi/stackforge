package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/infrastructure/database"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	healthChecker *database.HealthChecker
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(healthChecker *database.HealthChecker) *HealthHandler {
	return &HealthHandler{
		healthChecker: healthChecker,
	}
}

// HealthCheck handles the health check endpoint for probeness
// GET /health
// Requirements: 4.1, 4.2
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	response := gin.H{
		"status":  "ok",
		"message": "Service is healthy",
	}

	// If health checker is available, include database health status
	if h.healthChecker != nil {
		healthStatus := h.healthChecker.CheckHealth(c.Request.Context())
		
		// Add database health information
		response["databases"] = gin.H{
			"mongodb": gin.H{
				"status":     healthStatus.MongoDB.Status,
				"message":    healthStatus.MongoDB.Message,
				"latency_ms": healthStatus.MongoDB.Latency.Milliseconds(),
			},
			"postgresql": gin.H{
				"status":     healthStatus.PostgreSQL.Status,
				"message":    healthStatus.PostgreSQL.Message,
				"latency_ms": healthStatus.PostgreSQL.Latency.Milliseconds(),
			},
		}
		response["timestamp"] = healthStatus.Timestamp
		
		// If any database is unhealthy, set overall status to degraded
		if healthStatus.MongoDB.Status == "unhealthy" || healthStatus.PostgreSQL.Status == "unhealthy" {
			response["status"] = "degraded"
			response["message"] = "Service is running but some databases are unhealthy"
		}
	}

	c.JSON(http.StatusOK, response)
}

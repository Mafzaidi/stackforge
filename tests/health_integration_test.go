package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/delivery/http/handler"
	"github.com/mafzaidi/stackforge/internal/infrastructure/config"
	"github.com/mafzaidi/stackforge/internal/infrastructure/database"
	"github.com/mafzaidi/stackforge/internal/infrastructure/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthEndpoint_Integration tests the health endpoint integration
// Requirements: 4.1, 4.2
func TestHealthEndpoint_Integration(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	appLogger := logger.New()

	// Try to load database configuration
	dbConfig, err := config.LoadDatabaseConfig()
	
	var healthHandler *handler.HealthHandler
	
	if err == nil && (dbConfig.MongoDB.URI != "" || dbConfig.PostgreSQL.Host != "") {
		// Database configuration available - test with database health checks
		t.Log("Testing with database health checks")
		
		connManager := database.NewConnectionManager(&dbConfig, appLogger)
		healthChecker := database.NewHealthChecker(connManager, appLogger)
		healthHandler = handler.NewHealthHandler(healthChecker)
	} else {
		// No database configuration - test without database health checks
		t.Log("Testing without database health checks")
		healthHandler = handler.NewHealthHandler(nil)
	}

	// Create minimal router config
	r := gin.New()
	r.GET("/health", healthHandler.HealthCheck)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Execute
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	// Basic assertions
	assert.Contains(t, response, "status")
	assert.Contains(t, response, "message")
	
	// Log response for debugging
	t.Logf("Health response: %+v", response)
	
	// If databases are configured, check database health structure
	if databases, ok := response["databases"]; ok {
		t.Log("Database health information included in response")
		assert.NotNil(t, databases)
		
		dbMap := databases.(map[string]interface{})
		assert.Contains(t, dbMap, "mongodb")
		assert.Contains(t, dbMap, "postgresql")
	} else {
		t.Log("No database health information in response (expected when no databases configured)")
	}
}

// TestHealthEndpoint_StructuredResponse tests the health endpoint returns structured data
// Requirements: 4.1, 4.2
func TestHealthEndpoint_StructuredResponse(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create health handler without databases for predictable testing
	healthHandler := handler.NewHealthHandler(nil)

	r := gin.New()
	r.GET("/health", healthHandler.HealthCheck)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Execute
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	// Verify structured response
	assert.Equal(t, "ok", response["status"])
	assert.IsType(t, "", response["message"])
	assert.NotEmpty(t, response["message"])
}

// TestHealthEndpoint_WithDatabaseHealthChecker tests health endpoint with database health checker
// Requirements: 4.1, 4.2
func TestHealthEndpoint_WithDatabaseHealthChecker(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	appLogger := logger.New()

	// Create a connection manager without actually connecting
	dbConfig := config.DatabaseConfig{
		MongoDB: config.MongoDBConfig{
			URI:      "mongodb://localhost:27017",
			Database: "test",
		},
		PostgreSQL: config.PostgreSQLConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Database: "test",
			SSLMode:  "disable",
		},
	}

	connManager := database.NewConnectionManager(&dbConfig, appLogger)
	healthChecker := database.NewHealthChecker(connManager, appLogger)
	healthHandler := handler.NewHealthHandler(healthChecker)

	r := gin.New()
	r.GET("/health", healthHandler.HealthCheck)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Execute
	r.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	// Should include database health information
	assert.Contains(t, response, "databases")
	assert.Contains(t, response, "timestamp")
	
	databases := response["databases"].(map[string]interface{})
	
	// Check MongoDB health
	mongodb := databases["mongodb"].(map[string]interface{})
	assert.Contains(t, mongodb, "status")
	assert.Contains(t, mongodb, "message")
	assert.Contains(t, mongodb, "latency_ms")
	
	// Check PostgreSQL health
	postgresql := databases["postgresql"].(map[string]interface{})
	assert.Contains(t, postgresql, "status")
	assert.Contains(t, postgresql, "message")
	assert.Contains(t, postgresql, "latency_ms")
	
	// Since we didn't actually connect, both should be unhealthy
	assert.Equal(t, "unhealthy", mongodb["status"])
	assert.Equal(t, "unhealthy", postgresql["status"])
	
	// Overall status should be degraded
	assert.Equal(t, "degraded", response["status"])
}

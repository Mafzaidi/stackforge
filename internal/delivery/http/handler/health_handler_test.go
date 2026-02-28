package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/infrastructure/config"
	"github.com/mafzaidi/stackforge/internal/infrastructure/database"
	"github.com/mafzaidi/stackforge/internal/infrastructure/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler_HealthCheck_WithoutDatabases(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Create health handler without database health checker
	healthHandler := NewHealthHandler(nil)
	router.GET("/health", healthHandler.HealthCheck)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "Service is healthy", response["message"])
	assert.Nil(t, response["databases"], "Should not include database health when checker is nil")
}

func TestHealthHandler_HealthCheck_WithHealthyDatabases(t *testing.T) {
	// This test requires actual database connections
	// Skip if database configuration is not available
	dbConfig, err := config.LoadDatabaseConfig()
	if err != nil || (dbConfig.MongoDB.URI == "" && dbConfig.PostgreSQL.Host == "") {
		t.Skip("Skipping test: database configuration not available")
	}

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	appLogger := logger.New()

	// Create connection manager and connect to databases
	connManager := database.NewConnectionManager(&dbConfig, appLogger)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to available databases
	if dbConfig.MongoDB.URI != "" {
		if err := connManager.ConnectMongoDB(ctx); err != nil {
			t.Skipf("Skipping test: failed to connect to MongoDB: %v", err)
		}
		defer connManager.Shutdown(context.Background())
	}

	if dbConfig.PostgreSQL.Host != "" {
		if err := connManager.ConnectPostgreSQL(ctx); err != nil {
			t.Skipf("Skipping test: failed to connect to PostgreSQL: %v", err)
		}
		defer connManager.Shutdown(context.Background())
	}

	// Create health checker and handler
	healthChecker := database.NewHealthChecker(connManager, appLogger)
	healthHandler := NewHealthHandler(healthChecker)
	router.GET("/health", healthHandler.HealthCheck)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "ok", response["status"])
	assert.NotNil(t, response["databases"], "Should include database health")
	assert.NotNil(t, response["timestamp"], "Should include timestamp")
	
	// Check database health structure
	databases := response["databases"].(map[string]interface{})
	
	if dbConfig.MongoDB.URI != "" {
		mongodb := databases["mongodb"].(map[string]interface{})
		assert.Equal(t, "healthy", mongodb["status"])
		assert.NotNil(t, mongodb["latency_ms"])
	}
	
	if dbConfig.PostgreSQL.Host != "" {
		postgresql := databases["postgresql"].(map[string]interface{})
		assert.Equal(t, "healthy", postgresql["status"])
		assert.NotNil(t, postgresql["latency_ms"])
	}
}

func TestHealthHandler_HealthCheck_WithUnhealthyDatabases(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	appLogger := logger.New()

	// Create connection manager with invalid configuration (not connected)
	dbConfig := config.DatabaseConfig{
		MongoDB: config.MongoDBConfig{
			URI:      "mongodb://invalid:27017",
			Database: "test",
			Timeout:  5 * time.Second,
		},
		PostgreSQL: config.PostgreSQLConfig{
			Host:     "invalid",
			Port:     5432,
			User:     "test",
			Password: "test",
			Database: "test",
			SSLMode:  "disable",
		},
	}

	// Create connection manager without connecting
	connManager := database.NewConnectionManager(&dbConfig, appLogger)
	
	// Create health checker and handler
	healthChecker := database.NewHealthChecker(connManager, appLogger)
	healthHandler := NewHealthHandler(healthChecker)
	router.GET("/health", healthHandler.HealthCheck)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	// Status should be degraded when databases are unhealthy
	assert.Equal(t, "degraded", response["status"])
	assert.Equal(t, "Service is running but some databases are unhealthy", response["message"])
	assert.NotNil(t, response["databases"], "Should include database health")
	
	// Check database health structure
	databases := response["databases"].(map[string]interface{})
	
	mongodb := databases["mongodb"].(map[string]interface{})
	assert.Equal(t, "unhealthy", mongodb["status"])
	assert.NotEmpty(t, mongodb["message"], "Should include error message")
	
	postgresql := databases["postgresql"].(map[string]interface{})
	assert.Equal(t, "unhealthy", postgresql["status"])
	assert.NotEmpty(t, postgresql["message"], "Should include error message")
}

func TestHealthHandler_HealthCheck_ResponseStructure(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	appLogger := logger.New()

	// Create connection manager without connecting
	dbConfig := config.DatabaseConfig{
		MongoDB: config.MongoDBConfig{
			URI:      "mongodb://localhost:27017",
			Database: "test",
			Timeout:  5 * time.Second,
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
	healthHandler := NewHealthHandler(healthChecker)
	router.GET("/health", healthHandler.HealthCheck)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert response structure
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	// Check required fields
	assert.Contains(t, response, "status")
	assert.Contains(t, response, "message")
	assert.Contains(t, response, "databases")
	assert.Contains(t, response, "timestamp")
	
	// Check databases structure
	databases := response["databases"].(map[string]interface{})
	assert.Contains(t, databases, "mongodb")
	assert.Contains(t, databases, "postgresql")
	
	// Check MongoDB health structure
	mongodb := databases["mongodb"].(map[string]interface{})
	assert.Contains(t, mongodb, "status")
	assert.Contains(t, mongodb, "message")
	assert.Contains(t, mongodb, "latency_ms")
	
	// Check PostgreSQL health structure
	postgresql := databases["postgresql"].(map[string]interface{})
	assert.Contains(t, postgresql, "status")
	assert.Contains(t, postgresql, "message")
	assert.Contains(t, postgresql, "latency_ms")
}

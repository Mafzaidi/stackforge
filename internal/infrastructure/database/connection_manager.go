// Package database provides database connection management for MongoDB and PostgreSQL.
//
// This package implements connection pooling, health checks, retry logic with exponential backoff,
// and graceful shutdown for dual database support. It is designed to work with the stackforge
// application's clean architecture.
//
// # Connection Management
//
// The ConnectionManager handles the lifecycle of database connections:
//   - Establishes connections with retry logic
//   - Configures connection pooling for optimal performance
//   - Provides health checking capabilities
//   - Handles graceful shutdown
//
// # Example Usage
//
//	// Load database configuration
//	dbConfig, err := config.LoadDatabaseConfig()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create connection manager
//	logger := logger.New()
//	connMgr := database.NewConnectionManager(&dbConfig, logger)
//
//	// Connect to databases
//	ctx := context.Background()
//	if err := connMgr.ConnectMongoDB(ctx); err != nil {
//	    log.Fatal(err)
//	}
//	if err := connMgr.ConnectPostgreSQL(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Use the clients
//	mongoClient := connMgr.GetMongoClient()
//	postgresClient := connMgr.GetPostgresGormClient()
//
//	// Setup health checker
//	healthChecker := database.NewHealthChecker(connMgr, logger)
//	status := healthChecker.CheckHealth(ctx)
//
//	// Setup graceful shutdown
//	shutdownHandler := database.NewShutdownHandler(connMgr, logger)
//	go shutdownHandler.WaitForShutdown(ctx, cancel)
//
// # Connection Pooling
//
// MongoDB connection pool configuration:
//   - MinPoolSize: 5
//   - MaxPoolSize: 100
//   - MaxConnIdleTime: 5 minutes
//
// PostgreSQL connection pool configuration:
//   - MaxOpenConns: 100
//   - MaxIdleConns: 10
//   - ConnMaxLifetime: 1 hour
//   - ConnMaxIdleTime: 5 minutes
//
// # Retry Logic
//
// Connection attempts use exponential backoff:
//   - Max Attempts: 3
//   - Initial Delay: 1 second
//   - Max Delay: 10 seconds
//   - Backoff Multiplier: 2
//
// # Health Checks
//
// Health checks ping each database with a 5-second timeout and measure latency.
// Results include status (healthy/unhealthy), error messages, and response time.
//
// # Graceful Shutdown
//
// Shutdown closes all database connections with a 10-second timeout per database.
// The process handles errors gracefully and ensures cleanup even if individual
// shutdowns fail.
package database

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mafzaidi/stackforge/internal/infrastructure/config"
	"github.com/mafzaidi/stackforge/internal/infrastructure/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectionManager manages database connections for MongoDB and PostgreSQL.
//
// It provides connection pooling, retry logic, health checking, and graceful shutdown
// capabilities. The manager ensures reliable database connectivity with automatic
// retry on failures and proper resource cleanup.
//
// ConnectionManager is safe for concurrent use after initialization.
type ConnectionManager struct {
	mongoClient  *mongo.Client
	gormClient   *gorm.DB
	config       *config.DatabaseConfig
	logger       *logger.Logger
	shutdownOnce sync.Once
}

// NewConnectionManager creates a new connection manager with the provided configuration and logger.
//
// The returned ConnectionManager is ready to establish database connections using
// ConnectMongoDB and ConnectPostgreSQL methods. The manager must be initialized
// before calling GetMongoClient or GetPostgresClient.
//
// Parameters:
//   - cfg: Database configuration containing MongoDB and PostgreSQL settings
//   - log: Logger instance for connection lifecycle events
//
// Returns a ConnectionManager instance ready for connection establishment.
func NewConnectionManager(cfg *config.DatabaseConfig, log *logger.Logger) *ConnectionManager {
	return &ConnectionManager{
		config: cfg,
		logger: log,
	}
}

// GetMongoClient returns the connected MongoDB client.
//
// This method should only be called after successfully calling ConnectMongoDB.
// Returns nil if MongoDB connection has not been established.
//
// The returned client is safe for concurrent use and managed by the ConnectionManager.
// Do not call Disconnect on the returned client directly; use ConnectionManager.Shutdown instead.
func (cm *ConnectionManager) GetMongoClient() *mongo.Client {
	return cm.mongoClient
}

// GetPostgresGormClient returns the connected GORM PostgreSQL client.
//
// This method should only be called after successfully calling ConnectPostgreSQL.
// Returns nil if PostgreSQL connection has not been established.
//
// The returned client is safe for concurrent use and managed by the ConnectionManager.
// Do not close the returned client directly; use ConnectionManager.Shutdown instead.
func (cm *ConnectionManager) GetPostgresGormClient() *gorm.DB {
	return cm.gormClient
}

// retryConnect implements exponential backoff retry logic for database connections.
//
// This internal method attempts to establish a database connection up to maxAttempts times,
// using exponential backoff between attempts. The delay starts at initialDelay (1s) and
// doubles after each failure, up to maxDelay (10s).
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - connectFunc: Function that attempts to establish the connection
//   - dbType: Database type identifier for logging (e.g., "mongodb", "postgresql")
//
// Returns nil on successful connection, or an error after all retry attempts are exhausted.
// The error includes the number of attempts made and the last error encountered.
//
// Requirements: 5.1, 5.2, 5.3, 5.4, 5.5
func (cm *ConnectionManager) retryConnect(
	ctx context.Context,
	connectFunc func(context.Context) error,
	dbType string,
) error {
	const (
		maxAttempts  = 3
		initialDelay = 1 * time.Second
		maxDelay     = 10 * time.Second
	)

	var lastErr error
	delay := initialDelay

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		cm.logger.Info("Attempting database connection", logger.Fields{
			"database": dbType,
			"attempt":  attempt,
			"max_attempts": maxAttempts,
		})

		err := connectFunc(ctx)
		if err == nil {
			if attempt > 1 {
				cm.logger.Info("Successfully connected after retries", logger.Fields{
					"database": dbType,
					"attempts": attempt,
				})
			}
			return nil
		}

		lastErr = err
		cm.logger.Warn("Connection attempt failed", logger.Fields{
			"database": dbType,
			"attempt":  attempt,
			"max_attempts": maxAttempts,
			"error": err.Error(),
		})

		// Don't sleep after the last attempt
		if attempt < maxAttempts {
			cm.logger.Info("Retrying connection", logger.Fields{
				"database": dbType,
				"delay": delay.String(),
			})
			
			select {
			case <-time.After(delay):
				// Continue to next attempt
			case <-ctx.Done():
				return fmt.Errorf("connection retry cancelled: %w", ctx.Err())
			}

			// Exponential backoff: double the delay for next attempt
			delay *= 2
			if delay > maxDelay {
				delay = maxDelay
			}
		}
	}

	return fmt.Errorf("failed to connect to %s after %d attempts: %w", dbType, maxAttempts, lastErr)
}
// ConnectMongoDB establishes a MongoDB connection with retry logic and connection pooling.
//
// This method creates a MongoDB client with the following pool configuration:
//   - MinPoolSize: 5 connections
//   - MaxPoolSize: 100 connections
//   - MaxConnIdleTime: 5 minutes
//   - ConnectTimeout: from configuration (default 10s)
//
// The connection is established using exponential backoff retry logic (up to 3 attempts).
// On success, the MongoDB client is stored and can be retrieved via GetMongoClient.
//
// Parameters:
//   - ctx: Context for connection timeout and cancellation
//
// Returns nil on successful connection, or an error if all retry attempts fail.
// The error includes details about the failure and number of attempts made.
//
// Requirements: 2.1, 2.2, 2.3, 7.1, 7.2
func (cm *ConnectionManager) ConnectMongoDB(ctx context.Context) error {
	cm.logger.Info("Starting MongoDB connection", logger.Fields{
		"database": cm.config.MongoDB.Database,
	})

	connectFunc := func(ctx context.Context) error {
		// Create MongoDB client options
		clientOptions := options.Client().
			ApplyURI(cm.config.MongoDB.URI).
			SetMinPoolSize(5).
			SetMaxPoolSize(100).
			SetMaxConnIdleTime(5 * time.Minute).
			SetConnectTimeout(cm.config.MongoDB.Timeout)

		// Create MongoDB client
		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			return fmt.Errorf("failed to create MongoDB client: %w", err)
		}

		// Ping to verify connection
		pingCtx, cancel := context.WithTimeout(ctx, cm.config.MongoDB.Timeout)
		defer cancel()

		if err := client.Ping(pingCtx, nil); err != nil {
			// Close the client if ping fails
			_ = client.Disconnect(ctx)
			return fmt.Errorf("failed to ping MongoDB: %w", err)
		}

		cm.mongoClient = client
		return nil
	}

	// Use retry logic to establish connection
	if err := cm.retryConnect(ctx, connectFunc, "mongodb"); err != nil {
		cm.logger.Error("Failed to connect to MongoDB", logger.Fields{
			"error": err.Error(),
		})
		return err
	}

	cm.logger.Info("Successfully connected to MongoDB", logger.Fields{
		"database": cm.config.MongoDB.Database,
		"pool_config": map[string]interface{}{
			"min_pool_size":      5,
			"max_pool_size":      100,
			"max_conn_idle_time": "5m",
		},
	})

	return nil
}


// ConnectPostgreSQL establishes a PostgreSQL connection via GORM with retry logic and connection pooling.
//
// This method creates a GORM client with the following pool configuration (via underlying *sql.DB):
//   - MaxOpenConns: 100 connections
//   - MaxIdleConns: 10 connections
//   - ConnMaxLifetime: 1 hour
//   - ConnMaxIdleTime: 5 minutes
//
// Auto-migration is disabled — schema is managed by golang-migrate.
//
// The connection is established using exponential backoff retry logic (up to 3 attempts).
// On success, the GORM client is stored and can be retrieved via GetPostgresGormClient.
//
// Parameters:
//   - ctx: Context for connection timeout and cancellation
//
// Returns nil on successful connection, or an error if all retry attempts fail.
//
// Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 9.3
func (cm *ConnectionManager) ConnectPostgreSQL(ctx context.Context) error {
	cm.logger.Info("Starting PostgreSQL connection", logger.Fields{
		"host":     cm.config.PostgreSQL.Host,
		"port":     cm.config.PostgreSQL.Port,
		"database": cm.config.PostgreSQL.Database,
	})

	connectFunc := func(ctx context.Context) error {
		// Build PostgreSQL DSN for GORM
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cm.config.PostgreSQL.Host,
			cm.config.PostgreSQL.Port,
			cm.config.PostgreSQL.User,
			cm.config.PostgreSQL.Password,
			cm.config.PostgreSQL.Database,
			cm.config.PostgreSQL.SSLMode,
		)

		// Open GORM connection with auto-migration disabled
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("failed to open GORM PostgreSQL connection: %w", err)
		}

		// Get underlying *sql.DB to configure connection pool
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("failed to get underlying sql.DB from GORM: %w", err)
		}

		// Configure connection pool
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetConnMaxLifetime(1 * time.Hour)
		sqlDB.SetConnMaxIdleTime(5 * time.Minute)

		// Ping to verify connection
		pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := sqlDB.PingContext(pingCtx); err != nil {
			_ = sqlDB.Close()
			return fmt.Errorf("failed to ping PostgreSQL: %w", err)
		}

		cm.gormClient = db
		return nil
	}

	// Use retry logic to establish connection
	if err := cm.retryConnect(ctx, connectFunc, "postgresql"); err != nil {
		cm.logger.Error("Failed to connect to PostgreSQL", logger.Fields{
			"error": err.Error(),
		})
		return err
	}

	cm.logger.Info("Successfully connected to PostgreSQL", logger.Fields{
		"host":     cm.config.PostgreSQL.Host,
		"port":     cm.config.PostgreSQL.Port,
		"database": cm.config.PostgreSQL.Database,
		"pool_config": map[string]interface{}{
			"max_open_conns":     100,
			"max_idle_conns":     10,
			"conn_max_lifetime":  "1h",
			"conn_max_idle_time": "5m",
		},
	})

	return nil
}

// HealthChecker checks database health and connectivity.
//
// It provides methods to check the health of individual databases (MongoDB, PostgreSQL)
// or all databases at once. Health checks measure response latency and detect connection issues.
//
// HealthChecker is safe for concurrent use.
type HealthChecker struct {
	connManager *ConnectionManager
	logger      *logger.Logger
}

// NewHealthChecker creates a new health checker for database monitoring.
//
// The health checker uses the provided ConnectionManager to perform health checks
// on connected databases. It requires that databases are already connected via
// the ConnectionManager before health checks can succeed.
//
// Parameters:
//   - cm: ConnectionManager with established database connections
//   - log: Logger instance for health check events
//
// Returns a HealthChecker instance ready to perform health checks.
func NewHealthChecker(cm *ConnectionManager, log *logger.Logger) *HealthChecker {
	return &HealthChecker{
		connManager: cm,
		logger:      log,
	}
}

// HealthStatus represents the overall health status of all databases.
//
// It contains individual health information for each database along with
// a timestamp indicating when the health check was performed.
type HealthStatus struct {
	MongoDB    DatabaseHealth `json:"mongodb"`
	PostgreSQL DatabaseHealth `json:"postgresql"`
	Timestamp  time.Time      `json:"timestamp"`
}

// DatabaseHealth represents the health status of an individual database.
//
// Status can be "healthy" or "unhealthy". For unhealthy databases, Message
// contains error details. Latency measures the response time of the health check.
type DatabaseHealth struct {
	Status  string        `json:"status"`
	Message string        `json:"message,omitempty"`
	Latency time.Duration `json:"latency_ms"`
}

// CheckMongoDB checks the health of the MongoDB connection.
//
// This method pings the MongoDB server with a 5-second timeout and measures
// the response latency. If the client is not initialized or the ping fails,
// it returns an unhealthy status with error details.
//
// Parameters:
//   - ctx: Context for health check cancellation
//
// Returns DatabaseHealth with status, error message (if any), and latency.
//
// Requirements: 4.1, 4.3, 4.4, 4.5
func (hc *HealthChecker) CheckMongoDB(ctx context.Context) DatabaseHealth {
	const healthCheckTimeout = 5 * time.Second

	// Check if MongoDB client is available
	if hc.connManager.mongoClient == nil {
		return DatabaseHealth{
			Status:  "unhealthy",
			Message: "MongoDB client not initialized",
			Latency: 0,
		}
	}

	// Create context with timeout
	pingCtx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
	defer cancel()

	// Measure latency
	start := time.Now()
	err := hc.connManager.mongoClient.Ping(pingCtx, nil)
	latency := time.Since(start)

	if err != nil {
		hc.logger.Warn("MongoDB health check failed", logger.Fields{
			"error":   err.Error(),
			"latency": latency.String(),
		})
		return DatabaseHealth{
			Status:  "unhealthy",
			Message: fmt.Sprintf("ping failed: %v", err),
			Latency: latency,
		}
	}

	hc.logger.Info("MongoDB health check passed", logger.Fields{
		"latency": latency.String(),
	})

	return DatabaseHealth{
		Status:  "healthy",
		Message: "",
		Latency: latency,
	}
}

// CheckPostgreSQL checks the health of the PostgreSQL connection.
//
// This method gets the underlying *sql.DB from GORM, then pings the PostgreSQL
// server with a 5-second timeout and measures the response latency.
// If the client is not initialized or the ping fails, it returns an unhealthy status.
//
// Parameters:
//   - ctx: Context for health check cancellation
//
// Returns DatabaseHealth with status, error message (if any), and latency.
//
// Requirements: 7.1, 7.2, 7.3
func (hc *HealthChecker) CheckPostgreSQL(ctx context.Context) DatabaseHealth {
	const healthCheckTimeout = 5 * time.Second

	// Check if GORM client is available
	if hc.connManager.gormClient == nil {
		return DatabaseHealth{
			Status:  "unhealthy",
			Message: "PostgreSQL client not initialized",
			Latency: 0,
		}
	}

	// Get underlying *sql.DB from GORM
	sqlDB, err := hc.connManager.gormClient.DB()
	if err != nil {
		return DatabaseHealth{
			Status:  "unhealthy",
			Message: fmt.Sprintf("failed to get underlying sql.DB: %v", err),
			Latency: 0,
		}
	}

	// Create context with timeout
	pingCtx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
	defer cancel()

	// Measure latency
	start := time.Now()
	err = sqlDB.PingContext(pingCtx)
	latency := time.Since(start)

	if err != nil {
		hc.logger.Warn("PostgreSQL health check failed", logger.Fields{
			"error":   err.Error(),
			"latency": latency.String(),
		})
		return DatabaseHealth{
			Status:  "unhealthy",
			Message: fmt.Sprintf("ping failed: %v", err),
			Latency: latency,
		}
	}

	hc.logger.Info("PostgreSQL health check passed", logger.Fields{
		"latency": latency.String(),
	})

	return DatabaseHealth{
		Status:  "healthy",
		Message: "",
		Latency: latency,
	}
}

// CheckHealth checks the health of all databases concurrently.
//
// This method performs health checks on both MongoDB and PostgreSQL in parallel
// for better performance. It returns a HealthStatus containing the results for
// both databases along with a timestamp.
//
// Parameters:
//   - ctx: Context for health check cancellation
//
// Returns HealthStatus with individual database health results and timestamp.
//
// Requirements: 4.1, 4.2
func (hc *HealthChecker) CheckHealth(ctx context.Context) HealthStatus {
	// Check both databases concurrently for better performance
	mongoChan := make(chan DatabaseHealth, 1)
	postgresChan := make(chan DatabaseHealth, 1)

	// Check MongoDB in goroutine
	go func() {
		mongoChan <- hc.CheckMongoDB(ctx)
	}()

	// Check PostgreSQL in goroutine
	go func() {
		postgresChan <- hc.CheckPostgreSQL(ctx)
	}()

	// Wait for both results
	mongoHealth := <-mongoChan
	postgresHealth := <-postgresChan

	return HealthStatus{
		MongoDB:    mongoHealth,
		PostgreSQL: postgresHealth,
		Timestamp:  time.Now(),
	}
}
// Shutdown gracefully closes all database connections.
//
// This method closes MongoDB and PostgreSQL connections concurrently, each with
// a 10-second timeout. It uses sync.Once to ensure shutdown is only performed once,
// even if called multiple times.
//
// The method handles errors gracefully:
//   - Logs errors but continues shutdown for remaining databases
//   - Returns combined error if multiple databases fail to close
//   - Ensures cleanup attempts for all databases even if some fail
//
// Parameters:
//   - ctx: Context for shutdown cancellation (note: uses internal timeout)
//
// Returns nil if all connections closed successfully, or an error describing
// any failures that occurred during shutdown.
//
// Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 7.5
func (cm *ConnectionManager) Shutdown(ctx context.Context) error {
	var shutdownErr error

	// Use sync.Once to ensure shutdown is only called once
	cm.shutdownOnce.Do(func() {
		const shutdownTimeout = 10 * time.Second

		cm.logger.Info("Starting graceful shutdown of database connections", logger.Fields{})

		// Create channels to track shutdown completion
		mongoDone := make(chan error, 1)
		postgresDone := make(chan error, 1)

		// Shutdown MongoDB connection
		go func() {
			if cm.mongoClient == nil {
				cm.logger.Info("MongoDB client not initialized, skipping shutdown", logger.Fields{})
				mongoDone <- nil
				return
			}

			cm.logger.Info("Closing MongoDB connection", logger.Fields{})

			// Create context with timeout for MongoDB shutdown
			mongoCtx, mongoCancel := context.WithTimeout(ctx, shutdownTimeout)
			defer mongoCancel()

			if err := cm.mongoClient.Disconnect(mongoCtx); err != nil {
				cm.logger.Error("Failed to close MongoDB connection", logger.Fields{
					"error": err.Error(),
				})
				mongoDone <- fmt.Errorf("mongodb shutdown failed: %w", err)
				return
			}

			cm.logger.Info("Successfully closed MongoDB connection", logger.Fields{})
			mongoDone <- nil
		}()

		// Shutdown PostgreSQL connection
		go func() {
			if cm.gormClient == nil {
				cm.logger.Info("PostgreSQL client not initialized, skipping shutdown", logger.Fields{})
				postgresDone <- nil
				return
			}

			cm.logger.Info("Closing PostgreSQL connection", logger.Fields{})

			// Create context with timeout for PostgreSQL shutdown
			postgresCtx, postgresCancel := context.WithTimeout(ctx, shutdownTimeout)
			defer postgresCancel()

			// Get underlying *sql.DB from GORM to close the connection
			sqlDB, err := cm.gormClient.DB()
			if err != nil {
				cm.logger.Error("Failed to get underlying sql.DB from GORM", logger.Fields{
					"error": err.Error(),
				})
				postgresDone <- fmt.Errorf("postgresql shutdown failed: could not get sql.DB: %w", err)
				return
			}

			// Close the underlying sql.DB with timeout
			errChan := make(chan error, 1)
			go func() {
				errChan <- sqlDB.Close()
			}()

			select {
			case err := <-errChan:
				if err != nil {
					cm.logger.Error("Failed to close PostgreSQL connection", logger.Fields{
						"error": err.Error(),
					})
					postgresDone <- fmt.Errorf("postgresql shutdown failed: %w", err)
					return
				}
				cm.logger.Info("Successfully closed PostgreSQL connection", logger.Fields{})
				postgresDone <- nil
			case <-postgresCtx.Done():
				cm.logger.Error("PostgreSQL shutdown timed out", logger.Fields{
					"timeout": shutdownTimeout.String(),
				})
				postgresDone <- fmt.Errorf("postgresql shutdown timeout: %w", postgresCtx.Err())
			}
		}()

		// Wait for both shutdowns to complete
		mongoErr := <-mongoDone
		postgresErr := <-postgresDone

		// Combine errors if both failed
		if mongoErr != nil && postgresErr != nil {
			shutdownErr = fmt.Errorf("multiple shutdown errors: mongodb: %v, postgresql: %v", mongoErr, postgresErr)
			cm.logger.Error("Database shutdown completed with errors", logger.Fields{
				"error": shutdownErr.Error(),
			})
		} else if mongoErr != nil {
			shutdownErr = mongoErr
			cm.logger.Warn("Database shutdown completed with MongoDB error", logger.Fields{
				"error": mongoErr.Error(),
			})
		} else if postgresErr != nil {
			shutdownErr = postgresErr
			cm.logger.Warn("Database shutdown completed with PostgreSQL error", logger.Fields{
				"error": postgresErr.Error(),
			})
		} else {
			cm.logger.Info("Successfully completed graceful shutdown of all database connections", logger.Fields{})
		}
	})

	return shutdownErr
}


// ShutdownHandler handles graceful shutdown of database connections.
//
// It listens for OS signals (SIGINT, SIGTERM) and coordinates the shutdown
// sequence, ensuring all database connections are properly closed before
// the application exits.
//
// Requirements: 6.1
type ShutdownHandler struct {
	connManager *ConnectionManager
	logger      *logger.Logger
}

// NewShutdownHandler creates a new shutdown handler.
//
// The handler will use the provided ConnectionManager to close database
// connections when a shutdown signal is received.
//
// Parameters:
//   - cm: ConnectionManager to shutdown
//   - log: Logger instance for shutdown events
//
// Returns a ShutdownHandler ready to listen for shutdown signals.
//
// Requirements: 6.1
func NewShutdownHandler(cm *ConnectionManager, log *logger.Logger) *ShutdownHandler {
	return &ShutdownHandler{
		connManager: cm,
		logger:      log,
	}
}

// WaitForShutdown waits for a shutdown signal and performs cleanup.
//
// This method blocks until it receives SIGINT or SIGTERM, then cancels the
// provided context and initiates the shutdown sequence. It should be called
// in a goroutine during application startup.
//
// Parameters:
//   - ctx: Application context to cancel on shutdown
//   - cancel: Context cancel function to signal shutdown to other components
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	shutdownHandler := database.NewShutdownHandler(connMgr, logger)
//	go shutdownHandler.WaitForShutdown(ctx, cancel)
//
// Requirements: 6.1
func (sh *ShutdownHandler) WaitForShutdown(ctx context.Context, cancel context.CancelFunc) {
	// Create channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for signal
	sig := <-sigChan
	sh.logger.Info("Received shutdown signal", logger.Fields{
		"signal": sig.String(),
	})

	// Cancel the context to signal shutdown to other components
	cancel()

	// Perform shutdown
	if err := sh.handleShutdown(ctx); err != nil {
		sh.logger.Error("Shutdown completed with errors", logger.Fields{
			"error": err.Error(),
		})
	} else {
		sh.logger.Info("Shutdown completed successfully", logger.Fields{})
	}
}

// handleShutdown performs the actual shutdown sequence.
//
// This internal method creates a shutdown context with a 30-second timeout
// and calls the ConnectionManager's Shutdown method to close all database
// connections.
//
// Parameters:
//   - ctx: Context for shutdown (currently unused, uses internal timeout)
//
// Returns nil if shutdown succeeds, or an error if database shutdown fails.
//
// Requirements: 6.1, 6.2, 6.3, 6.4, 6.5
func (sh *ShutdownHandler) handleShutdown(ctx context.Context) error {
	sh.logger.Info("Starting shutdown sequence", logger.Fields{})

	// Create a context with timeout for shutdown operations
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown database connections
	if err := sh.connManager.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("database shutdown failed: %w", err)
	}

	return nil
}


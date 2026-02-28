package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/mafzaidi/stackforge/internal/delivery/http/handler"
	"github.com/mafzaidi/stackforge/internal/delivery/http/middleware"
	"github.com/mafzaidi/stackforge/internal/delivery/http/router"
	infraAuth "github.com/mafzaidi/stackforge/internal/infrastructure/auth"
	"github.com/mafzaidi/stackforge/internal/infrastructure/config"
	"github.com/mafzaidi/stackforge/internal/infrastructure/database"
	"github.com/mafzaidi/stackforge/internal/infrastructure/logger"
	"github.com/mafzaidi/stackforge/internal/infrastructure/persistence/memory"
	authUseCase "github.com/mafzaidi/stackforge/internal/usecase/auth"
	credentialUseCase "github.com/mafzaidi/stackforge/internal/usecase/credential"
	todoUseCase "github.com/mafzaidi/stackforge/internal/usecase/todo"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or could not be loaded: %v", err)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize structured logger
	appLogger := logger.New()

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Load database configuration
	dbConfig, err := config.LoadDatabaseConfig()
	if err != nil {
		log.Fatalf("Failed to load database configuration: %v", err)
	}

	// Initialize connection manager
	var connManager *database.ConnectionManager
	
	// Only initialize database connections if configuration is provided
	if dbConfig.MongoDB.URI != "" || dbConfig.PostgreSQL.Host != "" {
		// Validate database configuration
		if err := dbConfig.Validate(); err != nil {
			log.Fatalf("Invalid database configuration: %v", err)
		}

		appLogger.Info("Initializing database connections", logger.Fields{
			"mongodb_configured":    dbConfig.MongoDB.URI != "",
			"postgresql_configured": dbConfig.PostgreSQL.Host != "",
		})

		// Create connection manager
		connManager = database.NewConnectionManager(&dbConfig, appLogger)

		// Create context for connection initialization
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Connect to MongoDB if configured
		if dbConfig.MongoDB.URI != "" {
			if err := connManager.ConnectMongoDB(ctx); err != nil {
				log.Fatalf("Failed to connect to MongoDB: %v", err)
			}
		}

		// Connect to PostgreSQL if configured
		if dbConfig.PostgreSQL.Host != "" {
			if err := connManager.ConnectPostgreSQL(ctx); err != nil {
				log.Fatalf("Failed to connect to PostgreSQL: %v", err)
			}
		}

		// Setup graceful shutdown handler
		shutdownHandler := database.NewShutdownHandler(connManager, appLogger)
		shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
		
		// Start goroutine to wait for shutdown signal
		go shutdownHandler.WaitForShutdown(shutdownCtx, shutdownCancel)

		appLogger.Info("Database connections initialized successfully", logger.Fields{})
	} else {
		appLogger.Info("No database configuration provided, using in-memory repositories", logger.Fields{})
	}

	// Initialize infrastructure components
	// Auth infrastructure
	jwksURL := fmt.Sprintf("%s/.well-known/jwks.json", cfg.AuthorizerBaseURL)
	jwksClient := infraAuth.NewJWKSClient(jwksURL, time.Duration(cfg.JWKSCacheDuration)*time.Second, appLogger)
	jwtService := infraAuth.NewJWTService(jwksClient, cfg.AppCode, "authorizer")
	
	// Repository infrastructure (using in-memory implementations)
	// TODO: Update to use database repositories when connManager is available
	todoRepo := memory.NewTodoRepository()
	credentialRepo := memory.NewCredentialRepository()
	
	// Initialize use cases
	loginUC := authUseCase.NewLoginUseCase(cfg.AuthorizerBaseURL, cfg.AppCallbackURL)
	callbackUC := authUseCase.NewCallbackUseCase(jwtService)
	logoutUC := authUseCase.NewLogoutUseCase(cfg.AuthorizerBaseURL, cfg.AppCallbackURL)
	todoListUC := todoUseCase.NewListUseCase(todoRepo)
	credentialListUC := credentialUseCase.NewListUseCase(credentialRepo)
	
	// Initialize handlers
	authHandler := handler.NewAuthHandler(loginUC, callbackUC, logoutUC, appLogger)
	todoHandler := handler.NewTodoHandler(todoListUC)
	credentialHandler := handler.NewCredentialHandler(credentialListUC)
	
	// Initialize health handler with database health checker if available
	var healthHandler *handler.HealthHandler
	if connManager != nil {
		healthChecker := database.NewHealthChecker(connManager, appLogger)
		healthHandler = handler.NewHealthHandler(healthChecker)
	} else {
		healthHandler = handler.NewHealthHandler(nil)
	}
	
	// Setup router with all dependencies
	r := router.Setup(&router.RouterConfig{
		AuthHandler:       authHandler,
		TodoHandler:       todoHandler,
		CredentialHandler: credentialHandler,
		HealthHandler:     healthHandler,
		AuthMiddleware:    middleware.AuthMiddleware(jwtService, appLogger),
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "4040"
	}
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

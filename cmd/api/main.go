package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/mafzaidi/stackforge/internal/delivery/http/handler"
	"github.com/mafzaidi/stackforge/internal/delivery/http/middleware"
	"github.com/mafzaidi/stackforge/internal/delivery/http/router"
	"github.com/mafzaidi/stackforge/internal/domain/repository"
	infraAuth "github.com/mafzaidi/stackforge/internal/infrastructure/auth"
	"github.com/mafzaidi/stackforge/internal/infrastructure/config"
	"github.com/mafzaidi/stackforge/internal/infrastructure/database"
	"github.com/mafzaidi/stackforge/internal/infrastructure/encryption"
	"github.com/mafzaidi/stackforge/internal/infrastructure/logger"
	"github.com/mafzaidi/stackforge/internal/infrastructure/persistence/dualwrite"
	dualwriteCredential "github.com/mafzaidi/stackforge/internal/infrastructure/persistence/dualwrite/credential"
	"github.com/mafzaidi/stackforge/internal/infrastructure/persistence/memory"
	persistMongo "github.com/mafzaidi/stackforge/internal/infrastructure/persistence/mongodb"
	persistPostgres "github.com/mafzaidi/stackforge/internal/infrastructure/persistence/postgres"
	authUseCase "github.com/mafzaidi/stackforge/internal/usecase/auth"
	credentialUseCase "github.com/mafzaidi/stackforge/internal/usecase/credential"
	masterDataUseCase "github.com/mafzaidi/stackforge/internal/usecase/masterdata"
	menuUseCase "github.com/mafzaidi/stackforge/internal/usecase/menu"
	todoUseCase "github.com/mafzaidi/stackforge/internal/usecase/todo"
	userProfilesUseCase "github.com/mafzaidi/stackforge/internal/usecase/userprofiles"
	vaultUseCase "github.com/mafzaidi/stackforge/internal/usecase/vault"
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

	// Initialize encryption service
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		log.Fatal("ENCRYPTION_KEY environment variable is required")
	}
	encryptionSvc, err := encryption.NewAESGCMService(encryptionKey)
	if err != nil {
		log.Fatalf("Failed to initialize encryption service: %v", err)
	}

	// Load database configuration
	dbConfig, err := config.LoadDatabaseConfig()
	if err != nil {
		log.Fatalf("Failed to load database configuration: %v", err)
	}

	// Initialize connection manager
	var connManager *database.ConnectionManager

	if dbConfig.MongoDB.URI != "" || dbConfig.PostgreSQL.Host != "" {
		if err := dbConfig.Validate(); err != nil {
			log.Fatalf("Invalid database configuration: %v", err)
		}

		appLogger.Info("Initializing database connections", logger.Fields{
			"mongodb_configured":    dbConfig.MongoDB.URI != "",
			"postgresql_configured": dbConfig.PostgreSQL.Host != "",
		})

		connManager = database.NewConnectionManager(&dbConfig, appLogger)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if dbConfig.MongoDB.URI != "" {
			if err := connManager.ConnectMongoDB(ctx); err != nil {
				log.Fatalf("Failed to connect to MongoDB: %v", err)
			}
		}

		if dbConfig.PostgreSQL.Host != "" {
			if err := connManager.ConnectPostgreSQL(ctx); err != nil {
				log.Fatalf("Failed to connect to PostgreSQL: %v", err)
			}
		}

		appLogger.Info("Database connections initialized successfully", logger.Fields{})
	} else {
		appLogger.Info("No database configuration provided, using in-memory repositories", logger.Fields{})
	}

	// Initialize auth infrastructure
	jwksURL := fmt.Sprintf("%s/.well-known/jwks.json", cfg.AuthorizerBaseURL)
	jwksClient := infraAuth.NewJWKSClient(jwksURL, time.Duration(cfg.JWKSCacheDuration)*time.Second, appLogger)
	jwtService := infraAuth.NewJWTService(jwksClient, cfg.AppCode, "authorizer")

	// Initialize repositories
	// PostgreSQL = primary for all entities
	// MongoDB = optional secondary (dual-write) for selected entities only
	//
	// To enable dual-write for an entity, wrap it with dualwrite.New*Repository.
	// To use PostgreSQL only, assign the postgres repo directly.

	hasPostgres := connManager != nil && connManager.GetPostgresGormClient() != nil
	hasMongo := connManager != nil && connManager.GetMongoClient() != nil

	var userProfilesRepo repository.UserProfilesRepository
	var todoRepo repository.TodoRepository
	var credentialRepo repository.CredentialRepository
	var masterDataRepo repository.MasterDataRepository
	var vaultRepo repository.VaultRepository
	var tagRepo repository.TagRepository

	if hasPostgres {
		pgTodoRepo := persistPostgres.NewTodoRepository(connManager.GetPostgresGormClient())
		pgCredentialRepo := persistPostgres.NewCredentialRepository(connManager.GetPostgresGormClient())
		pgUserProfilesRepo := persistPostgres.NewUserProfilesRepository(connManager.GetPostgresGormClient())

		// Todo → PostgreSQL only
		todoRepo = pgTodoRepo

		// MasterData → PostgreSQL only
		masterDataRepo = persistPostgres.NewMasterDataRepository(connManager.GetPostgresGormClient())

		vaultRepo = persistPostgres.NewVaultRepository(connManager.GetPostgresGormClient())

		tagRepo = persistPostgres.NewTagRepository(connManager.GetPostgresGormClient())

		if hasMongo {
			mongoUserProfilesRepo := persistMongo.NewUserProfilesRepository(connManager.GetMongoClient(), dbConfig.MongoDB.Database)
			mongoCredentialRepo := persistMongo.NewCredentialRepository(connManager.GetMongoClient(), dbConfig.MongoDB.Database)

			// Credential → dual-write (PostgreSQL + MongoDB denormalized)
			denormalizedWriter := dualwriteCredential.NewDenormalizedWriter(
				connManager.GetMongoClient(),
				dbConfig.MongoDB.Database,
				vaultRepo,
				masterDataRepo,
				tagRepo,
				pgUserProfilesRepo,
				appLogger,
			)
			credentialRepo = dualwriteCredential.NewRepository(pgCredentialRepo, mongoCredentialRepo, denormalizedWriter, appLogger)

			// UserProfiles → dual-write (PostgreSQL + MongoDB)
			userProfilesRepo = dualwrite.NewUserProfilesRepository(pgUserProfilesRepo, mongoUserProfilesRepo, appLogger)

			appLogger.Info("Repositories initialized", logger.Fields{
				"todo":         "postgres",
				"credential":   "postgres + mongodb",
				"userProfiles": "postgres + mongodb",
			})
		} else {
			credentialRepo = pgCredentialRepo
			userProfilesRepo = pgUserProfilesRepo
			appLogger.Info("Repositories initialized (PostgreSQL only, MongoDB not configured)", logger.Fields{})
		}
	} else {
		appLogger.Info("Using in-memory repositories", logger.Fields{})
		todoRepo = memory.NewTodoRepository()
		credentialRepo = memory.NewCredentialRepository()
		userProfilesRepo = memory.NewUserProfilesRepository()
		masterDataRepo = memory.NewMasterDataRepository()
		vaultRepo = memory.NewVaultRepository()
		tagRepo = memory.NewTagRepository()
	}

	// Initialize use cases
	loginUC := authUseCase.NewLoginUseCase(cfg.AuthorizerBaseURL, cfg.AppCallbackURL)
	callbackUC := authUseCase.NewCallbackUseCase(jwtService)
	logoutUC := authUseCase.NewLogoutUseCase(cfg.AuthorizerBaseURL, cfg.AppCallbackURL)

	todoListUC := todoUseCase.NewListUseCase(todoRepo)

	credentialListUC := credentialUseCase.NewListUseCase(credentialRepo)
	credentialCreateUC := credentialUseCase.NewCreateUseCase(credentialRepo, userProfilesRepo, tagRepo, encryptionSvc, appLogger)

	userProfilesCreateUC := userProfilesUseCase.NewCreateUseCase(userProfilesRepo, appLogger)
	userProfilesGetByUserID := userProfilesUseCase.NewGetByUserIDUseCase(userProfilesRepo, appLogger)
	userProfilesSetupMasterPasswordUC := userProfilesUseCase.NewSetupMasterPasswordUseCase(userProfilesRepo, appLogger)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(loginUC, callbackUC, logoutUC, appLogger)
	todoHandler := handler.NewTodoHandler(todoListUC)
	credentialHandler := handler.NewCredentialHandler(credentialListUC, credentialCreateUC)
	userProfilesHandler := handler.NewUserProfilesHandler(userProfilesCreateUC, userProfilesGetByUserID, userProfilesSetupMasterPasswordUC)

	// Initialize master data use case and handler
	masterDataListUC := masterDataUseCase.NewListUseCase(masterDataRepo, appLogger)
	masterDataHandler := handler.NewMasterDataHandler(masterDataListUC, cfg)

	// Initialize menu use case and handler
	menuListUC := menuUseCase.NewListUseCase(masterDataRepo)
	menuHandler := handler.NewMenuHandler(menuListUC, cfg)

	// Initialize vault use case and handler
	vaultCreateUC := vaultUseCase.NewCreateUseCase(vaultRepo, appLogger)
	vaultListUC := vaultUseCase.NewListUseCase(vaultRepo, appLogger)
	vaultHandler := handler.NewVaultHandler(vaultCreateUC, vaultListUC)

	// Initialize health handler
	var healthHandler *handler.HealthHandler
	if connManager != nil {
		healthChecker := database.NewHealthChecker(connManager, appLogger)
		healthHandler = handler.NewHealthHandler(healthChecker)
	} else {
		healthHandler = handler.NewHealthHandler(nil)
	}

	// Setup router
	r := router.Setup(&router.RouterConfig{
		AuthHandler:         authHandler,
		MasterDataHandler:   masterDataHandler,
		UserProfilesHandler: userProfilesHandler,
		TodoHandler:         todoHandler,
		CredentialHandler:   credentialHandler,
		MenuHandler:         menuHandler,
		VaultHandler:        vaultHandler,
		HealthHandler:       healthHandler,
		AuthMiddleware:      middleware.AuthMiddleware(jwtService, appLogger),
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "4040"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal (Ctrl+C or kill)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down...", logger.Fields{})

	// 1. Shutdown HTTP server first — releases the port immediately
	httpCtx, httpCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer httpCancel()
	if err := srv.Shutdown(httpCtx); err != nil {
		appLogger.Error("HTTP server shutdown error", logger.Fields{"error": err.Error()})
	} else {
		appLogger.Info("HTTP server stopped", logger.Fields{})
	}

	// 2. Then close database connections
	if connManager != nil {
		dbCtx, dbCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer dbCancel()
		if err := connManager.Shutdown(dbCtx); err != nil {
			appLogger.Error("Database shutdown error", logger.Fields{"error": err.Error()})
		} else {
			appLogger.Info("Database connections closed", logger.Fields{})
		}
	}

	appLogger.Info("Server exited", logger.Fields{})
}

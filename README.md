# StackForge

A Go project following Clean Architecture with **layer-first** structure, enforcing the dependency rule where dependencies point strictly inward.

## Architecture Overview

The application is organized into four concentric layers:

```
┌─────────────────────────────────────────────────────────┐
│                    Infrastructure                        │
│  (Frameworks, Drivers, DB, External Services)           │
│  ┌───────────────────────────────────────────────────┐  │
│  │              Delivery (Interface Adapters)        │  │
│  │  (HTTP Handlers, Middleware, Routing)             │  │
│  │  ┌─────────────────────────────────────────────┐  │  │
│  │  │         Use Case (Application Logic)        │  │  │
│  │  │  (Business Rules, Orchestration)            │  │  │
│  │  │  ┌───────────────────────────────────────┐  │  │  │
│  │  │  │   Domain (Enterprise Logic)           │  │  │  │
│  │  │  │  (Entities, Repository Interfaces)    │  │  │  │
│  │  │  └───────────────────────────────────────┘  │  │  │
│  │  └─────────────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘

Dependencies flow inward: Infrastructure → Delivery → Use Case → Domain
```

## Project Structure

```
cmd/
  api/
    main.go                          # Application entry point, dependency wiring

internal/
  domain/                            # Enterprise Business Rules
    entity/                          # Domain entities (pure data structures)
      user.go
      todo.go
      credential.go
      auth.go                        # Auth-related entities (Claims, etc.)
    repository/                      # Repository interfaces
      todo_repository.go
      credential_repository.go

  usecase/                           # Application Business Rules
    auth/                            # Authentication use cases
      interface.go                   # Use case interfaces
      login_usecase.go
      callback_usecase.go
      logout_usecase.go
    todo/                            # Todo use cases
      interface.go
      list_usecase.go
    credential/                      # Credential use cases
      interface.go
      list_usecase.go

  delivery/                          # Interface Adapters
    http/
      handler/                       # HTTP handlers
        auth_handler.go
        todo_handler.go
        credential_handler.go
        health_handler.go
      middleware/                    # HTTP middleware
        auth.go                      # JWT validation middleware
        rbac.go                      # Role-based access control
        claims.go                    # Claims extraction
        request_id.go                # Request ID middleware
      router/                        # Route definitions
        router.go                    # Main router setup
        auth_routes.go               # Auth route registration
        api_routes.go                # Protected API routes

  infrastructure/                    # Frameworks & Drivers
    persistence/                     # Database implementations
      memory/                        # In-memory implementations
        todo_repository.go
        credential_repository.go
      postgres/                      # PostgreSQL implementations
        credential_repository.go
    auth/                            # Auth infrastructure
      jwks_client.go                 # JWKS fetching and caching
      jwt_service.go                 # JWT validation
    config/                          # Configuration
      config.go
    logger/                          # Logging
      logger.go

  pkg/                               # Shared utilities
    response/                        # HTTP response helpers
      response.go
    validator/                       # Validation helpers
      validator.go

migrations/                          # Database migrations
tests/                               # Integration tests
```

## Layer Responsibilities

### Domain Layer (`internal/domain/`)

The innermost layer containing enterprise business rules. This layer:
- Defines core business entities as pure data structures
- Defines repository interfaces for data persistence
- Has **no dependencies** on other layers
- Contains no framework-specific code

**Key principle:** Domain entities are framework-agnostic and contain no business logic methods.

### Use Case Layer (`internal/usecase/`)

The application business rules layer. This layer:
- Implements business logic and orchestration
- Depends **only on domain interfaces**
- Defines use case interfaces for dependency injection
- Coordinates between domain entities and repositories

**Key principle:** Use cases are testable in isolation by mocking repository interfaces.

### Delivery Layer (`internal/delivery/`)

The interface adapters layer. This layer:
- Translates HTTP requests into use case calls
- Depends **only on use case and domain interfaces**
- Contains HTTP handlers, middleware, and routing
- Handles request/response formatting

**Key principle:** Handlers are thin and delegate all business logic to use cases.

### Infrastructure Layer (`internal/infrastructure/`)

The outermost layer containing frameworks and drivers. This layer:
- Implements domain repository interfaces
- Provides external service integrations (JWT, JWKS)
- Manages configuration and logging
- Depends **only on domain interfaces**

**Key principle:** Infrastructure implementations can be swapped without affecting business logic.

## Getting Started

### Environment Variables

Create a `.env` file in the project root with the following configuration:

```bash
# Server Configuration
PORT=4040
GIN_MODE=debug

# MongoDB Configuration
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=stackforge
MONGODB_TIMEOUT=10s

# PostgreSQL Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=secret
POSTGRES_DATABASE=stackforge
POSTGRES_SSLMODE=disable

# Authentication Configuration
AUTHORIZER_BASE_URL=http://authorizer.localdev.me:4000
APP_CODE=STACKFORGE
APP_CALLBACK_URL=http://localhost:4040/auth/callback
JWKS_CACHE_DURATION=3600
```

### Configuration Details

#### Server Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `4040` |
| `GIN_MODE` | Gin framework mode (`debug`, `release`) | `debug` |

#### Database Configuration

**MongoDB:**

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `MONGODB_URI` | MongoDB connection URI (format: `mongodb://` or `mongodb+srv://`) | Yes | - |
| `MONGODB_DATABASE` | Database name | No | - |
| `MONGODB_TIMEOUT` | Connection timeout | No | `10s` |

**PostgreSQL:**

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `POSTGRES_HOST` | Database host | Yes | - |
| `POSTGRES_PORT` | Database port | No | `5432` |
| `POSTGRES_USER` | Database user | Yes | - |
| `POSTGRES_PASSWORD` | Database password | Yes | - |
| `POSTGRES_DATABASE` | Database name | Yes | - |
| `POSTGRES_SSLMODE` | SSL mode (`disable`, `require`, `verify-ca`, `verify-full`) | No | `disable` |

**Connection Pooling:**

The application automatically configures connection pooling for optimal performance:

- **MongoDB**: MinPoolSize: 5, MaxPoolSize: 100, MaxConnIdleTime: 5 minutes
- **PostgreSQL**: MaxOpenConns: 100, MaxIdleConns: 10, ConnMaxLifetime: 1 hour

**Retry Logic:**

Database connections use exponential backoff retry logic:
- Max Attempts: 3
- Initial Delay: 1 second
- Backoff Multiplier: 2 (exponential)
- Max Delay: 10 seconds

#### Authentication Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `AUTHORIZER_BASE_URL` | Base URL of the centralized Authorizer Service | `http://authorizer.localdev.me` |
| `APP_CODE` | Application identifier for JWT audience validation | `STACKFORGE` |
| `APP_CALLBACK_URL` | Full callback URL for SSO authentication | `http://localhost:4040/auth/callback` |
| `JWKS_CACHE_DURATION` | Public key cache duration in seconds | `3600` |

### Running the Application

```bash
go run cmd/api/main.go
```

## Database Setup

The application supports dual database configuration with MongoDB and PostgreSQL. Both databases are optional - if not configured, the application will use in-memory repositories as a fallback.

### Prerequisites

**MongoDB:**
```bash
# Using Docker
docker run -d -p 27017:27017 --name mongodb mongo:latest

# Or using Docker Compose
docker-compose up -d mongodb
```

**PostgreSQL:**
```bash
# Using Docker
docker run -d -p 5432:5432 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=secret \
  -e POSTGRES_DB=stackforge \
  --name postgres postgres:latest

# Or using Docker Compose
docker-compose up -d postgres
```

### Database Configuration

Configure database connections using environment variables in your `.env` file:

```env
# MongoDB Configuration
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=stackforge
MONGODB_TIMEOUT=10s

# PostgreSQL Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=secret
POSTGRES_DATABASE=stackforge
POSTGRES_SSLMODE=disable
```

### Health Checks

The application includes database health checks accessible via the health endpoint:

```bash
curl http://localhost:4040/health
```

Response includes database status:
```json
{
  "status": "healthy",
  "databases": {
    "mongodb": {
      "status": "healthy",
      "latency_ms": 5
    },
    "postgresql": {
      "status": "healthy",
      "latency_ms": 3
    }
  }
}
```

### Migration from In-Memory Repositories

The application uses a graceful fallback mechanism:

1. **With Database Configuration**: If MongoDB/PostgreSQL environment variables are provided, the application uses database repositories
2. **Without Database Configuration**: If environment variables are not set, the application falls back to in-memory repositories

**Migration Steps:**

1. **Add database environment variables** to your `.env` file
2. **Start database services** (MongoDB and/or PostgreSQL)
3. **Restart the application** - it will automatically detect and use database repositories
4. **Verify connection** by checking the health endpoint

**Example Migration:**

```bash
# Before: Running with in-memory repositories
go run cmd/api/main.go

# After: Add database config to .env
echo "MONGODB_URI=mongodb://localhost:27017" >> .env
echo "POSTGRES_HOST=localhost" >> .env
echo "POSTGRES_USER=postgres" >> .env
echo "POSTGRES_PASSWORD=secret" >> .env
echo "POSTGRES_DATABASE=stackforge" >> .env

# Restart application - now using database repositories
go run cmd/api/main.go
```

### Troubleshooting

**Connection Failures:**

If you see connection errors in the logs:

1. Verify databases are running:
   ```bash
   # MongoDB
   docker ps | grep mongodb
   
   # PostgreSQL
   docker ps | grep postgres
   ```

2. Check connection details in `.env` file
3. Verify network connectivity:
   ```bash
   # MongoDB
   telnet localhost 27017
   
   # PostgreSQL
   telnet localhost 5432
   ```

4. Check application logs for detailed error messages

**For more detailed database configuration and troubleshooting, see [Database Package Documentation](internal/infrastructure/database/README.md).**

## Authentication

This service uses SSO authentication via a centralized Authorizer Service. JWT tokens are signed with RS256 asymmetric cryptography and validated using public keys from the JWKS endpoint.

### Authentication Flow

1. User accesses a protected endpoint without a token
2. Service redirects to Authorizer login page
3. User authenticates with Authorizer Service
4. Authorizer redirects back with JWT token
5. Service validates token and sets secure cookie
6. User can now access protected endpoints

### Auth Endpoints

- **Login:** `GET /auth/login` - Redirects to Authorizer login page
- **Callback:** `GET /auth/callback?token=<jwt>` - Processes SSO callback and sets auth cookie
- **Logout:** `GET /auth/logout` - Clears auth cookie and redirects to Authorizer logout

### Public Endpoints

- **Health:** `GET /health` - Health check (no auth required)

### Protected Endpoints

- **Todo List:** `GET /api/todos` - Requires valid JWT token

## Protecting Endpoints with Middleware

### Basic Authentication

Apply the `AuthMiddleware` to require valid JWT tokens:

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/mafzaidi/stackforge/internal/delivery/http/middleware"
)

func setupRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
    // Public routes (no auth required)
    router.GET("/health", healthHandler)
    
    // Protected routes (auth required)
    api := router.Group("/api")
    api.Use(authMiddleware)
    {
        api.GET("/todos", todoHandler)
    }
}
```

### Role-Based Access Control

Use `RequireRole` to restrict access to users with specific roles:

```go
import "github.com/mafzaidi/stackforge/internal/delivery/http/middleware"

// Require "admin" role
router.DELETE("/api/todos/:id", 
    authMiddleware,
    middleware.RequireRole("admin"),
    deleteTodoHandler,
)

// Require any of multiple roles
router.POST("/api/todos", 
    authMiddleware,
    middleware.RequireAnyRole("admin", "editor"),
    createTodoHandler,
)
```

### Permission-Based Access Control

Use `RequirePermission` to restrict access based on specific permissions:

```go
import "github.com/mafzaidi/stackforge/internal/delivery/http/middleware"

// Require specific permission
router.GET("/api/todos", 
    authMiddleware,
    middleware.RequirePermission("todo.read"),
    listTodosHandler,
)

// Require any of multiple permissions
router.POST("/api/todos", 
    authMiddleware,
    middleware.RequireAnyPermission("todo.write", "todo.admin"),
    createTodoHandler,
)
```

### Accessing User Claims in Handlers

Extract authenticated user information from the request context:

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/mafzaidi/stackforge/internal/delivery/http/middleware"
)

func myHandler(c *gin.Context) {
    // Get claims (returns error if not authenticated)
    claims, err := middleware.GetClaims(c)
    if err != nil {
        c.JSON(401, gin.H{"error": "Unauthorized"})
        return
    }
    
    // Access user information
    userID := claims.Subject
    username := claims.Username
    email := claims.Email
    
    // Check roles and permissions programmatically
    hasAdminRole := claims.HasRole("STACKFORGE", "admin")
    canWrite := claims.HasPermission("STACKFORGE", "todo.write")
    hasWildcard := claims.HasAnyPermission("STACKFORGE")
    
    c.JSON(200, gin.H{
        "user_id": userID,
        "username": username,
        "email": email,
    })
}
```

### Authorization Claim Structure

JWT tokens contain authorization claims with roles and permissions per application:

```json
{
  "authorization": [
    {
      "app": "GLOBAL",
      "roles": ["admin"],
      "permissions": ["*"]
    },
    {
      "app": "STACKFORGE",
      "roles": ["user"],
      "permissions": ["todo.read", "todo.write"]
    }
  ]
}
```

- **GLOBAL** authorizations apply to all applications
- **App-specific** authorizations apply only to the specified app
- **Wildcard permission** (`*`) grants all permissions
- Middleware checks both GLOBAL and app-specific authorizations

## Adding New Features

When adding a new feature to the application, follow the Clean Architecture layers:

### 1. Define Domain Layer

Create entities and repository interfaces:

```go
// internal/domain/entity/myfeature.go
package entity

type MyFeature struct {
    ID        string
    Name      string
    CreatedAt time.Time
}

// internal/domain/repository/myfeature_repository.go
package repository

import (
    "context"
    "github.com/mafzaidi/stackforge/internal/domain/entity"
)

type MyFeatureRepository interface {
    GetByID(ctx context.Context, id string) (*entity.MyFeature, error)
    List(ctx context.Context) ([]*entity.MyFeature, error)
    Create(ctx context.Context, f *entity.MyFeature) error
}
```

### 2. Implement Use Cases

Create use case interfaces and implementations:

```go
// internal/usecase/myfeature/interface.go
package myfeature

import (
    "context"
    "github.com/mafzaidi/stackforge/internal/domain/entity"
)

type ListUseCase interface {
    Execute(ctx context.Context) ([]*entity.MyFeature, error)
}

// internal/usecase/myfeature/list_usecase.go
package myfeature

import (
    "context"
    "github.com/mafzaidi/stackforge/internal/domain/entity"
    "github.com/mafzaidi/stackforge/internal/domain/repository"
)

type listUseCase struct {
    repo repository.MyFeatureRepository
}

func NewListUseCase(repo repository.MyFeatureRepository) ListUseCase {
    return &listUseCase{repo: repo}
}

func (uc *listUseCase) Execute(ctx context.Context) ([]*entity.MyFeature, error) {
    return uc.repo.List(ctx)
}
```

### 3. Create HTTP Handler

Implement the delivery layer handler:

```go
// internal/delivery/http/handler/myfeature_handler.go
package handler

import (
    "github.com/gin-gonic/gin"
    "github.com/mafzaidi/stackforge/internal/usecase/myfeature"
    "github.com/mafzaidi/stackforge/internal/pkg/response"
)

type MyFeatureHandler struct {
    listUC myfeature.ListUseCase
}

func NewMyFeatureHandler(listUC myfeature.ListUseCase) *MyFeatureHandler {
    return &MyFeatureHandler{listUC: listUC}
}

func (h *MyFeatureHandler) List(c *gin.Context) {
    features, err := h.listUC.Execute(c.Request.Context())
    if err != nil {
        response.InternalServerError(c, "Failed to list features")
        return
    }
    response.Success(c, gin.H{"items": features})
}
```

### 4. Implement Infrastructure

Create repository implementation:

```go
// internal/infrastructure/persistence/memory/myfeature_repository.go
package memory

import (
    "context"
    "github.com/mafzaidi/stackforge/internal/domain/entity"
    "github.com/mafzaidi/stackforge/internal/domain/repository"
)

type myFeatureRepository struct {
    data map[string]*entity.MyFeature
}

func NewMyFeatureRepository() repository.MyFeatureRepository {
    return &myFeatureRepository{
        data: make(map[string]*entity.MyFeature),
    }
}

func (r *myFeatureRepository) List(ctx context.Context) ([]*entity.MyFeature, error) {
    // Implementation...
}
```

### 5. Wire Dependencies in main.go

Connect all layers using dependency injection:

```go
// cmd/api/main.go

// Infrastructure layer
myFeatureRepo := memory.NewMyFeatureRepository()

// Use case layer
listMyFeatureUC := myfeature.NewListUseCase(myFeatureRepo)

// Delivery layer
myFeatureHandler := handler.NewMyFeatureHandler(listMyFeatureUC)

// Register routes
router.GET("/api/myfeatures", authMiddleware, myFeatureHandler.List)
```

## Dependency Injection Pattern

The application uses **constructor-based dependency injection** to wire layers together while maintaining the dependency rule.

### Dependency Flow

```
main.go (wiring)
    ↓
Infrastructure Implementations (concrete types)
    ↓
Use Cases (depend on repository interfaces)
    ↓
Handlers (depend on use case interfaces)
    ↓
Router (receives handlers)
```

### Key Principles

1. **Interfaces in Inner Layers**: Domain layer defines repository interfaces; use case layer defines use case interfaces
2. **Implementations in Outer Layers**: Infrastructure layer implements repository interfaces; use case layer implements use case interfaces
3. **Constructor Injection**: All dependencies are passed through constructor functions (`New*()`)
4. **Dependency Inversion**: Outer layers depend on inner layer interfaces, not concrete types

### Example Wiring

```go
// cmd/api/main.go
func main() {
    // Load configuration
    cfg := config.LoadConfig()
    
    // Initialize logger
    logger := logger.NewLogger()
    
    // Infrastructure layer - external services
    jwksClient := auth.NewJWKSClient(cfg.JWKSURL, cfg.JWKSCacheDuration, logger)
    jwtService := auth.NewJWTService(jwksClient, cfg.AppCode, cfg.Issuer)
    
    // Infrastructure layer - repositories
    todoRepo := memory.NewTodoRepository()
    credentialRepo := postgres.NewCredentialRepository(cfg.DBConnStr)
    
    // Use case layer - inject repository interfaces
    loginUC := auth.NewLoginUseCase(cfg.AuthorizerBaseURL)
    callbackUC := auth.NewCallbackUseCase(jwtService)
    logoutUC := auth.NewLogoutUseCase(cfg.AuthorizerBaseURL)
    listTodoUC := todo.NewListUseCase(todoRepo)
    listCredentialUC := credential.NewListUseCase(credentialRepo)
    
    // Delivery layer - inject use case interfaces
    authHandler := handler.NewAuthHandler(loginUC, callbackUC, logoutUC)
    todoHandler := handler.NewTodoHandler(listTodoUC)
    credentialHandler := handler.NewCredentialHandler(listCredentialUC)
    healthHandler := handler.NewHealthHandler()
    
    // Middleware - inject infrastructure services
    authMiddleware := middleware.AuthMiddleware(jwtService, logger)
    
    // Router - inject handlers and middleware
    router := router.Setup(&router.RouterConfig{
        AuthHandler:       authHandler,
        TodoHandler:       todoHandler,
        CredentialHandler: credentialHandler,
        HealthHandler:     healthHandler,
        AuthMiddleware:    authMiddleware,
    })
    
    // Start server
    router.Run(":" + cfg.Port)
}
```

### Benefits

- **Testability**: Each layer can be tested in isolation by mocking interfaces
- **Flexibility**: Implementations can be swapped without changing business logic
- **Maintainability**: Clear dependency boundaries make the codebase easier to understand
- **Decoupling**: Inner layers have no knowledge of outer layer implementations

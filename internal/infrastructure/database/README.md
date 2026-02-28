# Database Package

Package database menyediakan manajemen koneksi untuk dual database (MongoDB dan PostgreSQL) dengan fitur connection pooling, health checks, retry logic, dan graceful shutdown.

## Fitur

- **Dual Database Support**: Mengelola koneksi MongoDB dan PostgreSQL secara bersamaan
- **Connection Pooling**: Konfigurasi pool yang optimal untuk performa tinggi
- **Retry Logic**: Exponential backoff untuk menangani kegagalan koneksi sementara
- **Health Checks**: Monitoring kesehatan koneksi database secara real-time
- **Graceful Shutdown**: Penutupan koneksi yang aman saat aplikasi berhenti
- **Comprehensive Logging**: Log detail untuk monitoring dan debugging

## Komponen Utama

### ConnectionManager

Mengelola lifecycle koneksi database dengan connection pooling dan retry logic.

```go
type ConnectionManager struct {
    mongoClient    *mongo.Client
    postgresClient *sql.DB
    config         *config.DatabaseConfig
    logger         *logger.Logger
    shutdownOnce   sync.Once
}
```

### HealthChecker

Memeriksa status kesehatan koneksi database dengan timeout yang dikonfigurasi.

```go
type HealthChecker struct {
    connManager *ConnectionManager
    logger      *logger.Logger
}
```

### ShutdownHandler

Menangani graceful shutdown dengan mendengarkan signal SIGINT/SIGTERM.

```go
type ShutdownHandler struct {
    connManager *ConnectionManager
    logger      *logger.Logger
}
```

## Konfigurasi

### Environment Variables

#### MongoDB
- `MONGODB_URI` (required): MongoDB connection URI (format: `mongodb://` atau `mongodb+srv://`)
- `MONGODB_DATABASE` (optional): Nama database
- `MONGODB_TIMEOUT` (optional, default: `10s`): Connection timeout

#### PostgreSQL
- `POSTGRES_HOST` (required): Database host
- `POSTGRES_PORT` (optional, default: `5432`): Database port
- `POSTGRES_USER` (required): Database user
- `POSTGRES_PASSWORD` (required): Database password
- `POSTGRES_DATABASE` (required): Database name
- `POSTGRES_SSLMODE` (optional, default: `disable`): SSL mode (`disable`, `require`, `verify-ca`, `verify-full`)

### Contoh .env File

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

### Connection Pool Configuration

#### MongoDB
- **MinPoolSize**: 5 koneksi
- **MaxPoolSize**: 100 koneksi
- **MaxConnIdleTime**: 5 menit

#### PostgreSQL
- **MaxOpenConns**: 100 koneksi
- **MaxIdleConns**: 10 koneksi
- **ConnMaxLifetime**: 1 jam
- **ConnMaxIdleTime**: 5 menit

### Retry Configuration

- **Max Attempts**: 3 kali
- **Initial Delay**: 1 detik
- **Backoff Multiplier**: 2 (exponential)
- **Max Delay**: 10 detik

## Usage Examples

### Basic Setup

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/mafzaidi/stackforge/internal/infrastructure/config"
    "github.com/mafzaidi/stackforge/internal/infrastructure/database"
    "github.com/mafzaidi/stackforge/internal/infrastructure/logger"
)

func main() {
    // Initialize logger
    log := logger.New()

    // Load database configuration
    dbConfig, err := config.LoadDatabaseConfig()
    if err != nil {
        log.Fatal("Failed to load database config", logger.Fields{"error": err.Error()})
    }

    // Validate configuration
    if err := dbConfig.Validate(); err != nil {
        log.Fatal("Invalid database config", logger.Fields{"error": err.Error()})
    }

    // Create connection manager
    connManager := database.NewConnectionManager(&dbConfig, log)

    // Connect to databases with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Connect to MongoDB
    if err := connManager.ConnectMongoDB(ctx); err != nil {
        log.Fatal("Failed to connect to MongoDB", logger.Fields{"error": err.Error()})
    }

    // Connect to PostgreSQL
    if err := connManager.ConnectPostgreSQL(ctx); err != nil {
        log.Fatal("Failed to connect to PostgreSQL", logger.Fields{"error": err.Error()})
    }

    // Get database clients
    mongoClient := connManager.GetMongoClient()
    postgresClient := connManager.GetPostgresClient()

    // Use clients for database operations
    // ...

    // Setup graceful shutdown
    shutdownHandler := database.NewShutdownHandler(connManager, log)
    go shutdownHandler.WaitForShutdown(ctx, cancel)

    // Application logic
    // ...
}
```

### Health Checks

```go
// Create health checker
healthChecker := database.NewHealthChecker(connManager, log)

// Check all databases
ctx := context.Background()
healthStatus := healthChecker.CheckHealth(ctx)

// Check individual databases
mongoHealth := healthChecker.CheckMongoDB(ctx)
postgresHealth := healthChecker.CheckPostgreSQL(ctx)

// Example response
fmt.Printf("MongoDB Status: %s (Latency: %v)\n", 
    mongoHealth.Status, mongoHealth.Latency)
fmt.Printf("PostgreSQL Status: %s (Latency: %v)\n", 
    postgresHealth.Status, postgresHealth.Latency)
```

### Using Database Clients

```go
// MongoDB operations
mongoClient := connManager.GetMongoClient()
collection := mongoClient.Database("stackforge").Collection("todos")

// Insert document
result, err := collection.InsertOne(ctx, bson.M{
    "title": "Example Todo",
    "completed": false,
})

// PostgreSQL operations
postgresClient := connManager.GetPostgresClient()

// Query
rows, err := postgresClient.QueryContext(ctx, 
    "SELECT id, title FROM todos WHERE completed = $1", false)
defer rows.Close()

// Execute
_, err = postgresClient.ExecContext(ctx,
    "UPDATE todos SET completed = $1 WHERE id = $2", true, 1)
```

### Graceful Shutdown

```go
// Setup shutdown handler
shutdownHandler := database.NewShutdownHandler(connManager, log)

// Wait for shutdown signal in goroutine
ctx, cancel := context.WithCancel(context.Background())
go shutdownHandler.WaitForShutdown(ctx, cancel)

// Or manually trigger shutdown
shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
defer shutdownCancel()

if err := connManager.Shutdown(shutdownCtx); err != nil {
    log.Error("Shutdown failed", logger.Fields{"error": err.Error()})
}
```

## Error Handling

Package ini menyediakan error types yang spesifik untuk berbagai skenario:

### ConfigError

Error terkait konfigurasi database.

```go
type ConfigError struct {
    Field   string // Field yang menyebabkan error
    Message string // Deskripsi error
}

// Example
err := &database.ConfigError{
    Field:   "MONGODB_URI",
    Message: "invalid URI format",
}
// Output: "configuration error in field 'MONGODB_URI': invalid URI format"
```

### ConnectionError

Error terkait koneksi database.

```go
type ConnectionError struct {
    Database string // Database type (mongodb/postgresql)
    Cause    error  // Underlying error
    Retries  int    // Jumlah retry yang dilakukan
}

// Example
err := &database.ConnectionError{
    Database: "mongodb",
    Cause:    errors.New("connection refused"),
    Retries:  3,
}
// Output: "failed to connect to mongodb after 3 retries: connection refused"
```

### HealthCheckError

Error terkait health check.

```go
type HealthCheckError struct {
    Database string // Database type
    Cause    error  // Underlying error
}

// Example
err := &database.HealthCheckError{
    Database: "postgresql",
    Cause:    errors.New("timeout"),
}
// Output: "health check failed for postgresql: timeout"
```

### ShutdownError

Error terkait graceful shutdown.

```go
type ShutdownError struct {
    Database string // Database type
    Cause    error  // Underlying error
}

// Example
err := &database.ShutdownError{
    Database: "mongodb",
    Cause:    errors.New("context deadline exceeded"),
}
// Output: "shutdown failed for mongodb: context deadline exceeded"
```

### Error Handling Best Practices

```go
// 1. Check for specific error types
if err := connManager.ConnectMongoDB(ctx); err != nil {
    var connErr *database.ConnectionError
    if errors.As(err, &connErr) {
        log.Error("Connection failed", logger.Fields{
            "database": connErr.Database,
            "retries":  connErr.Retries,
        })
    }
}

// 2. Unwrap errors to get root cause
if err := connManager.ConnectPostgreSQL(ctx); err != nil {
    rootErr := errors.Unwrap(err)
    log.Error("Root cause", logger.Fields{"error": rootErr.Error()})
}

// 3. Handle validation errors
if err := dbConfig.Validate(); err != nil {
    // Validation errors are descriptive and identify the problematic field
    log.Fatal("Configuration validation failed", logger.Fields{"error": err.Error()})
}

// 4. Handle shutdown errors gracefully
if err := connManager.Shutdown(ctx); err != nil {
    // Log but don't panic - application should exit anyway
    log.Error("Shutdown completed with errors", logger.Fields{"error": err.Error()})
}
```

## Logging

Package ini menghasilkan log detail untuk monitoring dan debugging:

### Connection Logs

```
INFO  Attempting database connection database=mongodb attempt=1 max_attempts=3
INFO  Successfully connected to MongoDB database=stackforge pool_config={"min_pool_size":5,"max_pool_size":100,"max_conn_idle_time":"5m"}
INFO  Starting PostgreSQL connection host=localhost port=5432 database=stackforge
INFO  Successfully connected to PostgreSQL host=localhost port=5432 database=stackforge pool_config={"max_open_conns":100,"max_idle_conns":10,"conn_max_lifetime":"1h","conn_max_idle_time":"5m"}
```

### Retry Logs

```
WARN  Connection attempt failed database=mongodb attempt=1 max_attempts=3 error="connection refused"
INFO  Retrying connection database=mongodb delay=1s
INFO  Successfully connected after retries database=mongodb attempts=2
```

### Health Check Logs

```
INFO  MongoDB health check passed latency=5ms
WARN  PostgreSQL health check failed error="timeout" latency=5s
```

### Shutdown Logs

```
INFO  Received shutdown signal signal=interrupt
INFO  Starting graceful shutdown of database connections
INFO  Closing MongoDB connection
INFO  Successfully closed MongoDB connection
INFO  Closing PostgreSQL connection
INFO  Successfully closed PostgreSQL connection
INFO  Successfully completed graceful shutdown of all database connections
```

### Security Note

**Password tidak pernah di-log**. Connection string dan konfigurasi yang mengandung password akan di-sanitize sebelum logging.

## Testing

### Unit Tests

```bash
# Run all tests
go test ./internal/infrastructure/database/...

# Run with coverage
go test -cover ./internal/infrastructure/database/...

# Run specific test
go test -run TestConnectionManager_ConnectMongoDB ./internal/infrastructure/database/
```

### Integration Tests

Integration tests menggunakan testcontainers untuk spin up database instances:

```bash
# Run integration tests (requires Docker)
go test -tags=integration ./internal/infrastructure/database/...
```

## Performance Considerations

1. **Connection Pooling**: Pool size dikonfigurasi untuk menangani high concurrency tanpa exhausting connections
2. **Health Check Timeout**: 5 detik timeout mencegah blocking pada health checks
3. **Retry Backoff**: Exponential backoff mencegah overwhelming failed databases
4. **Graceful Shutdown**: 10 detik timeout per database memastikan aplikasi tidak hang

## Security Considerations

1. **Password Handling**: Password tidak pernah di-log atau di-expose dalam error messages
2. **SSL/TLS Support**: PostgreSQL mendukung berbagai SSL modes untuk secure connections
3. **Connection String Validation**: URI dan connection strings divalidasi sebelum digunakan
4. **Error Messages**: Error messages tidak mengandung informasi sensitif

## Troubleshooting

### Connection Failures

**Problem**: `failed to connect to mongodb after 3 retries`

**Solutions**:
- Verify MongoDB is running: `docker ps` atau `systemctl status mongod`
- Check MONGODB_URI format: harus dimulai dengan `mongodb://` atau `mongodb+srv://`
- Verify network connectivity: `telnet localhost 27017`
- Check firewall rules

**Problem**: `failed to connect to postgresql: connection refused`

**Solutions**:
- Verify PostgreSQL is running: `pg_isready -h localhost -p 5432`
- Check POSTGRES_HOST dan POSTGRES_PORT
- Verify credentials (POSTGRES_USER, POSTGRES_PASSWORD)
- Check pg_hba.conf untuk authentication settings

### Health Check Failures

**Problem**: Health check returns `unhealthy` status

**Solutions**:
- Check database logs untuk errors
- Verify database is not overloaded (check CPU, memory, connections)
- Increase health check timeout jika database response lambat
- Check network latency antara application dan database

### Validation Errors

**Problem**: `invalid MongoDB URI format`

**Solution**: URI harus dimulai dengan `mongodb://` atau `mongodb+srv://`

**Problem**: `invalid port: must be between 1 and 65535`

**Solution**: Verify POSTGRES_PORT value dalam range yang valid

**Problem**: `invalid sslmode`

**Solution**: POSTGRES_SSLMODE harus salah satu dari: `disable`, `require`, `verify-ca`, `verify-full`

### Shutdown Issues

**Problem**: Application hangs on shutdown

**Solutions**:
- Check untuk long-running transactions yang belum selesai
- Verify shutdown timeout cukup (default: 10s per database)
- Check logs untuk shutdown errors
- Force kill jika necessary: `kill -9 <pid>`

## Migration Guide

### From In-Memory to Database

1. **Add environment variables** ke .env file atau environment
2. **Update main.go** untuk initialize ConnectionManager
3. **Update repositories** untuk use database clients instead of in-memory storage
4. **Test thoroughly** dengan integration tests
5. **Deploy gradually** dengan feature flags jika possible

### Example Migration

```go
// Before (in-memory)
todoRepo := memory.NewTodoRepository()

// After (database)
dbConfig, _ := config.LoadDatabaseConfig()
connManager := database.NewConnectionManager(&dbConfig, log)
connManager.ConnectMongoDB(ctx)
mongoClient := connManager.GetMongoClient()
todoRepo := mongodb.NewTodoRepository(mongoClient, "stackforge")
```

## Contributing

Saat menambahkan fitur baru atau melakukan perubahan:

1. Update tests (unit dan integration)
2. Update documentation (README dan godoc comments)
3. Ensure backward compatibility
4. Follow existing code style dan patterns
5. Add logging untuk debugging
6. Handle errors gracefully

## License

Copyright © 2024 Stackforge. All rights reserved.

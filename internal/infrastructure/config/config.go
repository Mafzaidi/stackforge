// Package config provides application configuration management.
//
// This package handles loading and validating configuration from environment
// variables, including database configuration for MongoDB and PostgreSQL,
// authentication settings, and application settings.
//
// # Configuration Loading
//
// Configuration is loaded from environment variables with sensible defaults:
//
//	cfg, err := config.Load()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Database Configuration
//
// Database configuration can be loaded separately:
//
//	dbConfig, err := config.LoadDatabaseConfig()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Validate the configuration
//	if err := dbConfig.Validate(); err != nil {
//	    log.Fatal(err)
//	}
//
// # Environment Variables
//
// MongoDB configuration:
//   - MONGODB_URI: MongoDB connection URI (required)
//   - MONGODB_DATABASE: Database name (required)
//   - MONGODB_TIMEOUT: Connection timeout (default: 10s)
//
// PostgreSQL configuration:
//   - POSTGRES_HOST: Database host (required)
//   - POSTGRES_PORT: Database port (default: 5432)
//   - POSTGRES_USER: Database user (required)
//   - POSTGRES_PASSWORD: Database password (required)
//   - POSTGRES_DATABASE: Database name (required)
//   - POSTGRES_SSLMODE: SSL mode (default: disable)
//
// Application configuration:
//   - PORT: Server port (default: 8080)
//   - GIN_MODE: Gin framework mode (default: release)
//   - AUTHORIZER_BASE_URL: Authorizer service URL
//   - APP_CODE: Application code for authentication
//   - JWKS_CACHE_DURATION: JWKS cache duration in seconds
//   - APP_CALLBACK_URL: OAuth callback URL
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/mafzaidi/stackforge/internal/pkg/validator"
)

// Config holds application configuration.
//
// It implements the domain/service.Config interface and contains all
// configuration needed for the application including server settings,
// authentication, and database configuration.
type Config struct {
	Port      string
	DBConnStr string
	GinMode   string

	// Auth configuration
	AuthorizerBaseURL string // Base URL of Authorizer Service
	AppCode           string // Application code for audience validation
	JWKSCacheDuration int    // JWKS cache duration in seconds
	AppCallbackURL    string // Full callback URL for this service

	// Database configuration
	Database DatabaseConfig
}

// DatabaseConfig holds database connection configurations for both MongoDB and PostgreSQL.
//
// It provides a unified configuration structure for dual database support,
// allowing the application to connect to both databases simultaneously.
type DatabaseConfig struct {
	MongoDB    MongoDBConfig
	PostgreSQL PostgreSQLConfig
}

// MongoDBConfig holds MongoDB connection configuration.
//
// URI should be a valid MongoDB connection string starting with "mongodb://"
// or "mongodb+srv://". Database specifies the database name to use.
// Timeout controls connection and operation timeouts.
//
// Example:
//
//	cfg := MongoDBConfig{
//	    URI:      "mongodb://localhost:27017",
//	    Database: "myapp",
//	    Timeout:  10 * time.Second,
//	}
type MongoDBConfig struct {
	URI      string        // MongoDB connection URI
	Database string        // Database name
	Timeout  time.Duration // Connection timeout
}

// PostgreSQLConfig holds PostgreSQL connection configuration.
//
// All fields except Port and SSLMode are required. Port defaults to 5432
// and SSLMode defaults to "disable". Valid SSLMode values are:
//   - disable: No SSL
//   - require: SSL required but no verification
//   - verify-ca: SSL with CA verification
//   - verify-full: SSL with full verification
//
// Example:
//
//	cfg := PostgreSQLConfig{
//	    Host:     "localhost",
//	    Port:     5432,
//	    User:     "myuser",
//	    Password: "mypassword",
//	    Database: "myapp",
//	    SSLMode:  "disable",
//	}
type PostgreSQLConfig struct {
	Host     string // Database host
	Port     int    // Database port
	User     string // Database user
	Password string // Database password
	Database string // Database name
	SSLMode  string // SSL mode (disable, require, verify-ca, verify-full)
}

// GetAppCode returns the application code.
//
// This method implements the domain/service.Config interface and is used
// for audience validation in JWT authentication.
func (c *Config) GetAppCode() string {
	return c.AppCode
}

// Load loads application configuration from environment variables.
//
// It reads configuration from environment variables and applies default values
// where appropriate. The AUTHORIZER_BASE_URL is validated to ensure it's a
// valid URL format.
//
// Returns a Config instance with all settings loaded, or an error if validation fails.
//
// Example:
//
//	cfg, err := config.Load()
//	if err != nil {
//	    log.Fatalf("Failed to load config: %v", err)
//	}
//	fmt.Printf("Server will run on port %s\n", cfg.Port)
func Load() (*Config, error) {
	authorizerBaseURL := getEnv("AUTHORIZER_BASE_URL", "http://authorizer.localdev.me")

	// Validate AUTHORIZER_BASE_URL format
	if err := validator.ValidateURL(authorizerBaseURL); err != nil {
		return nil, fmt.Errorf("invalid AUTHORIZER_BASE_URL: %w", err)
	}

	return &Config{
		Port:              getEnv("PORT", "8080"),
		DBConnStr:         getEnv("DATABASE_URL", ""),
		GinMode:           getEnv("GIN_MODE", "release"),
		AuthorizerBaseURL: authorizerBaseURL,
		AppCode:           getEnv("APP_CODE", "STACKFORGE"),
		JWKSCacheDuration: getEnvInt("JWKS_CACHE_DURATION", 3600),
		AppCallbackURL:    getEnv("APP_CALLBACK_URL", "http://localhost:4040/auth/callback"),
	}, nil
}

// getEnv retrieves an environment variable value or returns a default.
//
// Parameters:
//   - key: Environment variable name
//   - defaultVal: Default value if environment variable is not set or empty
//
// Returns the environment variable value if set, otherwise the default value.
func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// getEnvInt retrieves an environment variable as an integer or returns a default.
//
// If the environment variable is not set, empty, or cannot be parsed as an
// integer, the default value is returned.
//
// Parameters:
//   - key: Environment variable name
//   - defaultVal: Default value if environment variable is not set or invalid
//
// Returns the parsed integer value if valid, otherwise the default value.
func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

// getEnvDuration retrieves an environment variable as a duration or returns a default.
//
// The environment variable should be in a format parseable by time.ParseDuration
// (e.g., "10s", "5m", "1h"). If the variable is not set, empty, or cannot be
// parsed, the default value is returned.
//
// Parameters:
//   - key: Environment variable name
//   - defaultVal: Default duration if environment variable is not set or invalid
//
// Returns the parsed duration if valid, otherwise the default value.
func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}

// LoadDatabaseConfig loads database configuration from environment variables.
//
// It reads MongoDB and PostgreSQL configuration from their respective environment
// variables and applies default values where appropriate. The returned configuration
// should be validated using the Validate method before use.
//
// Environment variables:
//   - MONGODB_URI: MongoDB connection URI
//   - MONGODB_DATABASE: MongoDB database name
//   - MONGODB_TIMEOUT: Connection timeout (default: 10s)
//   - POSTGRES_HOST: PostgreSQL host
//   - POSTGRES_PORT: PostgreSQL port (default: 5432)
//   - POSTGRES_USER: PostgreSQL user
//   - POSTGRES_PASSWORD: PostgreSQL password
//   - POSTGRES_DATABASE: PostgreSQL database name
//   - POSTGRES_SSLMODE: SSL mode (default: disable)
//
// Returns a DatabaseConfig with loaded values. Always returns a valid struct,
// never returns an error (validation should be done separately).
//
// Example:
//
//	dbConfig, err := config.LoadDatabaseConfig()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if err := dbConfig.Validate(); err != nil {
//	    log.Fatalf("Invalid database config: %v", err)
//	}
func LoadDatabaseConfig() (DatabaseConfig, error) {
	// Load MongoDB configuration
	mongoConfig := MongoDBConfig{
		URI:      getEnv("MONGODB_URI", ""),
		Database: getEnv("MONGODB_DATABASE", ""),
		Timeout:  getEnvDuration("MONGODB_TIMEOUT", 10*time.Second),
	}

	// Load PostgreSQL configuration
	postgresConfig := PostgreSQLConfig{
		Host:     getEnv("POSTGRES_HOST", ""),
		Port:     getEnvInt("POSTGRES_PORT", 5432),
		User:     getEnv("POSTGRES_USER", ""),
		Password: getEnv("POSTGRES_PASSWORD", ""),
		Database: getEnv("POSTGRES_DATABASE", ""),
		SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
	}

	return DatabaseConfig{
		MongoDB:    mongoConfig,
		PostgreSQL: postgresConfig,
	}, nil
}

// Validate validates the database configuration.
//
// It checks both MongoDB and PostgreSQL configurations for validity,
// including URI format, required fields, and value ranges.
//
// Returns nil if the configuration is valid, or an error describing
// the first validation failure encountered.
//
// Example:
//
//	if err := dbConfig.Validate(); err != nil {
//	    log.Fatalf("Invalid configuration: %v", err)
//	}
func (c *DatabaseConfig) Validate() error {
	// Validate MongoDB configuration
	if err := c.MongoDB.Validate(); err != nil {
		return fmt.Errorf("mongodb configuration: %w", err)
	}

	// Validate PostgreSQL configuration
	if err := c.PostgreSQL.Validate(); err != nil {
		return fmt.Errorf("postgresql configuration: %w", err)
	}

	return nil
}

// Validate validates MongoDB configuration.
//
// It checks:
//   - URI format (must start with "mongodb://" or "mongodb+srv://")
//   - Timeout value (must be non-negative)
//
// Note: This method allows empty URI and Database fields to support
// optional MongoDB configuration. Callers should check for empty required
// fields separately if needed.
//
// Returns nil if valid, or an error describing the validation failure.
func (m *MongoDBConfig) Validate() error {
	// Validate MongoDB URI format
	if m.URI != "" {
		if !isValidMongoDBURI(m.URI) {
			return fmt.Errorf("invalid MongoDB URI format: must start with 'mongodb://' or 'mongodb+srv://'")
		}
	}

	// Validate timeout range
	if m.Timeout < 0 {
		return fmt.Errorf("invalid timeout: must be non-negative, got %v", m.Timeout)
	}

	return nil
}

// Validate validates PostgreSQL configuration.
//
// It checks:
//   - Required fields: Host, User, Password, Database
//   - Port range: 1-65535
//   - SSLMode: must be one of [disable, require, verify-ca, verify-full]
//
// Returns nil if valid, or an error describing the first validation failure.
//
// Example:
//
//	cfg := PostgreSQLConfig{
//	    Host:     "localhost",
//	    Port:     5432,
//	    User:     "user",
//	    Password: "pass",
//	    Database: "db",
//	    SSLMode:  "disable",
//	}
//	if err := cfg.Validate(); err != nil {
//	    log.Fatal(err)
//	}
func (p *PostgreSQLConfig) Validate() error {
	// Validate required fields
	if p.Host == "" {
		return fmt.Errorf("host is required")
	}
	if p.User == "" {
		return fmt.Errorf("user is required")
	}
	if p.Password == "" {
		return fmt.Errorf("password is required")
	}
	if p.Database == "" {
		return fmt.Errorf("database is required")
	}

	// Validate port range (1-65535)
	if p.Port < 1 || p.Port > 65535 {
		return fmt.Errorf("invalid port: must be between 1 and 65535, got %d", p.Port)
	}

	// Validate SSLMode
	validSSLModes := map[string]bool{
		"disable":     true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}
	if !validSSLModes[p.SSLMode] {
		return fmt.Errorf("invalid sslmode: must be one of [disable, require, verify-ca, verify-full], got '%s'", p.SSLMode)
	}

	return nil
}

// isValidMongoDBURI checks if the URI has a valid MongoDB scheme.
//
// Valid MongoDB URIs must start with either:
//   - "mongodb://" for standard connections
//   - "mongodb+srv://" for DNS seedlist connections
//
// Returns true if the URI has a valid MongoDB scheme, false otherwise.
func isValidMongoDBURI(uri string) bool {
	return len(uri) > 10 && (uri[:10] == "mongodb://" || (len(uri) > 14 && uri[:14] == "mongodb+srv://"))
}

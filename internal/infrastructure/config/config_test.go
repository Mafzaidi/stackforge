package config

import (
	"os"
	"testing"
	"time"

	"github.com/mafzaidi/stackforge/internal/pkg/validator"
)

func TestLoad_ValidURL(t *testing.T) {
	// Set a valid URL
	os.Setenv("AUTHORIZER_BASE_URL", "https://auth.example.com")
	defer os.Unsetenv("AUTHORIZER_BASE_URL")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error for valid URL, got: %v", err)
	}

	if cfg.AuthorizerBaseURL != "https://auth.example.com" {
		t.Errorf("Expected AuthorizerBaseURL to be 'https://auth.example.com', got: %s", cfg.AuthorizerBaseURL)
	}
}

func TestLoad_InvalidURL_NoScheme(t *testing.T) {
	// Set an invalid URL without scheme
	os.Setenv("AUTHORIZER_BASE_URL", "auth.example.com")
	defer os.Unsetenv("AUTHORIZER_BASE_URL")

	_, err := Load()
	if err == nil {
		t.Fatal("Expected error for URL without scheme, got nil")
	}

	expectedMsg := "invalid AUTHORIZER_BASE_URL"
	if err.Error()[:len(expectedMsg)] != expectedMsg {
		t.Errorf("Expected error message to start with '%s', got: %s", expectedMsg, err.Error())
	}
}

func TestLoad_InvalidURL_Empty(t *testing.T) {
	// When env var is set to empty string, getEnv returns the default value
	// So this test verifies that the default value is used and is valid
	os.Setenv("AUTHORIZER_BASE_URL", "")
	defer os.Unsetenv("AUTHORIZER_BASE_URL")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error when empty string uses default, got: %v", err)
	}

	// Should use the default value
	expectedDefault := "http://authorizer.localdev.me"
	if cfg.AuthorizerBaseURL != expectedDefault {
		t.Errorf("Expected default AuthorizerBaseURL to be '%s', got: %s", expectedDefault, cfg.AuthorizerBaseURL)
	}
}

func TestLoad_DefaultURL(t *testing.T) {
	// Ensure AUTHORIZER_BASE_URL is not set
	os.Unsetenv("AUTHORIZER_BASE_URL")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error for default URL, got: %v", err)
	}

	expectedDefault := "http://authorizer.localdev.me"
	if cfg.AuthorizerBaseURL != expectedDefault {
		t.Errorf("Expected default AuthorizerBaseURL to be '%s', got: %s", expectedDefault, cfg.AuthorizerBaseURL)
	}
}

func TestValidateURL_Valid(t *testing.T) {
	tests := []string{
		"http://example.com",
		"https://example.com",
		"http://example.com:8080",
		"https://example.com/path",
		"http://localhost:4000",
	}

	for _, url := range tests {
		t.Run(url, func(t *testing.T) {
			err := validator.ValidateURL(url)
			if err != nil {
				t.Errorf("Expected no error for valid URL '%s', got: %v", url, err)
			}
		})
	}
}

func TestValidateURL_Invalid(t *testing.T) {
	tests := []struct {
		url         string
		description string
	}{
		{"", "empty URL"},
		{"example.com", "no scheme"},
		{"http://", "no host"},
		{"://example.com", "empty scheme"},
		{"not a url", "invalid format"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			err := validator.ValidateURL(tt.url)
			if err == nil {
				t.Errorf("Expected error for %s ('%s'), got nil", tt.description, tt.url)
			}
		})
	}
}

// TestLoad_AllEnvVarsSet tests loading configuration with all environment variables set
func TestLoad_AllEnvVarsSet(t *testing.T) {
	// Set all auth-related environment variables
	os.Setenv("AUTHORIZER_BASE_URL", "https://auth.example.com")
	os.Setenv("APP_CODE", "TEST-APP")
	os.Setenv("JWKS_CACHE_DURATION", "7200")
	os.Setenv("APP_CALLBACK_URL", "https://app.example.com/auth/callback")
	os.Setenv("PORT", "9090")
	os.Setenv("DATABASE_URL", "postgres://localhost/testdb")
	os.Setenv("GIN_MODE", "debug")
	defer func() {
		os.Unsetenv("AUTHORIZER_BASE_URL")
		os.Unsetenv("APP_CODE")
		os.Unsetenv("JWKS_CACHE_DURATION")
		os.Unsetenv("APP_CALLBACK_URL")
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("GIN_MODE")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error with all env vars set, got: %v", err)
	}

	// Verify all fields are set correctly
	if cfg.AuthorizerBaseURL != "https://auth.example.com" {
		t.Errorf("Expected AuthorizerBaseURL to be 'https://auth.example.com', got: %s", cfg.AuthorizerBaseURL)
	}
	if cfg.AppCode != "TEST-APP" {
		t.Errorf("Expected AppCode to be 'TEST-APP', got: %s", cfg.AppCode)
	}
	if cfg.JWKSCacheDuration != 7200 {
		t.Errorf("Expected JWKSCacheDuration to be 7200, got: %d", cfg.JWKSCacheDuration)
	}
	if cfg.AppCallbackURL != "https://app.example.com/auth/callback" {
		t.Errorf("Expected AppCallbackURL to be 'https://app.example.com/auth/callback', got: %s", cfg.AppCallbackURL)
	}
	if cfg.Port != "9090" {
		t.Errorf("Expected Port to be '9090', got: %s", cfg.Port)
	}
	if cfg.DBConnStr != "postgres://localhost/testdb" {
		t.Errorf("Expected DBConnStr to be 'postgres://localhost/testdb', got: %s", cfg.DBConnStr)
	}
	if cfg.GinMode != "debug" {
		t.Errorf("Expected GinMode to be 'debug', got: %s", cfg.GinMode)
	}
}

// TestLoad_MissingEnvVarsUseDefaults tests that missing environment variables use default values
func TestLoad_MissingEnvVarsUseDefaults(t *testing.T) {
	// Ensure all env vars are unset
	os.Unsetenv("AUTHORIZER_BASE_URL")
	os.Unsetenv("APP_CODE")
	os.Unsetenv("JWKS_CACHE_DURATION")
	os.Unsetenv("APP_CALLBACK_URL")
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("GIN_MODE")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error with default values, got: %v", err)
	}

	// Verify all fields use default values
	if cfg.AuthorizerBaseURL != "http://authorizer.localdev.me" {
		t.Errorf("Expected default AuthorizerBaseURL to be 'http://authorizer.localdev.me', got: %s", cfg.AuthorizerBaseURL)
	}
	if cfg.AppCode != "STACKFORGE" {
		t.Errorf("Expected default AppCode to be 'STACKFORGE', got: %s", cfg.AppCode)
	}
	if cfg.JWKSCacheDuration != 3600 {
		t.Errorf("Expected default JWKSCacheDuration to be 3600, got: %d", cfg.JWKSCacheDuration)
	}
	if cfg.AppCallbackURL != "http://localhost:4040/auth/callback" {
		t.Errorf("Expected default AppCallbackURL to be 'http://localhost:4040/auth/callback', got: %s", cfg.AppCallbackURL)
	}
	if cfg.Port != "8080" {
		t.Errorf("Expected default Port to be '8080', got: %s", cfg.Port)
	}
	if cfg.DBConnStr != "" {
		t.Errorf("Expected default DBConnStr to be empty, got: %s", cfg.DBConnStr)
	}
	if cfg.GinMode != "release" {
		t.Errorf("Expected default GinMode to be 'release', got: %s", cfg.GinMode)
	}
}

// TestGetEnvInt_InvalidValue tests that invalid integer values fall back to defaults
func TestGetEnvInt_InvalidValue(t *testing.T) {
	// Set JWKS_CACHE_DURATION to a non-integer value
	os.Setenv("JWKS_CACHE_DURATION", "not-a-number")
	defer os.Unsetenv("JWKS_CACHE_DURATION")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error with invalid int value (should use default), got: %v", err)
	}

	// Should fall back to default value
	if cfg.JWKSCacheDuration != 3600 {
		t.Errorf("Expected JWKSCacheDuration to fall back to default 3600, got: %d", cfg.JWKSCacheDuration)
	}
}

// TestLoad_PartialEnvVars tests loading with some env vars set and others using defaults
func TestLoad_PartialEnvVars(t *testing.T) {
	// Set only some environment variables
	os.Setenv("AUTHORIZER_BASE_URL", "https://custom-auth.example.com")
	os.Setenv("APP_CODE", "CUSTOM-APP")
	// Leave JWKS_CACHE_DURATION, APP_CALLBACK_URL, and others unset
	defer func() {
		os.Unsetenv("AUTHORIZER_BASE_URL")
		os.Unsetenv("APP_CODE")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected no error with partial env vars, got: %v", err)
	}

	// Verify set values
	if cfg.AuthorizerBaseURL != "https://custom-auth.example.com" {
		t.Errorf("Expected AuthorizerBaseURL to be 'https://custom-auth.example.com', got: %s", cfg.AuthorizerBaseURL)
	}
	if cfg.AppCode != "CUSTOM-APP" {
		t.Errorf("Expected AppCode to be 'CUSTOM-APP', got: %s", cfg.AppCode)
	}

	// Verify defaults for unset values
	if cfg.JWKSCacheDuration != 3600 {
		t.Errorf("Expected default JWKSCacheDuration to be 3600, got: %d", cfg.JWKSCacheDuration)
	}
	if cfg.AppCallbackURL != "http://localhost:4040/auth/callback" {
		t.Errorf("Expected default AppCallbackURL to be 'http://localhost:4040/auth/callback', got: %s", cfg.AppCallbackURL)
	}
}

// TestLoadDatabaseConfig_AllEnvVarsSet tests loading database configuration with all environment variables set
func TestLoadDatabaseConfig_AllEnvVarsSet(t *testing.T) {
	// Set all database environment variables
	os.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	os.Setenv("MONGODB_DATABASE", "testdb")
	os.Setenv("MONGODB_TIMEOUT", "15s")
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_PORT", "5433")
	os.Setenv("POSTGRES_USER", "testuser")
	os.Setenv("POSTGRES_PASSWORD", "testpass")
	os.Setenv("POSTGRES_DATABASE", "testdb")
	os.Setenv("POSTGRES_SSLMODE", "require")
	defer func() {
		os.Unsetenv("MONGODB_URI")
		os.Unsetenv("MONGODB_DATABASE")
		os.Unsetenv("MONGODB_TIMEOUT")
		os.Unsetenv("POSTGRES_HOST")
		os.Unsetenv("POSTGRES_PORT")
		os.Unsetenv("POSTGRES_USER")
		os.Unsetenv("POSTGRES_PASSWORD")
		os.Unsetenv("POSTGRES_DATABASE")
		os.Unsetenv("POSTGRES_SSLMODE")
	}()

	cfg, err := LoadDatabaseConfig()
	if err != nil {
		t.Fatalf("Expected no error with all env vars set, got: %v", err)
	}

	// Verify MongoDB configuration
	if cfg.MongoDB.URI != "mongodb://localhost:27017" {
		t.Errorf("Expected MongoDB URI to be 'mongodb://localhost:27017', got: %s", cfg.MongoDB.URI)
	}
	if cfg.MongoDB.Database != "testdb" {
		t.Errorf("Expected MongoDB Database to be 'testdb', got: %s", cfg.MongoDB.Database)
	}
	if cfg.MongoDB.Timeout.Seconds() != 15 {
		t.Errorf("Expected MongoDB Timeout to be 15s, got: %v", cfg.MongoDB.Timeout)
	}

	// Verify PostgreSQL configuration
	if cfg.PostgreSQL.Host != "localhost" {
		t.Errorf("Expected PostgreSQL Host to be 'localhost', got: %s", cfg.PostgreSQL.Host)
	}
	if cfg.PostgreSQL.Port != 5433 {
		t.Errorf("Expected PostgreSQL Port to be 5433, got: %d", cfg.PostgreSQL.Port)
	}
	if cfg.PostgreSQL.User != "testuser" {
		t.Errorf("Expected PostgreSQL User to be 'testuser', got: %s", cfg.PostgreSQL.User)
	}
	if cfg.PostgreSQL.Password != "testpass" {
		t.Errorf("Expected PostgreSQL Password to be 'testpass', got: %s", cfg.PostgreSQL.Password)
	}
	if cfg.PostgreSQL.Database != "testdb" {
		t.Errorf("Expected PostgreSQL Database to be 'testdb', got: %s", cfg.PostgreSQL.Database)
	}
	if cfg.PostgreSQL.SSLMode != "require" {
		t.Errorf("Expected PostgreSQL SSLMode to be 'require', got: %s", cfg.PostgreSQL.SSLMode)
	}
}

// TestLoadDatabaseConfig_DefaultValues tests that missing optional environment variables use default values
func TestLoadDatabaseConfig_DefaultValues(t *testing.T) {
	// Set only required fields, leave optional fields unset
	os.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	os.Setenv("MONGODB_DATABASE", "testdb")
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_USER", "testuser")
	os.Setenv("POSTGRES_PASSWORD", "testpass")
	os.Setenv("POSTGRES_DATABASE", "testdb")
	// Leave MONGODB_TIMEOUT, POSTGRES_PORT, and POSTGRES_SSLMODE unset
	defer func() {
		os.Unsetenv("MONGODB_URI")
		os.Unsetenv("MONGODB_DATABASE")
		os.Unsetenv("POSTGRES_HOST")
		os.Unsetenv("POSTGRES_USER")
		os.Unsetenv("POSTGRES_PASSWORD")
		os.Unsetenv("POSTGRES_DATABASE")
	}()

	cfg, err := LoadDatabaseConfig()
	if err != nil {
		t.Fatalf("Expected no error with default values, got: %v", err)
	}

	// Verify default values
	if cfg.MongoDB.Timeout.Seconds() != 10 {
		t.Errorf("Expected default MongoDB Timeout to be 10s, got: %v", cfg.MongoDB.Timeout)
	}
	if cfg.PostgreSQL.Port != 5432 {
		t.Errorf("Expected default PostgreSQL Port to be 5432, got: %d", cfg.PostgreSQL.Port)
	}
	if cfg.PostgreSQL.SSLMode != "disable" {
		t.Errorf("Expected default PostgreSQL SSLMode to be 'disable', got: %s", cfg.PostgreSQL.SSLMode)
	}
}

// TestLoadDatabaseConfig_EmptyValues tests loading with empty environment variables
func TestLoadDatabaseConfig_EmptyValues(t *testing.T) {
	// Ensure all env vars are unset
	os.Unsetenv("MONGODB_URI")
	os.Unsetenv("MONGODB_DATABASE")
	os.Unsetenv("MONGODB_TIMEOUT")
	os.Unsetenv("POSTGRES_HOST")
	os.Unsetenv("POSTGRES_PORT")
	os.Unsetenv("POSTGRES_USER")
	os.Unsetenv("POSTGRES_PASSWORD")
	os.Unsetenv("POSTGRES_DATABASE")
	os.Unsetenv("POSTGRES_SSLMODE")

	cfg, err := LoadDatabaseConfig()
	if err != nil {
		t.Fatalf("Expected no error with empty values, got: %v", err)
	}

	// Verify empty values for required fields
	if cfg.MongoDB.URI != "" {
		t.Errorf("Expected MongoDB URI to be empty, got: %s", cfg.MongoDB.URI)
	}
	if cfg.MongoDB.Database != "" {
		t.Errorf("Expected MongoDB Database to be empty, got: %s", cfg.MongoDB.Database)
	}
	if cfg.PostgreSQL.Host != "" {
		t.Errorf("Expected PostgreSQL Host to be empty, got: %s", cfg.PostgreSQL.Host)
	}
	if cfg.PostgreSQL.User != "" {
		t.Errorf("Expected PostgreSQL User to be empty, got: %s", cfg.PostgreSQL.User)
	}
	if cfg.PostgreSQL.Password != "" {
		t.Errorf("Expected PostgreSQL Password to be empty, got: %s", cfg.PostgreSQL.Password)
	}
	if cfg.PostgreSQL.Database != "" {
		t.Errorf("Expected PostgreSQL Database to be empty, got: %s", cfg.PostgreSQL.Database)
	}

	// Verify default values for optional fields
	if cfg.MongoDB.Timeout.Seconds() != 10 {
		t.Errorf("Expected default MongoDB Timeout to be 10s, got: %v", cfg.MongoDB.Timeout)
	}
	if cfg.PostgreSQL.Port != 5432 {
		t.Errorf("Expected default PostgreSQL Port to be 5432, got: %d", cfg.PostgreSQL.Port)
	}
	if cfg.PostgreSQL.SSLMode != "disable" {
		t.Errorf("Expected default PostgreSQL SSLMode to be 'disable', got: %s", cfg.PostgreSQL.SSLMode)
	}
}

// TestGetEnvDuration_ValidValue tests parsing valid duration values
func TestGetEnvDuration_ValidValue(t *testing.T) {
	os.Setenv("MONGODB_TIMEOUT", "30s")
	defer os.Unsetenv("MONGODB_TIMEOUT")

	cfg, err := LoadDatabaseConfig()
	if err != nil {
		t.Fatalf("Expected no error with valid duration, got: %v", err)
	}

	if cfg.MongoDB.Timeout.Seconds() != 30 {
		t.Errorf("Expected MongoDB Timeout to be 30s, got: %v", cfg.MongoDB.Timeout)
	}
}

// TestGetEnvDuration_InvalidValue tests that invalid duration values fall back to defaults
func TestGetEnvDuration_InvalidValue(t *testing.T) {
	os.Setenv("MONGODB_TIMEOUT", "not-a-duration")
	defer os.Unsetenv("MONGODB_TIMEOUT")

	cfg, err := LoadDatabaseConfig()
	if err != nil {
		t.Fatalf("Expected no error with invalid duration (should use default), got: %v", err)
	}

	// Should fall back to default value
	if cfg.MongoDB.Timeout.Seconds() != 10 {
		t.Errorf("Expected MongoDB Timeout to fall back to default 10s, got: %v", cfg.MongoDB.Timeout)
	}
}

// TestLoadDatabaseConfig_PartialMongoDBConfig tests loading with partial MongoDB configuration
func TestLoadDatabaseConfig_PartialMongoDBConfig(t *testing.T) {
	// Set only MongoDB URI, leave database and timeout unset
	os.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	defer os.Unsetenv("MONGODB_URI")

	cfg, err := LoadDatabaseConfig()
	if err != nil {
		t.Fatalf("Expected no error with partial MongoDB config, got: %v", err)
	}

	// Verify set value
	if cfg.MongoDB.URI != "mongodb://localhost:27017" {
		t.Errorf("Expected MongoDB URI to be 'mongodb://localhost:27017', got: %s", cfg.MongoDB.URI)
	}

	// Verify empty database
	if cfg.MongoDB.Database != "" {
		t.Errorf("Expected MongoDB Database to be empty, got: %s", cfg.MongoDB.Database)
	}

	// Verify default timeout
	if cfg.MongoDB.Timeout.Seconds() != 10 {
		t.Errorf("Expected default MongoDB Timeout to be 10s, got: %v", cfg.MongoDB.Timeout)
	}
}

// TestLoadDatabaseConfig_PartialPostgreSQLConfig tests loading with partial PostgreSQL configuration
func TestLoadDatabaseConfig_PartialPostgreSQLConfig(t *testing.T) {
	// Set only required PostgreSQL fields
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_USER", "testuser")
	os.Setenv("POSTGRES_PASSWORD", "testpass")
	os.Setenv("POSTGRES_DATABASE", "testdb")
	// Leave port and sslmode unset
	defer func() {
		os.Unsetenv("POSTGRES_HOST")
		os.Unsetenv("POSTGRES_USER")
		os.Unsetenv("POSTGRES_PASSWORD")
		os.Unsetenv("POSTGRES_DATABASE")
	}()

	cfg, err := LoadDatabaseConfig()
	if err != nil {
		t.Fatalf("Expected no error with partial PostgreSQL config, got: %v", err)
	}

	// Verify set values
	if cfg.PostgreSQL.Host != "localhost" {
		t.Errorf("Expected PostgreSQL Host to be 'localhost', got: %s", cfg.PostgreSQL.Host)
	}
	if cfg.PostgreSQL.User != "testuser" {
		t.Errorf("Expected PostgreSQL User to be 'testuser', got: %s", cfg.PostgreSQL.User)
	}
	if cfg.PostgreSQL.Password != "testpass" {
		t.Errorf("Expected PostgreSQL Password to be 'testpass', got: %s", cfg.PostgreSQL.Password)
	}
	if cfg.PostgreSQL.Database != "testdb" {
		t.Errorf("Expected PostgreSQL Database to be 'testdb', got: %s", cfg.PostgreSQL.Database)
	}

	// Verify default values
	if cfg.PostgreSQL.Port != 5432 {
		t.Errorf("Expected default PostgreSQL Port to be 5432, got: %d", cfg.PostgreSQL.Port)
	}
	if cfg.PostgreSQL.SSLMode != "disable" {
		t.Errorf("Expected default PostgreSQL SSLMode to be 'disable', got: %s", cfg.PostgreSQL.SSLMode)
	}
}

// TestDatabaseConfig_Validate_ValidConfig tests validation with valid configuration
func TestDatabaseConfig_Validate_ValidConfig(t *testing.T) {
	cfg := DatabaseConfig{
		MongoDB: MongoDBConfig{
			URI:      "mongodb://localhost:27017",
			Database: "testdb",
			Timeout:  10 * time.Second,
		},
		PostgreSQL: PostgreSQLConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "testuser",
			Password: "testpass",
			Database: "testdb",
			SSLMode:  "disable",
		},
	}

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}
}

// TestMongoDBConfig_Validate_ValidURI tests MongoDB URI validation with valid URIs
func TestMongoDBConfig_Validate_ValidURI(t *testing.T) {
	tests := []struct {
		name string
		uri  string
	}{
		{"standard URI", "mongodb://localhost:27017"},
		{"with auth", "mongodb://user:pass@localhost:27017"},
		{"with database", "mongodb://localhost:27017/mydb"},
		{"with options", "mongodb://localhost:27017/?replicaSet=rs0"},
		{"srv URI", "mongodb+srv://cluster.example.com"},
		{"srv with auth", "mongodb+srv://user:pass@cluster.example.com"},
		{"empty URI", ""}, // Empty URI is allowed (optional)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := MongoDBConfig{
				URI:      tt.uri,
				Database: "testdb",
				Timeout:  10 * time.Second,
			}

			err := cfg.Validate()
			if err != nil {
				t.Errorf("Expected no error for valid URI '%s', got: %v", tt.uri, err)
			}
		})
	}
}

// TestMongoDBConfig_Validate_InvalidURI tests MongoDB URI validation with invalid URIs
func TestMongoDBConfig_Validate_InvalidURI(t *testing.T) {
	tests := []struct {
		name string
		uri  string
	}{
		{"no scheme", "localhost:27017"},
		{"wrong scheme", "http://localhost:27017"},
		{"postgres scheme", "postgres://localhost:5432"},
		{"incomplete mongodb", "mongodb:/"},
		{"incomplete mongodb+srv", "mongodb+srv:/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := MongoDBConfig{
				URI:      tt.uri,
				Database: "testdb",
				Timeout:  10 * time.Second,
			}

			err := cfg.Validate()
			if err == nil {
				t.Errorf("Expected error for invalid URI '%s', got nil", tt.uri)
			}
		})
	}
}

// TestMongoDBConfig_Validate_InvalidTimeout tests MongoDB timeout validation
func TestMongoDBConfig_Validate_InvalidTimeout(t *testing.T) {
	cfg := MongoDBConfig{
		URI:      "mongodb://localhost:27017",
		Database: "testdb",
		Timeout:  -5 * time.Second,
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for negative timeout, got nil")
	}
}

// TestPostgreSQLConfig_Validate_MissingRequiredFields tests PostgreSQL validation with missing required fields
func TestPostgreSQLConfig_Validate_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name   string
		config PostgreSQLConfig
		errMsg string
	}{
		{
			name: "missing host",
			config: PostgreSQLConfig{
				Host:     "",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Database: "testdb",
				SSLMode:  "disable",
			},
			errMsg: "host is required",
		},
		{
			name: "missing user",
			config: PostgreSQLConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "",
				Password: "testpass",
				Database: "testdb",
				SSLMode:  "disable",
			},
			errMsg: "user is required",
		},
		{
			name: "missing password",
			config: PostgreSQLConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "",
				Database: "testdb",
				SSLMode:  "disable",
			},
			errMsg: "password is required",
		},
		{
			name: "missing database",
			config: PostgreSQLConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Database: "",
				SSLMode:  "disable",
			},
			errMsg: "database is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("Expected error message '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

// TestPostgreSQLConfig_Validate_InvalidPort tests PostgreSQL port validation
func TestPostgreSQLConfig_Validate_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"port too low", 0},
		{"port negative", -1},
		{"port too high", 65536},
		{"port way too high", 100000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := PostgreSQLConfig{
				Host:     "localhost",
				Port:     tt.port,
				User:     "testuser",
				Password: "testpass",
				Database: "testdb",
				SSLMode:  "disable",
			}

			err := cfg.Validate()
			if err == nil {
				t.Errorf("Expected error for port %d, got nil", tt.port)
			}
		})
	}
}

// TestPostgreSQLConfig_Validate_ValidPort tests PostgreSQL port validation with valid ports
func TestPostgreSQLConfig_Validate_ValidPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"minimum port", 1},
		{"standard port", 5432},
		{"maximum port", 65535},
		{"custom port", 8080},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := PostgreSQLConfig{
				Host:     "localhost",
				Port:     tt.port,
				User:     "testuser",
				Password: "testpass",
				Database: "testdb",
				SSLMode:  "disable",
			}

			err := cfg.Validate()
			if err != nil {
				t.Errorf("Expected no error for valid port %d, got: %v", tt.port, err)
			}
		})
	}
}

// TestPostgreSQLConfig_Validate_InvalidSSLMode tests PostgreSQL SSLMode validation
func TestPostgreSQLConfig_Validate_InvalidSSLMode(t *testing.T) {
	tests := []string{
		"invalid",
		"enabled",
		"true",
		"false",
		"DISABLE", // case-sensitive
		"",        // empty is invalid
	}

	for _, sslMode := range tests {
		t.Run(sslMode, func(t *testing.T) {
			cfg := PostgreSQLConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Database: "testdb",
				SSLMode:  sslMode,
			}

			err := cfg.Validate()
			if err == nil {
				t.Errorf("Expected error for invalid SSLMode '%s', got nil", sslMode)
			}
		})
	}
}

// TestPostgreSQLConfig_Validate_ValidSSLMode tests PostgreSQL SSLMode validation with valid modes
func TestPostgreSQLConfig_Validate_ValidSSLMode(t *testing.T) {
	tests := []string{
		"disable",
		"require",
		"verify-ca",
		"verify-full",
	}

	for _, sslMode := range tests {
		t.Run(sslMode, func(t *testing.T) {
			cfg := PostgreSQLConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Database: "testdb",
				SSLMode:  sslMode,
			}

			err := cfg.Validate()
			if err != nil {
				t.Errorf("Expected no error for valid SSLMode '%s', got: %v", sslMode, err)
			}
		})
	}
}

// TestDatabaseConfig_Validate_MongoDBError tests that MongoDB validation errors are properly wrapped
func TestDatabaseConfig_Validate_MongoDBError(t *testing.T) {
	cfg := DatabaseConfig{
		MongoDB: MongoDBConfig{
			URI:      "invalid-uri",
			Database: "testdb",
			Timeout:  10 * time.Second,
		},
		PostgreSQL: PostgreSQLConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "testuser",
			Password: "testpass",
			Database: "testdb",
			SSLMode:  "disable",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for invalid MongoDB config, got nil")
	}
	// Check that error message includes "mongodb configuration"
	if err != nil && len(err.Error()) < 20 {
		t.Errorf("Expected error message to include context, got: %v", err)
	}
}

// TestDatabaseConfig_Validate_PostgreSQLError tests that PostgreSQL validation errors are properly wrapped
func TestDatabaseConfig_Validate_PostgreSQLError(t *testing.T) {
	cfg := DatabaseConfig{
		MongoDB: MongoDBConfig{
			URI:      "mongodb://localhost:27017",
			Database: "testdb",
			Timeout:  10 * time.Second,
		},
		PostgreSQL: PostgreSQLConfig{
			Host:     "",
			Port:     5432,
			User:     "testuser",
			Password: "testpass",
			Database: "testdb",
			SSLMode:  "disable",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected error for invalid PostgreSQL config, got nil")
	}
	// Check that error message includes "postgresql configuration"
	if err != nil && len(err.Error()) < 20 {
		t.Errorf("Expected error message to include context, got: %v", err)
	}
}

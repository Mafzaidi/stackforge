// Package database provides database connection management for MongoDB and PostgreSQL.
//
// This file defines error types used throughout the database package for
// configuration, connection, health check, and shutdown errors.
package database

import "fmt"

// ConfigError represents configuration-related errors.
//
// It provides detailed information about which configuration field caused
// the error and a descriptive message explaining the problem.
//
// Example:
//
//	err := &ConfigError{
//	    Field:   "MONGODB_URI",
//	    Message: "invalid URI format",
//	}
//	fmt.Println(err) // Output: configuration error in field 'MONGODB_URI': invalid URI format
type ConfigError struct {
	Field   string // Field that caused the error
	Message string // Error description
}

// Error implements the error interface for ConfigError.
//
// It formats the error message to include both the field name and the
// descriptive error message.
func (e *ConfigError) Error() string {
	return fmt.Sprintf("configuration error in field '%s': %s", e.Field, e.Message)
}

// ConnectionError represents connection-related errors.
//
// It wraps the underlying connection error and provides context about which
// database failed and how many retry attempts were made.
//
// Example:
//
//	err := &ConnectionError{
//	    Database: "mongodb",
//	    Cause:    originalErr,
//	    Retries:  3,
//	}
//	fmt.Println(err) // Output: failed to connect to mongodb after 3 retries: <original error>
type ConnectionError struct {
	Database string // Database type (mongodb/postgresql)
	Cause    error  // Underlying error
	Retries  int    // Number of retries attempted
}

// Error implements the error interface for ConnectionError.
//
// It formats the error message to include the database type, number of retries
// (if any), and the underlying cause.
func (e *ConnectionError) Error() string {
	if e.Retries > 0 {
		return fmt.Sprintf("failed to connect to %s after %d retries: %v", e.Database, e.Retries, e.Cause)
	}
	return fmt.Sprintf("failed to connect to %s: %v", e.Database, e.Cause)
}

// Unwrap returns the underlying cause error for ConnectionError.
//
// This allows ConnectionError to work with errors.Is and errors.As for
// error chain inspection.
func (e *ConnectionError) Unwrap() error {
	return e.Cause
}

// HealthCheckError represents health check errors.
//
// It wraps the underlying error that occurred during a health check and
// provides context about which database was being checked.
//
// Example:
//
//	err := &HealthCheckError{
//	    Database: "postgresql",
//	    Cause:    originalErr,
//	}
//	fmt.Println(err) // Output: health check failed for postgresql: <original error>
type HealthCheckError struct {
	Database string // Database type
	Cause    error  // Underlying error
}

// Error implements the error interface for HealthCheckError.
//
// It formats the error message to include the database type and the
// underlying cause.
func (e *HealthCheckError) Error() string {
	return fmt.Sprintf("health check failed for %s: %v", e.Database, e.Cause)
}

// Unwrap returns the underlying cause error for HealthCheckError.
//
// This allows HealthCheckError to work with errors.Is and errors.As for
// error chain inspection.
func (e *HealthCheckError) Unwrap() error {
	return e.Cause
}

// ShutdownError represents shutdown errors.
//
// It wraps the underlying error that occurred during database shutdown and
// provides context about which database failed to close properly.
//
// Example:
//
//	err := &ShutdownError{
//	    Database: "mongodb",
//	    Cause:    originalErr,
//	}
//	fmt.Println(err) // Output: shutdown failed for mongodb: <original error>
type ShutdownError struct {
	Database string // Database type
	Cause    error  // Underlying error
}

// Error implements the error interface for ShutdownError.
//
// It formats the error message to include the database type and the
// underlying cause.
func (e *ShutdownError) Error() string {
	return fmt.Sprintf("shutdown failed for %s: %v", e.Database, e.Cause)
}

// Unwrap returns the underlying cause error for ShutdownError.
//
// This allows ShutdownError to work with errors.Is and errors.As for
// error chain inspection.
func (e *ShutdownError) Unwrap() error {
	return e.Cause
}

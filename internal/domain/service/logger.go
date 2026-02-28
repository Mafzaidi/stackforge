package service

// Fields represents structured log fields
type Fields map[string]interface{}

// Logger defines the interface for application logging
// This interface allows the delivery layer to depend on logging
// without coupling to the infrastructure implementation
type Logger interface {
	Info(message string, fields Fields)
	Warn(message string, fields Fields)
	Error(message string, fields Fields)
}

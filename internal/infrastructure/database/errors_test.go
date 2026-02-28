package database

import (
	"errors"
	"testing"
)

func TestConfigError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ConfigError
		expected string
	}{
		{
			name: "config error with field and message",
			err: &ConfigError{
				Field:   "MONGODB_URI",
				Message: "missing required field",
			},
			expected: "configuration error in field 'MONGODB_URI': missing required field",
		},
		{
			name: "config error with empty field",
			err: &ConfigError{
				Field:   "",
				Message: "invalid configuration",
			},
			expected: "configuration error in field '': invalid configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("ConfigError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConnectionError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ConnectionError
		expected string
	}{
		{
			name: "connection error without retries",
			err: &ConnectionError{
				Database: "mongodb",
				Cause:    errors.New("connection refused"),
				Retries:  0,
			},
			expected: "failed to connect to mongodb: connection refused",
		},
		{
			name: "connection error with retries",
			err: &ConnectionError{
				Database: "postgresql",
				Cause:    errors.New("timeout"),
				Retries:  3,
			},
			expected: "failed to connect to postgresql after 3 retries: timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("ConnectionError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConnectionError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &ConnectionError{
		Database: "mongodb",
		Cause:    cause,
		Retries:  0,
	}

	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Errorf("ConnectionError.Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestHealthCheckError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *HealthCheckError
		expected string
	}{
		{
			name: "health check error for mongodb",
			err: &HealthCheckError{
				Database: "mongodb",
				Cause:    errors.New("ping timeout"),
			},
			expected: "health check failed for mongodb: ping timeout",
		},
		{
			name: "health check error for postgresql",
			err: &HealthCheckError{
				Database: "postgresql",
				Cause:    errors.New("connection lost"),
			},
			expected: "health check failed for postgresql: connection lost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("HealthCheckError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestHealthCheckError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &HealthCheckError{
		Database: "mongodb",
		Cause:    cause,
	}

	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Errorf("HealthCheckError.Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestShutdownError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ShutdownError
		expected string
	}{
		{
			name: "shutdown error for mongodb",
			err: &ShutdownError{
				Database: "mongodb",
				Cause:    errors.New("disconnect timeout"),
			},
			expected: "shutdown failed for mongodb: disconnect timeout",
		},
		{
			name: "shutdown error for postgresql",
			err: &ShutdownError{
				Database: "postgresql",
				Cause:    errors.New("close failed"),
			},
			expected: "shutdown failed for postgresql: close failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("ShutdownError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestShutdownError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &ShutdownError{
		Database: "postgresql",
		Cause:    cause,
	}

	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Errorf("ShutdownError.Unwrap() = %v, want %v", unwrapped, cause)
	}
}

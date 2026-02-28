package validator

import (
	"strings"
	"testing"
)

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"tabs and spaces", "\t  \n", true},
		{"non-empty string", "hello", false},
		{"string with spaces", "hello world", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmpty(tt.input)
			if result != tt.expected {
				t.Errorf("IsEmpty(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "valid http URL",
			input:     "http://example.com",
			shouldErr: false,
		},
		{
			name:      "valid https URL",
			input:     "https://example.com/path",
			shouldErr: false,
		},
		{
			name:      "empty URL",
			input:     "",
			shouldErr: true,
			errMsg:    "URL cannot be empty",
		},
		{
			name:      "URL without scheme",
			input:     "example.com",
			shouldErr: true,
			errMsg:    "URL must include a scheme",
		},
		{
			name:      "URL without host",
			input:     "http://",
			shouldErr: true,
			errMsg:    "URL must include a host",
		},
		{
			name:      "invalid URL format",
			input:     "ht!tp://invalid",
			shouldErr: true,
			errMsg:    "failed to parse URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.input)
			if tt.shouldErr {
				if err == nil {
					t.Errorf("ValidateURL(%q) expected error, got nil", tt.input)
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateURL(%q) error = %v, want error containing %q", tt.input, err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateURL(%q) unexpected error: %v", tt.input, err)
				}
			}
		})
	}
}

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     string
		shouldErr bool
	}{
		{"valid value", "username", "john", false},
		{"empty value", "username", "", true},
		{"whitespace only", "email", "   ", true},
		{"valid with spaces", "name", "John Doe", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequired(tt.fieldName, tt.value)
			if tt.shouldErr && err == nil {
				t.Errorf("ValidateRequired(%q, %q) expected error, got nil", tt.fieldName, tt.value)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("ValidateRequired(%q, %q) unexpected error: %v", tt.fieldName, tt.value, err)
			}
			if tt.shouldErr && err != nil && !strings.Contains(err.Error(), tt.fieldName) {
				t.Errorf("ValidateRequired(%q, %q) error should contain field name, got: %v", tt.fieldName, tt.value, err)
			}
		})
	}
}

func TestValidateMinLength(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     string
		minLength int
		shouldErr bool
	}{
		{"meets minimum", "password", "12345", 5, false},
		{"exceeds minimum", "password", "123456", 5, false},
		{"below minimum", "password", "1234", 5, true},
		{"empty string", "password", "", 5, true},
		{"whitespace trimmed", "password", "  12345  ", 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMinLength(tt.fieldName, tt.value, tt.minLength)
			if tt.shouldErr && err == nil {
				t.Errorf("ValidateMinLength(%q, %q, %d) expected error, got nil", tt.fieldName, tt.value, tt.minLength)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("ValidateMinLength(%q, %q, %d) unexpected error: %v", tt.fieldName, tt.value, tt.minLength, err)
			}
		})
	}
}

func TestValidateMaxLength(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     string
		maxLength int
		shouldErr bool
	}{
		{"within maximum", "username", "john", 10, false},
		{"at maximum", "username", "1234567890", 10, false},
		{"exceeds maximum", "username", "12345678901", 10, true},
		{"empty string", "username", "", 10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMaxLength(tt.fieldName, tt.value, tt.maxLength)
			if tt.shouldErr && err == nil {
				t.Errorf("ValidateMaxLength(%q, %q, %d) expected error, got nil", tt.fieldName, tt.value, tt.maxLength)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("ValidateMaxLength(%q, %q, %d) unexpected error: %v", tt.fieldName, tt.value, tt.maxLength, err)
			}
		})
	}
}

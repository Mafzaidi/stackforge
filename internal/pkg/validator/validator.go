package validator

import (
	"fmt"
	"net/url"
	"strings"
)

// IsEmpty checks if a string is empty or contains only whitespace
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// ValidateURL checks if the provided string is a valid URL format
// Returns an error if the URL is invalid
func ValidateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	// Ensure the URL has a scheme (http or https)
	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL must include a scheme (http or https)")
	}

	// Ensure the URL has a host
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must include a host")
	}

	return nil
}

// ValidateRequired checks if a required field is present
// Returns an error with the field name if validation fails
func ValidateRequired(fieldName, value string) error {
	if IsEmpty(value) {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidateMinLength checks if a string meets minimum length requirement
func ValidateMinLength(fieldName, value string, minLength int) error {
	if len(strings.TrimSpace(value)) < minLength {
		return fmt.Errorf("%s must be at least %d characters", fieldName, minLength)
	}
	return nil
}

// ValidateMaxLength checks if a string doesn't exceed maximum length
func ValidateMaxLength(fieldName, value string, maxLength int) error {
	if len(value) > maxLength {
		return fmt.Errorf("%s must not exceed %d characters", fieldName, maxLength)
	}
	return nil
}

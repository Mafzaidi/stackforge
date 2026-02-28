package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/mafzaidi/stackforge/internal/pkg/validator"
)

// **Validates: Requirements 8.5**
// Property 10: Configuration URL Validation
// For any string provided as AUTHORIZER_BASE_URL, if it's not a valid URL format,
// the configuration loading should return an error; if it's valid, it should be accepted.
func TestProperty_ConfigurationURLValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Valid URLs are accepted", prop.ForAll(
		func(scheme string, host string, port int, path string) bool {
			// Build a valid URL
			var urlStr string
			if port > 0 && port < 65536 {
				urlStr = fmt.Sprintf("%s://%s:%d%s", scheme, host, port, path)
			} else {
				urlStr = fmt.Sprintf("%s://%s%s", scheme, host, path)
			}

			// Set environment variable
			os.Setenv("AUTHORIZER_BASE_URL", urlStr)
			defer os.Unsetenv("AUTHORIZER_BASE_URL")

			// Load configuration
			cfg, err := Load()

			// Valid URLs should not return an error
			if err != nil {
				t.Logf("Expected valid URL %s to be accepted, got error: %v", urlStr, err)
				return false
			}

			// Verify the URL was loaded correctly
			if cfg.AuthorizerBaseURL != urlStr {
				t.Logf("Expected AuthorizerBaseURL=%s, got %s", urlStr, cfg.AuthorizerBaseURL)
				return false
			}

			return true
		},
		gen.OneConstOf("http", "https"),
		genValidHostname(),
		gen.IntRange(0, 65535),
		genValidPath(),
	))

	properties.Property("URLs without scheme are rejected", prop.ForAll(
		func(host string) bool {
			// Build URL without scheme
			urlStr := host

			// Set environment variable
			os.Setenv("AUTHORIZER_BASE_URL", urlStr)
			defer os.Unsetenv("AUTHORIZER_BASE_URL")

			// Load configuration
			_, err := Load()

			// URLs without scheme should return an error
			if err == nil {
				t.Logf("Expected URL without scheme %s to be rejected, but it was accepted", urlStr)
				return false
			}

			return true
		},
		genValidHostname(),
	))

	properties.Property("URLs without host are rejected", prop.ForAll(
		func(scheme string, path string) bool {
			// Build URL without host (just scheme and path)
			urlStr := fmt.Sprintf("%s://%s", scheme, path)

			// Set environment variable
			os.Setenv("AUTHORIZER_BASE_URL", urlStr)
			defer os.Unsetenv("AUTHORIZER_BASE_URL")

			// Load configuration
			_, err := Load()

			// URLs without host should return an error
			if err == nil {
				t.Logf("Expected URL without host %s to be rejected, but it was accepted", urlStr)
				return false
			}

			return true
		},
		gen.OneConstOf("http", "https"),
		genValidPath(),
	))

	properties.Property("Empty env var uses default URL", prop.ForAll(
		func() bool {
			// Set empty URL (which will trigger default value usage)
			os.Setenv("AUTHORIZER_BASE_URL", "")
			defer os.Unsetenv("AUTHORIZER_BASE_URL")

			// Load configuration
			cfg, err := Load()

			// Should not return an error (uses default)
			if err != nil {
				t.Logf("Expected default URL to be used when env var is empty, got error: %v", err)
				return false
			}

			// Verify default URL is set
			expectedDefault := "http://authorizer.localdev.me"
			if cfg.AuthorizerBaseURL != expectedDefault {
				t.Logf("Expected default URL %s, got %s", expectedDefault, cfg.AuthorizerBaseURL)
				return false
			}

			return true
		},
	))

	properties.Property("Malformed URLs are rejected", prop.ForAll(
		func(malformedURL string) bool {
			// Skip if accidentally generated a valid URL
			if err := validator.ValidateURL(malformedURL); err == nil {
				return true // Skip this test case
			}

			// Set environment variable
			os.Setenv("AUTHORIZER_BASE_URL", malformedURL)
			defer os.Unsetenv("AUTHORIZER_BASE_URL")

			// Load configuration
			_, err := Load()

			// Malformed URLs should return an error
			if err == nil {
				t.Logf("Expected malformed URL %s to be rejected, but it was accepted", malformedURL)
				return false
			}

			return true
		},
		genMalformedURL(),
	))

	properties.Property("URLs with special characters in host are handled correctly", prop.ForAll(
		func(scheme string, hostWithSpecialChars string) bool {
			// Build URL with special characters
			urlStr := fmt.Sprintf("%s://%s", scheme, hostWithSpecialChars)

			// Set environment variable
			os.Setenv("AUTHORIZER_BASE_URL", urlStr)
			defer os.Unsetenv("AUTHORIZER_BASE_URL")

			// Load configuration
			_, err := Load()

			// Check if URL parsing succeeds or fails
			// The behavior depends on whether the special characters are valid in a hostname
			validationErr := validator.ValidateURL(urlStr)
			
			if validationErr != nil && err == nil {
				t.Logf("Expected URL %s to be rejected, but it was accepted", urlStr)
				return false
			}

			if validationErr == nil && err != nil {
				t.Logf("Expected URL %s to be accepted, but got error: %v", urlStr, err)
				return false
			}

			return true
		},
		gen.OneConstOf("http", "https"),
		genHostnameWithSpecialChars(),
	))

	properties.Property("Default URL is used when env var is not set", prop.ForAll(
		func() bool {
			// Ensure env var is not set
			os.Unsetenv("AUTHORIZER_BASE_URL")

			// Load configuration
			cfg, err := Load()

			// Should not return an error
			if err != nil {
				t.Logf("Expected default URL to be used, got error: %v", err)
				return false
			}

			// Verify default URL is set
			expectedDefault := "http://authorizer.localdev.me"
			if cfg.AuthorizerBaseURL != expectedDefault {
				t.Logf("Expected default URL %s, got %s", expectedDefault, cfg.AuthorizerBaseURL)
				return false
			}

			return true
		},
	))

	properties.TestingRun(t)
}

// Generator functions for property testing

// genValidHostname generates valid hostnames
func genValidHostname() gopter.Gen {
	return gen.OneConstOf(
		"localhost",
		"authorizer.test",
		"auth.example.com",
		"sso.company.com",
		"login.service.local",
		"api.domain.org",
		"192.168.1.1",
		"10.0.0.1",
	)
}

// genValidPath generates valid URL paths
func genValidPath() gopter.Gen {
	return gen.OneConstOf(
		"",
		"/",
		"/api",
		"/auth",
		"/api/v1",
		"/auth/login",
	)
}

// genMalformedURL generates malformed URLs
func genMalformedURL() gopter.Gen {
	return gen.OneConstOf(
		"ht!tp://invalid",
		"://noscheme.com",
		"http://",
		"http:///nohost",
		"not a url at all",
		"ftp://unsupported.com", // Valid URL but might be considered invalid by app logic
		"http://host with spaces.com",
		"http://[invalid",
		"http://]invalid",
		"http://host:99999", // Invalid port
		"http://host:-1",    // Invalid port
	)
}

// genHostnameWithSpecialChars generates hostnames with special characters
func genHostnameWithSpecialChars() gopter.Gen {
	return gen.OneConstOf(
		"host-with-dash.com",
		"host_with_underscore.com",
		"host.with.dots.com",
		"192.168.1.1",
		"[::1]", // IPv6
		"host:8080",
		"host@invalid.com",
		"host#invalid.com",
		"host$invalid.com",
	)
}

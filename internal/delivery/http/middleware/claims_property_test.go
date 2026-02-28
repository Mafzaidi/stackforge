package middleware

import (
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

// **Validates: Requirements 6.1, 6.2, 6.3**
// Property 6: Claims Context Round-Trip
// For any valid claims structure, if we store it in the User_Context and then
// retrieve it, we should get back an equivalent claims structure with all
// fields intact (subject, username, email, authorization).
func TestProperty_ClaimsContextRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Claims context round-trip preserves all fields", prop.ForAll(
		func(issuer string, subject string, audience []string, exp int64, iat int64,
			username string, email string, authorizations []testAuthorization) bool {

			// Create a gin context for testing
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Build the claims structure
			originalClaims := &entity.Claims{
				Issuer:    issuer,
				Subject:   subject,
				Audience:  audience,
				ExpiresAt: exp,
				IssuedAt:  iat,
				Username:  username,
				Email:     email,
			}

			// Convert test authorizations to domain authorizations
			for _, auth := range authorizations {
				originalClaims.Authorization = append(originalClaims.Authorization, entity.Authorization{
					App:         auth.App,
					Roles:       auth.Roles,
					Permissions: auth.Permissions,
				})
			}

			// Store claims in context
			SetClaims(c, originalClaims)

			// Retrieve claims from context
			retrievedClaims, err := GetClaims(c)
			if err != nil {
				t.Logf("Failed to retrieve claims: %v", err)
				return false
			}

			// Verify all fields match
			if retrievedClaims.Issuer != originalClaims.Issuer {
				t.Logf("Issuer mismatch: expected '%s', got '%s'", originalClaims.Issuer, retrievedClaims.Issuer)
				return false
			}

			if retrievedClaims.Subject != originalClaims.Subject {
				t.Logf("Subject mismatch: expected '%s', got '%s'", originalClaims.Subject, retrievedClaims.Subject)
				return false
			}

			if !stringSlicesEqual(retrievedClaims.Audience, originalClaims.Audience) {
				t.Logf("Audience mismatch: expected %v, got %v", originalClaims.Audience, retrievedClaims.Audience)
				return false
			}

			if retrievedClaims.ExpiresAt != originalClaims.ExpiresAt {
				t.Logf("ExpiresAt mismatch: expected %d, got %d", originalClaims.ExpiresAt, retrievedClaims.ExpiresAt)
				return false
			}

			if retrievedClaims.IssuedAt != originalClaims.IssuedAt {
				t.Logf("IssuedAt mismatch: expected %d, got %d", originalClaims.IssuedAt, retrievedClaims.IssuedAt)
				return false
			}

			if retrievedClaims.Username != originalClaims.Username {
				t.Logf("Username mismatch: expected '%s', got '%s'", originalClaims.Username, retrievedClaims.Username)
				return false
			}

			if retrievedClaims.Email != originalClaims.Email {
				t.Logf("Email mismatch: expected '%s', got '%s'", originalClaims.Email, retrievedClaims.Email)
				return false
			}

			// Verify authorization array
			if len(retrievedClaims.Authorization) != len(originalClaims.Authorization) {
				t.Logf("Authorization length mismatch: expected %d, got %d",
					len(originalClaims.Authorization), len(retrievedClaims.Authorization))
				return false
			}

			for i, originalAuth := range originalClaims.Authorization {
				retrievedAuth := retrievedClaims.Authorization[i]

				if retrievedAuth.App != originalAuth.App {
					t.Logf("Authorization[%d].App mismatch: expected '%s', got '%s'",
						i, originalAuth.App, retrievedAuth.App)
					return false
				}

				if !stringSlicesEqual(retrievedAuth.Roles, originalAuth.Roles) {
					t.Logf("Authorization[%d].Roles mismatch: expected %v, got %v",
						i, originalAuth.Roles, retrievedAuth.Roles)
					return false
				}

				if !stringSlicesEqual(retrievedAuth.Permissions, originalAuth.Permissions) {
					t.Logf("Authorization[%d].Permissions mismatch: expected %v, got %v",
						i, originalAuth.Permissions, retrievedAuth.Permissions)
					return false
				}
			}

			return true
		},
		genNonEmptyString(),                    // issuer
		genNonEmptyString(),                    // subject
		genAudienceSlice(),                     // audience
		gen.Int64Range(1000000000, 2000000000), // exp (reasonable timestamp range)
		gen.Int64Range(1000000000, 2000000000), // iat (reasonable timestamp range)
		genOptionalString(),                    // username
		genOptionalEmail(),                     // email
		genAuthorizationSlice(),                // authorizations
	))

	properties.Property("GetClaims returns error when claims not set", prop.ForAll(
		func() bool {
			// Create a gin context without setting claims
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Try to retrieve claims - should fail
			retrievedClaims, err := GetClaims(c)
			if err == nil {
				t.Logf("Expected error when claims not set, but got nil error")
				return false
			}
			if retrievedClaims != nil {
				t.Logf("Expected nil claims when not set, but got claims")
				return false
			}

			return true
		},
	))

	properties.Property("MustGetClaims panics when claims not set", prop.ForAll(
		func() bool {
			// Create a gin context without setting claims
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Track whether panic occurred
			panicked := false

			// Try to retrieve claims with MustGetClaims - should panic
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
					}
				}()
				_ = MustGetClaims(c)
			}()

			// Test passes if panic occurred
			if !panicked {
				t.Logf("Expected panic when claims not set, but no panic occurred")
				return false
			}

			return true
		},
	))

	properties.Property("MustGetClaims succeeds when claims are set", prop.ForAll(
		func(subject string, username string, email string) bool {
			// Create a gin context and set claims
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			originalClaims := &entity.Claims{
				Subject:  subject,
				Username: username,
				Email:    email,
			}

			SetClaims(c, originalClaims)

			// Try to retrieve claims with MustGetClaims - should succeed
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Unexpected panic when claims are set: %v", r)
				}
			}()

			retrievedClaims := MustGetClaims(c)
			if retrievedClaims == nil {
				t.Logf("Expected claims, but got nil")
				return false
			}

			if retrievedClaims.Subject != subject {
				t.Logf("Subject mismatch: expected '%s', got '%s'", subject, retrievedClaims.Subject)
				return false
			}

			return true
		},
		genNonEmptyString(),
		genOptionalString(),
		genOptionalEmail(),
	))

	properties.TestingRun(t)
}

// testAuthorization is a helper struct for generating authorization data
type testAuthorization struct {
	App         string
	Roles       []string
	Permissions []string
}

// genNonEmptyString generates non-empty strings for required fields
func genNonEmptyString() gopter.Gen {
	return gen.Identifier().Map(func(s string) string {
		if len(s) > 50 {
			return s[:50]
		}
		return s
	})
}

// genOptionalString generates strings that can be empty
func genOptionalString() gopter.Gen {
	return gen.OneGenOf(
		gen.Const(""),
		gen.Identifier().Map(func(s string) string {
			if len(s) > 50 {
				return s[:50]
			}
			return s
		}),
	)
}

// genOptionalEmail generates email-like strings
func genOptionalEmail() gopter.Gen {
	return gen.OneGenOf(
		gen.Const(""),
		gen.Identifier().Map(func(s string) string {
			if len(s) > 20 {
				s = s[:20]
			}
			return s + "@example.com"
		}),
	)
}

// genAudienceSlice generates slices of audience strings
func genAudienceSlice() gopter.Gen {
	return gen.SliceOfN(3, gen.Identifier().Map(func(s string) string {
		if len(s) > 30 {
			return s[:30]
		}
		return s
	}))
}

// genStringSlice generates slices of strings
func genStringSlice() gopter.Gen {
	return gen.SliceOfN(5, gen.Identifier().Map(func(s string) string {
		if len(s) > 20 {
			return s[:20]
		}
		return s
	}))
}

// genAuthorizationSlice generates slices of authorization structures
func genAuthorizationSlice() gopter.Gen {
	return gen.SliceOfN(3, gen.Struct(reflect.TypeOf(testAuthorization{}), map[string]gopter.Gen{
		"App": gen.OneConstOf(
			"STACKFORGE",
			"GLOBAL",
			"OTHER-APP",
		),
		"Roles":       genStringSlice(),
		"Permissions": genStringSlice(),
	}))
}

// stringSlicesEqual compares two string slices for equality
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

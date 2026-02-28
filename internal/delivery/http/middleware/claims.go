package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

// ClaimsContextKey is the key used to store claims in gin.Context
const ClaimsContextKey = "auth_claims"

// SetClaims stores claims in gin.Context
func SetClaims(c *gin.Context, claims *entity.Claims) {
	c.Set(ClaimsContextKey, claims)
}

// GetClaims retrieves claims from gin.Context
// Returns an error if claims are not found or have wrong type
func GetClaims(c *gin.Context) (*entity.Claims, error) {
	value, exists := c.Get(ClaimsContextKey)
	if !exists {
		return nil, errors.New("claims not found in context")
	}

	claims, ok := value.(*entity.Claims)
	if !ok {
		return nil, errors.New("invalid claims type in context")
	}

	return claims, nil
}

// MustGetClaims retrieves claims or panics (for use after auth middleware)
// This should only be used in handlers that are protected by auth middleware
func MustGetClaims(c *gin.Context) *entity.Claims {
	claims, err := GetClaims(c)
	if err != nil {
		panic("claims not found in context: " + err.Error())
	}
	return claims
}

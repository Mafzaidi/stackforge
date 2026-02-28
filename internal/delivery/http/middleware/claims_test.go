package middleware

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestSetClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("stores claims in context", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		claims := &entity.Claims{
			Subject:  "user123",
			Username: "testuser",
			Email:    "test@example.com",
		}

		SetClaims(c, claims)

		value, exists := c.Get(ClaimsContextKey)
		assert.True(t, exists)
		assert.Equal(t, claims, value)
	})

	t.Run("overwrites existing claims", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		claims1 := &entity.Claims{Subject: "user1"}
		claims2 := &entity.Claims{Subject: "user2"}

		SetClaims(c, claims1)
		SetClaims(c, claims2)

		value, _ := c.Get(ClaimsContextKey)
		assert.Equal(t, claims2, value)
	})
}

func TestGetClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("retrieves claims successfully", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		expectedClaims := &entity.Claims{
			Subject:  "user123",
			Username: "testuser",
			Email:    "test@example.com",
			Authorization: []entity.Authorization{
				{
					App:         "STACKFORGE",
					Roles:       []string{"user"},
					Permissions: []string{"todo.read"},
				},
			},
		}

		c.Set(ClaimsContextKey, expectedClaims)

		claims, err := GetClaims(c)
		assert.NoError(t, err)
		assert.Equal(t, expectedClaims, claims)
	})

	t.Run("returns error when claims not found", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)

		claims, err := GetClaims(c)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Equal(t, "claims not found in context", err.Error())
	})

	t.Run("returns error when claims have wrong type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		c.Set(ClaimsContextKey, "invalid type")

		claims, err := GetClaims(c)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Equal(t, "invalid claims type in context", err.Error())
	})
}

func TestMustGetClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("retrieves claims successfully", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		expectedClaims := &entity.Claims{
			Subject:  "user123",
			Username: "testuser",
			Email:    "test@example.com",
		}

		c.Set(ClaimsContextKey, expectedClaims)

		claims := MustGetClaims(c)
		assert.Equal(t, expectedClaims, claims)
	})

	t.Run("panics when claims not found", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)

		assert.Panics(t, func() {
			MustGetClaims(c)
		})
	})

	t.Run("panics when claims have wrong type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		c.Set(ClaimsContextKey, "invalid type")

		assert.Panics(t, func() {
			MustGetClaims(c)
		})
	})
}

func TestClaimsContextKey(t *testing.T) {
	t.Run("constant has expected value", func(t *testing.T) {
		assert.Equal(t, "auth_claims", ClaimsContextKey)
	})
}

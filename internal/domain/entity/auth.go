package entity

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"math/big"
)

// Claims represents the JWT token claims structure
type Claims struct {
	Subject       string
	Username      string
	Email         string
	Issuer        string
	Audience      []string
	ExpiresAt     int64
	IssuedAt      int64
	Authorization []Authorization
}

// Authorization represents app-specific roles and permissions
type Authorization struct {
	App         string
	Roles       []string
	Permissions []string
}

// JWK represents a single JSON Web Key
type JWK struct {
	KeyType   string `json:"kty"` // Key type (RSA)
	KeyID     string `json:"kid"` // Key identifier
	Use       string `json:"use"` // Key usage (sig)
	Algorithm string `json:"alg"` // Algorithm (RS256)
	N         string `json:"n"`   // RSA modulus
	E         string `json:"e"`   // RSA exponent
}

// JWKSResponse represents the JSON Web Key Set response from the JWKS endpoint
type JWKSResponse struct {
	Keys []JWK `json:"keys"` // Array of JSON Web Keys
}

// ToPublicKey converts JWK to RSA public key
func (j *JWK) ToPublicKey() (*rsa.PublicKey, error) {
	// Decode the modulus (n)
	nBytes, err := base64.RawURLEncoding.DecodeString(j.N)
	if err != nil {
		return nil, errors.New("failed to decode modulus")
	}

	// Decode the exponent (e)
	eBytes, err := base64.RawURLEncoding.DecodeString(j.E)
	if err != nil {
		return nil, errors.New("failed to decode exponent")
	}

	// Convert bytes to big.Int
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	// Create RSA public key
	publicKey := &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}

	return publicKey, nil
}

// HasRole checks if user has a specific role for the app or globally
func (c *Claims) HasRole(appCode string, role string) bool {
	for _, auth := range c.Authorization {
		if auth.App == appCode || auth.App == "GLOBAL" {
			for _, r := range auth.Roles {
				if r == role {
					return true
				}
			}
		}
	}
	return false
}

// HasPermission checks if user has a specific permission for the app or globally
func (c *Claims) HasPermission(appCode string, permission string) bool {
	for _, auth := range c.Authorization {
		if auth.App == appCode || auth.App == "GLOBAL" {
			for _, p := range auth.Permissions {
				if p == permission || p == "*" {
					return true
				}
			}
		}
	}
	return false
}

// HasAnyPermission checks if user has wildcard permission
func (c *Claims) HasAnyPermission(appCode string) bool {
	for _, auth := range c.Authorization {
		if auth.App == appCode || auth.App == "GLOBAL" {
			for _, p := range auth.Permissions {
				if p == "*" {
					return true
				}
			}
		}
	}
	return false
}

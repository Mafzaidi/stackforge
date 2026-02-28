package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

// JWTService defines the interface for JWT token validation
type JWTService interface {
	// ValidateToken validates a JWT token and returns claims
	ValidateToken(tokenString string) (*entity.Claims, error)
}

// jwtService implements the JWTService interface
type jwtService struct {
	jwksClient JWKSClient
	appCode    string
	issuer     string
}

// NewJWTService creates a new JWT service
func NewJWTService(jwksClient JWKSClient, appCode string, issuer string) JWTService {
	return &jwtService{
		jwksClient: jwksClient,
		appCode:    appCode,
		issuer:     issuer,
	}
}

// ValidateToken performs complete token validation:
// 1. Parse token and extract header
// 2. Fetch public key using kid from header
// 3. Verify signature using RS256
// 4. Validate standard claims (exp, iss, aud)
// 5. Parse and return custom claims
func (s *jwtService) ValidateToken(tokenString string) (*entity.Claims, error) {
	// Parse token without validation first to extract kid from header
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method is RS256
		if token.Method.Alg() != "RS256" {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Extract kid from header
		kid, ok := token.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, errors.New("missing or invalid kid in token header")
		}

		// Fetch public key from JWKS
		publicKey, err := s.jwksClient.GetKey(kid)
		if err != nil {
			return nil, fmt.Errorf("failed to get public key: %w", err)
		}

		return publicKey, nil
	})

	if err != nil {
		// Check if error is due to expiration
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("Token expired")
		}
		return nil, fmt.Errorf("token parsing failed: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Extract claims from token
	claims, err := s.extractClaims(token)
	if err != nil {
		return nil, fmt.Errorf("failed to extract claims: %w", err)
	}

	// Validate standard claims
	if err := s.validateClaims(claims); err != nil {
		return nil, err
	}

	return claims, nil
}

// extractClaims extracts and parses custom claims from the token
func (s *jwtService) extractClaims(token *jwt.Token) (*entity.Claims, error) {
	claimsMap, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("failed to parse token claims")
	}

	claims := &entity.Claims{}

	// Extract standard claims
	if iss, ok := claimsMap["iss"].(string); ok {
		claims.Issuer = iss
	}

	if sub, ok := claimsMap["sub"].(string); ok {
		claims.Subject = sub
	}

	// Extract audience (can be string or array)
	if aud, ok := claimsMap["aud"].([]interface{}); ok {
		for _, a := range aud {
			if audStr, ok := a.(string); ok {
				claims.Audience = append(claims.Audience, audStr)
			}
		}
	} else if audStr, ok := claimsMap["aud"].(string); ok {
		claims.Audience = []string{audStr}
	}

	// Extract timestamps
	if exp, ok := claimsMap["exp"].(float64); ok {
		claims.ExpiresAt = int64(exp)
	}

	if iat, ok := claimsMap["iat"].(float64); ok {
		claims.IssuedAt = int64(iat)
	}

	// Extract custom claims
	if username, ok := claimsMap["username"].(string); ok {
		claims.Username = username
	}

	if email, ok := claimsMap["email"].(string); ok {
		claims.Email = email
	}

	// Extract authorization array
	if authArray, ok := claimsMap["authorization"].([]interface{}); ok {
		for _, authItem := range authArray {
			if authMap, ok := authItem.(map[string]interface{}); ok {
				auth := entity.Authorization{}

				if app, ok := authMap["app"].(string); ok {
					auth.App = app
				}

				if roles, ok := authMap["roles"].([]interface{}); ok {
					for _, role := range roles {
						if roleStr, ok := role.(string); ok {
							auth.Roles = append(auth.Roles, roleStr)
						}
					}
				}

				if perms, ok := authMap["permissions"].([]interface{}); ok {
					for _, perm := range perms {
						if permStr, ok := perm.(string); ok {
							auth.Permissions = append(auth.Permissions, permStr)
						}
					}
				}

				claims.Authorization = append(claims.Authorization, auth)
			}
		}
	}

	return claims, nil
}

// validateClaims validates standard JWT claims
func (s *jwtService) validateClaims(claims *entity.Claims) error {
	// Validate expiration (already checked by jwt.Parse, but double-check)
	if claims.ExpiresAt == 0 {
		return errors.New("missing expiration claim")
	}
	if time.Now().Unix() >= claims.ExpiresAt {
		return errors.New("Token expired")
	}

	// Validate issuer
	if claims.Issuer != s.issuer {
		return errors.New("Invalid token issuer")
	}

	// Validate audience
	if len(claims.Audience) == 0 {
		return errors.New("missing audience claim")
	}

	audienceValid := false
	for _, aud := range claims.Audience {
		if aud == s.appCode {
			audienceValid = true
			break
		}
	}
	if !audienceValid {
		return errors.New("Invalid token audience")
	}

	return nil
}

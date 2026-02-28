package auth

import (
	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

// JWTValidator defines the interface for JWT token validation
type JWTValidator interface {
	ValidateToken(tokenString string) (*entity.Claims, error)
}

type callbackUseCase struct {
	jwtValidator JWTValidator
}

// NewCallbackUseCase creates a new callback use case
func NewCallbackUseCase(jwtValidator JWTValidator) CallbackUseCase {
	return &callbackUseCase{
		jwtValidator: jwtValidator,
	}
}

// ValidateToken validates the JWT token and returns claims
func (uc *callbackUseCase) ValidateToken(token string) (*entity.Claims, error) {
	return uc.jwtValidator.ValidateToken(token)
}

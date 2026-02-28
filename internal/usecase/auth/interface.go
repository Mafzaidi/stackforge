package auth

import "github.com/mafzaidi/stackforge/internal/domain/entity"

// LoginUseCase handles SSO login flow
type LoginUseCase interface {
	BuildLoginURL() (string, error)
}

// CallbackUseCase handles SSO callback and token validation
type CallbackUseCase interface {
	ValidateToken(token string) (*entity.Claims, error)
}

// LogoutUseCase handles SSO logout flow
type LogoutUseCase interface {
	BuildLogoutURL() (string, error)
}

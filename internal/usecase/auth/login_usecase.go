package auth

import (
	"net/url"
)

type loginUseCase struct {
	authorizerBaseURL string
	callbackURL       string
}

// NewLoginUseCase creates a new login use case
func NewLoginUseCase(authorizerBaseURL string, callbackURL string) LoginUseCase {
	return &loginUseCase{
		authorizerBaseURL: authorizerBaseURL,
		callbackURL:       callbackURL,
	}
}

// BuildLoginURL constructs the SSO login URL with callback parameter
func (uc *loginUseCase) BuildLoginURL() (string, error) {
	authURL, err := url.Parse(uc.authorizerBaseURL)
	if err != nil {
		return "", err
	}

	authURL.Path = "/auth/login"
	query := authURL.Query()
	query.Set("redirect_uri", uc.callbackURL)
	authURL.RawQuery = query.Encode()

	return authURL.String(), nil
}

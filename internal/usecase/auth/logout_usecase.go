package auth

import (
	"net/url"
)

type logoutUseCase struct {
	authorizerBaseURL string
	redirectURL       string
}

// NewLogoutUseCase creates a new logout use case
func NewLogoutUseCase(authorizerBaseURL string, redirectURL string) LogoutUseCase {
	return &logoutUseCase{
		authorizerBaseURL: authorizerBaseURL,
		redirectURL:       redirectURL,
	}
}

// BuildLogoutURL constructs the SSO logout URL with redirect parameter
func (uc *logoutUseCase) BuildLogoutURL() (string, error) {
	authURL, err := url.Parse(uc.authorizerBaseURL)
	if err != nil {
		return "", err
	}

	authURL.Path = "/auth/logout"
	query := authURL.Query()
	query.Set("redirect_uri", uc.redirectURL)
	authURL.RawQuery = query.Encode()

	return authURL.String(), nil
}

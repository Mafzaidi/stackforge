package service

import "context"

// AuthorizerUser represents user data fetched from the Authorizer service.
type AuthorizerUser struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	FullName    string  `json:"full_name"`
	PhoneNumber *string `json:"phone_number"`
	Email       string  `json:"email"`
}

// AuthorizerClient defines the interface for communicating with the Authorizer service.
type AuthorizerClient interface {
	// GetUserByID fetches user data from Authorizer by user ID.
	// The token is used for authentication against the Authorizer API.
	GetUserByID(ctx context.Context, userID string, token string) (*AuthorizerUser, error)
}

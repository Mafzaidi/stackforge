package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mafzaidi/stackforge/internal/domain/service"
)

// authorizerResponse represents the standard Authorizer API response envelope.
type authorizerResponse struct {
	Message string              `json:"message"`
	Data    *service.AuthorizerUser `json:"data"`
}

// authorizerClient implements service.AuthorizerClient using HTTP.
type authorizerClient struct {
	baseURL    string
	httpClient *http.Client
	logger     service.Logger
}

// NewAuthorizerClient creates a new Authorizer HTTP client.
func NewAuthorizerClient(baseURL string, logger service.Logger) service.AuthorizerClient {
	return &authorizerClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// GetUserByID fetches user data from Authorizer by user ID.
func (c *authorizerClient) GetUserByID(ctx context.Context, userID string, token string) (*service.AuthorizerUser, error) {
	url := fmt.Sprintf("%s/authorizer/v1/users/%s", c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("authorizer-client: request failed", service.Fields{
			"error":   err.Error(),
			"user_id": userID,
			"url":     url,
		})
		return nil, fmt.Errorf("authorizer request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("authorizer-client: non-200 response", service.Fields{
			"status":  resp.StatusCode,
			"user_id": userID,
		})
		return nil, fmt.Errorf("authorizer returned status %d", resp.StatusCode)
	}

	var result authorizerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode authorizer response: %w", err)
	}

	return result.Data, nil
}

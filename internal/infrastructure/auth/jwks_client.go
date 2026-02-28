package auth

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
	"github.com/mafzaidi/stackforge/internal/infrastructure/logger"
)

// JWKSClient defines the interface for fetching and caching JWKS public keys
type JWKSClient interface {
	// GetKey fetches a public key by key ID
	GetKey(kid string) (*rsa.PublicKey, error)

	// RefreshKeys forces a refresh of cached keys
	RefreshKeys() error
}

// jwksClient implements the JWKSClient interface with caching and retry logic
type jwksClient struct {
	jwksURL       string
	cache         map[string]*cachedKey
	cacheDuration time.Duration
	httpClient    *http.Client
	mu            sync.RWMutex
	fetchMu       sync.Mutex // Separate mutex to prevent concurrent fetches
	log           *logger.Logger
}

// cachedKey represents a cached public key with expiration
type cachedKey struct {
	key       *rsa.PublicKey
	expiresAt time.Time
}

// NewJWKSClient creates a new JWKS client with caching
func NewJWKSClient(jwksURL string, cacheDuration time.Duration, log *logger.Logger) JWKSClient {
	return &jwksClient{
		jwksURL:       jwksURL,
		cache:         make(map[string]*cachedKey),
		cacheDuration: cacheDuration,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		log: log,
	}
}

// GetKey retrieves a public key by key ID, using cache if available
func (c *jwksClient) GetKey(kid string) (*rsa.PublicKey, error) {
	// Check cache first
	if c.isCacheValid(kid) {
		c.mu.RLock()
		key := c.cache[kid].key
		c.mu.RUnlock()

		c.log.Info("JWKS cache hit", logger.Fields{
			"kid": kid,
		})

		return key, nil
	}

	c.log.Info("JWKS cache miss, fetching keys", logger.Fields{
		"kid": kid,
	})

	// Use fetchMu to ensure only one goroutine fetches at a time
	c.fetchMu.Lock()
	defer c.fetchMu.Unlock()

	// Double-check cache after acquiring fetchMu (another goroutine might have fetched)
	if c.isCacheValid(kid) {
		c.mu.RLock()
		key := c.cache[kid].key
		c.mu.RUnlock()
		return key, nil
	}

	// Cache miss or expired, fetch fresh keys
	keys, err := c.fetchKeysWithRetry()
	if err != nil {
		c.log.Error("JWKS fetch failed", logger.Fields{
			"kid":   kid,
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	// Update cache with all fetched keys
	c.mu.Lock()
	expiresAt := time.Now().Add(c.cacheDuration)
	for keyID, publicKey := range keys {
		c.cache[keyID] = &cachedKey{
			key:       publicKey,
			expiresAt: expiresAt,
		}
	}
	c.mu.Unlock()

	c.log.Info("JWKS keys cached successfully", logger.Fields{
		"kid":        kid,
		"key_count":  len(keys),
		"expires_at": expiresAt.Format(time.RFC3339),
	})

	// Return the requested key
	key, exists := keys[kid]
	if !exists {
		c.log.Error("JWKS key not found", logger.Fields{
			"kid":       kid,
			"available": getKeyIDs(keys),
		})
		return nil, fmt.Errorf("key ID %s not found in JWKS", kid)
	}

	return key, nil
}

// RefreshKeys forces a refresh of cached keys
func (c *jwksClient) RefreshKeys() error {
	c.log.Info("JWKS refresh requested", logger.Fields{})

	keys, err := c.fetchKeysWithRetry()
	if err != nil {
		c.log.Error("JWKS refresh failed", logger.Fields{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to refresh JWKS: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Clear old cache and update with fresh keys
	c.cache = make(map[string]*cachedKey)
	expiresAt := time.Now().Add(c.cacheDuration)
	for keyID, publicKey := range keys {
		c.cache[keyID] = &cachedKey{
			key:       publicKey,
			expiresAt: expiresAt,
		}
	}

	c.log.Info("JWKS refresh successful", logger.Fields{
		"key_count":  len(keys),
		"expires_at": expiresAt.Format(time.RFC3339),
	})

	return nil
}

// isCacheValid checks if a cached key exists and hasn't expired
func (c *jwksClient) isCacheValid(kid string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, exists := c.cache[kid]
	if !exists {
		return false
	}

	return time.Now().Before(cached.expiresAt)
}

// fetchKeysWithRetry fetches keys with exponential backoff retry logic
func (c *jwksClient) fetchKeysWithRetry() (map[string]*rsa.PublicKey, error) {
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			c.log.Warn("JWKS fetch retry attempt", logger.Fields{
				"attempt": attempt + 1,
				"max":     maxRetries,
			})
		}

		keys, err := c.fetchKeys()
		if err == nil {
			return keys, nil
		}

		lastErr = err

		// Don't retry on the last attempt
		if attempt < maxRetries-1 {
			// Exponential backoff: 1s, 2s, 4s
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			time.Sleep(backoff)
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

// fetchKeys makes HTTP request to JWKS endpoint and parses response
func (c *jwksClient) fetchKeys() (map[string]*rsa.PublicKey, error) {
	c.log.Info("JWKS fetch started", logger.Fields{
		"url": c.jwksURL,
	})

	resp, err := c.httpClient.Get(c.jwksURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var jwksResp entity.JWKSResponse
	if err := json.Unmarshal(body, &jwksResp); err != nil {
		return nil, fmt.Errorf("failed to parse JWKS response: %w", err)
	}

	if len(jwksResp.Keys) == 0 {
		return nil, errors.New("JWKS response contains no keys")
	}

	// Convert JWKs to RSA public keys
	keys := make(map[string]*rsa.PublicKey)
	for _, jwk := range jwksResp.Keys {
		if jwk.KeyID == "" {
			continue // Skip keys without ID
		}

		publicKey, err := jwk.ToPublicKey()
		if err != nil {
			return nil, fmt.Errorf("failed to convert JWK to public key (kid=%s): %w", jwk.KeyID, err)
		}

		keys[jwk.KeyID] = publicKey
	}

	if len(keys) == 0 {
		return nil, errors.New("no valid keys found in JWKS response")
	}

	c.log.Info("JWKS fetch completed", logger.Fields{
		"url":       c.jwksURL,
		"key_count": len(keys),
	})

	return keys, nil
}

// getKeyIDs extracts key IDs from a map of keys for logging
func getKeyIDs(keys map[string]*rsa.PublicKey) []string {
	ids := make([]string, 0, len(keys))
	for kid := range keys {
		ids = append(ids, kid)
	}
	return ids
}

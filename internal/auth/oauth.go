package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	tokenURL       = "https://api.myponto.com/oauth2/token"
	tokenBufferSec = 60 // refresh token 60s before expiry
)

// Token represents an OAuth2 access token.
type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	ExpiresAt   time.Time
}

// IsExpired checks if the token is expired or about to expire.
func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt.Add(-tokenBufferSec * time.Second))
}

// tokenCache caches tokens in memory.
var (
	tokenCache   = make(map[string]*Token)
	tokenCacheMu sync.RWMutex
)

// GetAccessToken retrieves an access token using client credentials.
func GetAccessToken(ctx context.Context, clientID, clientSecret string) (*Token, error) {
	cacheKey := clientID

	// Check cache
	tokenCacheMu.RLock()
	if cached, ok := tokenCache[cacheKey]; ok && !cached.IsExpired() {
		tokenCacheMu.RUnlock()

		return cached, nil
	}
	tokenCacheMu.RUnlock()

	// Fetch new token
	token, err := fetchToken(ctx, clientID, clientSecret)
	if err != nil {
		return nil, err
	}

	// Cache token
	tokenCacheMu.Lock()
	tokenCache[cacheKey] = token
	tokenCacheMu.Unlock()

	return token, nil
}

// ClearTokenCache clears the token cache.
func ClearTokenCache() {
	tokenCacheMu.Lock()
	tokenCache = make(map[string]*Token)
	tokenCacheMu.Unlock()
}

func fetchToken(ctx context.Context, clientID, clientSecret string) (*Token, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create token request: %w", err)
	}

	// Basic auth header
	auth := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed: %s - %s", resp.Status, string(body))
	}

	var token Token
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("parse token response: %w", err)
	}

	token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)

	return &token, nil
}

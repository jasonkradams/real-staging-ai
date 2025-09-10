//go:build integration

package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type TokenConfig struct {
	Auth0 struct {
		Domain       string `yaml:"domain"`
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
		Audience     string `yaml:"audience"`
		GrantType    string `yaml:"grant_type"`
	} `yaml:"auth0"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// GetTestToken returns a valid Auth0 token for integration tests
// Caches tokens in .env.test and auto-regenerates when needed
func GetTestToken() (string, error) {
	// Check cache first
	if token, valid := getCachedToken(); valid {
		return token, nil
	}

	// Generate new token
	tokenResp, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Cache the token
	if err := cacheToken(tokenResp); err != nil {
		log.Printf("Warning: failed to cache token: %v", err)
	}

	return tokenResp.AccessToken, nil
}

// getCachedToken checks .env.test for a valid cached token
func getCachedToken() (string, bool) {
	// envPath := ".env.test"

	data, err := os.ReadFile("../../secrets.yml")
	if err != nil {
		return "", false
	}

	lines := strings.Split(string(data), "\n")
	var token string
	var expiresAt time.Time

	for _, line := range lines {
		if after, ok := strings.CutPrefix(line, "AUTH0_TEST_TOKEN="); ok {
			token = after
		}
		if after, ok := strings.CutPrefix(line, "AUTH0_TOKEN_EXPIRES_AT="); ok {
			expiresAtStr := after
			if parsed, err := time.Parse(time.RFC3339, expiresAtStr); err == nil {
				expiresAt = parsed
			}
		}
	}

	// Check if token exists and has >1 hour remaining
	if token != "" && time.Until(expiresAt) > time.Hour {
		return token, true
	}

	return "", false
}

// cacheToken saves the token to .env.test
func cacheToken(tokenResp *TokenResponse) error {
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	content := fmt.Sprintf(`# Auto-generated test token - do not edit manually
AUTH0_TEST_TOKEN=%s
AUTH0_TOKEN_EXPIRES_AT=%s
`, tokenResp.AccessToken, expiresAt.Format(time.RFC3339))

	return os.WriteFile(".env.test", []byte(content), 0644)
}

// generateToken creates a new Auth0 token using secrets.yml
func generateToken() (*TokenResponse, error) {
	config, err := loadTokenConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	url := fmt.Sprintf("https://%s/oauth/token", config.Auth0.Domain)

	payload := map[string]string{
		"client_id":     config.Auth0.ClientID,
		"client_secret": config.Auth0.ClientSecret,
		"audience":      config.Auth0.Audience,
		"grant_type":    config.Auth0.GrantType,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth0 returned %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &tokenResp, nil
}

// loadTokenConfig loads Auth0 config from secrets.yml
func loadTokenConfig() (*TokenConfig, error) {
	var config TokenConfig

	data, err := os.ReadFile("../../secrets.yml")
	if err != nil {
		return nil, fmt.Errorf("failed to read secrets.yml: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

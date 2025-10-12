package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/real-staging-ai/api/internal/logging"
)

type config struct {
	Auth0 struct {
		ClientID     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
		Audience     string `yaml:"audience"`
		GrantType    string `yaml:"grant_type"`
		Domain       string `yaml:"domain"`
	} `yaml:"auth0"`
}

// make requestPayload a struct type
type requestPayload struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Audience     string `json:"audience"`
	GrantType    string `json:"grant_type"`
}

func main() {
	log := logging.Default()
	cfg := loadConfig()
	ctx := context.Background()

	domain := os.Getenv("AUTH0_DOMAIN")
	if domain == "" {
		domain = cfg.Auth0.Domain
	}
	if domain == "" {
		log.Error(ctx, "AUTH0_DOMAIN not set (set env or .secrets/auth0.yml)")
	}

	url := fmt.Sprintf("https://%s/oauth/token", domain)

	auth0Config := &requestPayload{
		ClientID:     cfg.Auth0.ClientID,
		ClientSecret: cfg.Auth0.ClientSecret,
		Audience:     cfg.Auth0.Audience,
		GrantType:    cfg.Auth0.GrantType,
	}

	// convert auth0Config to json
	payload, err := json.Marshal(auth0Config)
	if err != nil {
		log.Error(ctx, "failed to marshal auth0Config", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		log.Error(ctx, "failed to create request", err)
	}

	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error(ctx, "failed to do request", err)
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Error(ctx, "failed to close response body", err)
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error(ctx, "failed to read response body", err)
	}

	fmt.Println(string(body))
}

func loadConfig() config {
	var cfg config
	log := logging.Default()
	ctx := context.Background()

	// 1) Env-first
	if v := os.Getenv("AUTH0_CLIENT_ID"); v != "" {
		cfg.Auth0.ClientID = v
	}
	if v := os.Getenv("AUTH0_CLIENT_SECRET"); v != "" {
		cfg.Auth0.ClientSecret = v
	}
	if v := os.Getenv("AUTH0_AUDIENCE"); v != "" {
		cfg.Auth0.Audience = v
	}
	if v := os.Getenv("AUTH0_GRANT_TYPE"); v != "" {
		cfg.Auth0.GrantType = v
	}
	if v := os.Getenv("AUTH0_DOMAIN"); v != "" {
		cfg.Auth0.Domain = v
	}

	// 2) Fallback to .secrets/auth0.yml or secrets.yml if any critical is missing
	if cfg.Auth0.ClientID == "" || cfg.Auth0.ClientSecret == "" || cfg.Auth0.Audience == "" || cfg.Auth0.Domain == "" {
		paths := []string{".secrets/auth0.yml", "secrets.yml"}
		for _, p := range paths {
			// #nosec G304 -- Reading from predefined secrets file paths
			data, err := os.ReadFile(p)
			if err != nil {
				continue
			}
			if err := yaml.Unmarshal(data, &cfg); err == nil {
				break
			}
		}
	}

	// Default grant type
	if cfg.Auth0.GrantType == "" {
		cfg.Auth0.GrantType = "client_credentials"
	}

	// Validate
	if cfg.Auth0.ClientID == "" || cfg.Auth0.ClientSecret == "" || cfg.Auth0.Audience == "" {
		log.Error(ctx, "missing AUTH0_CLIENT_ID/SECRET/AUDIENCE (set env or .secrets/auth0.yml)")
	}

	return cfg
}

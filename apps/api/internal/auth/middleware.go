package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

// JWKSet represents a JSON Web Key Set
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// Auth0Config holds Auth0 configuration.
type Auth0Config struct {
	Context  context.Context
	Domain   string
	Audience string
	Issuer   string
}

// NewAuth0Config creates Auth0 configuration from provided values.
func NewAuth0Config(ctx context.Context, domain, audience string) *Auth0Config {
	return &Auth0Config{
		Context:  ctx,
		Domain:   domain,
		Audience: audience,
		Issuer:   fmt.Sprintf("https://%s/", domain),
	}
}

// JWTMiddleware creates JWT validation middleware for Auth0
func JWTMiddleware(config *Auth0Config) echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		KeyFunc: func(token *jwt.Token) (interface{}, error) {
			// Verify the signing method
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// Get the kid from token header
			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, fmt.Errorf("kid not found in token header")
			}

			// Validate audience and issuer claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return nil, fmt.Errorf("invalid token claims")
			}

			// Check audience
			aud, ok := claims["aud"].(string)
			if !ok {
				// audience might be an array
				audList, ok := claims["aud"].([]interface{})
				if !ok || len(audList) == 0 {
					return nil, fmt.Errorf("invalid or missing audience")
				}
				// Check if our audience is in the list
				found := false
				for _, a := range audList {
					if audStr, ok := a.(string); ok && audStr == config.Audience {
						found = true
						break
					}
				}
				if !found {
					return nil, fmt.Errorf("invalid audience")
				}
			} else if aud != config.Audience {
				return nil, fmt.Errorf("invalid audience")
			}

			// Check issuer
			iss, ok := claims["iss"].(string)
			if !ok || iss != config.Issuer {
				return nil, fmt.Errorf("invalid issuer")
			}

			// Get the public key from Auth0's JWKS endpoint
			return getPublicKey(config.Context, config.Domain, kid)
		},
		// Allow tokens via Authorization header or access_token query param (for browser EventSource)
		TokenLookup: "header:Authorization:Bearer ,query:access_token",
		ErrorHandler: func(c echo.Context, err error) error {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or missing JWT token")
		},
	})
}

// OptionalJWTMiddleware creates optional JWT validation middleware
func OptionalJWTMiddleware(config *Auth0Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check for Authorization header or access_token query param
			authHeader := c.Request().Header.Get("Authorization")
			tokenParam := c.QueryParam("access_token")
			hasBearer := authHeader != "" && strings.HasPrefix(authHeader, "Bearer ")
			if !hasBearer && tokenParam == "" {
				// No token provided, continue without authentication
				return next(c)
			}

			// Token provided in header or query, validate it
			jwtMiddleware := JWTMiddleware(config)
			return jwtMiddleware(next)(c)
		}
	}
}

// getPublicKey fetches and parses the public key from Auth0's JWKS endpoint
func getPublicKey(ctx context.Context, domain, kid string) (*rsa.PublicKey, error) {
	jwksURL := fmt.Sprintf("https://%s/.well-known/jwks.json", domain)
	// #nosec G107 -- URL is constructed from trusted Auth0 domain configuration
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWKS request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	var jwks JWKSet
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	// Find the key with matching kid
	for _, key := range jwks.Keys {
		if key.Kid == kid && key.Kty == "RSA" {
			return parseRSAPublicKey(key)
		}
	}

	return nil, fmt.Errorf("key with kid %s not found", kid)
}

// parseRSAPublicKey converts JWK to RSA public key
func parseRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	// Decode the modulus
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode the exponent
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert to big integers
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}

// GetUserID extracts user ID from JWT token in context
func GetUserID(c echo.Context) (string, error) {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return "", fmt.Errorf("no JWT token found in context")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid JWT claims")
	}

	// Auth0 typically uses 'sub' claim for user ID
	sub, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("sub claim not found or not a string")
	}

	return sub, nil
}

// GetUserIDOrDefault extracts user ID from JWT token in context, or returns a default test user ID
func GetUserIDOrDefault(c echo.Context) (string, error) {
	// Allow test isolation via header override
	if hv := c.Request().Header.Get("X-Test-User"); hv != "" {
		return hv, nil
	}

	// Try to get user ID from JWT token first
	userID, err := GetUserID(c)
	if err == nil {
		return userID, nil
	}

	// If no JWT token (e.g., in test environment), return default test user
	return "auth0|testuser", nil
}

// GetUserEmail extracts user email from JWT token in context
func GetUserEmail(c echo.Context) (string, error) {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return "", fmt.Errorf("no JWT token found in context")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid JWT claims")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return "", fmt.Errorf("email claim not found or not a string")
	}

	return email, nil
}

package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
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

// Auth0Config holds Auth0 configuration
type Auth0Config struct {
	Domain   string
	Audience string
	Issuer   string
}

// NewAuth0Config creates Auth0 configuration from environment variables
func NewAuth0Config() *Auth0Config {
	return &Auth0Config{
		Domain:   os.Getenv("AUTH0_DOMAIN"),
		Audience: os.Getenv("AUTH0_AUDIENCE"),
		Issuer:   fmt.Sprintf("https://%s/", os.Getenv("AUTH0_DOMAIN")),
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

			// Get the public key from Auth0's JWKS endpoint
			return getPublicKey(config.Domain, kid)
		},
		TokenLookup: "header:Authorization:Bearer ",
		ErrorHandler: func(c echo.Context, err error) error {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or missing JWT token")
		},
	})
}

// OptionalJWTMiddleware creates optional JWT validation middleware
func OptionalJWTMiddleware(config *Auth0Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if Authorization header exists
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				// No token provided, continue without authentication
				return next(c)
			}

			// Token provided, validate it
			jwtMiddleware := JWTMiddleware(config)
			return jwtMiddleware(next)(c)
		}
	}
}

// getPublicKey fetches and parses the public key from Auth0's JWKS endpoint
func getPublicKey(domain, kid string) (*rsa.PublicKey, error) {
	jwksURL := fmt.Sprintf("https://%s/.well-known/jwks.json", domain)

	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
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

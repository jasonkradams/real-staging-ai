package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewAuth0Config(t *testing.T) {
	config := NewAuth0Config("test-domain.auth0.com", "test-audience")

	assert.Equal(t, "test-domain.auth0.com", config.Domain)
	assert.Equal(t, "test-audience", config.Audience)
	assert.Equal(t, "https://test-domain.auth0.com/", config.Issuer)
}

func TestJWTMiddleware(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	goodJWK := JWK{
		Kty: "RSA",
		Kid: "test-kid",
		Use: "sig",
		N:   base64.RawURLEncoding.EncodeToString(privateKey.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes()),
	}

	type testCase struct {
		name         string
		jwk          JWK
		jwksBody     string
		setup func(
			req *http.Request, cfg *Auth0Config,
			createToken func(key *rsa.PrivateKey, kid string, expiresAt time.Time, aud string) string,
		)
		wantCode     int
		customDomain string
		errContains  string
	}

	cases := []testCase{
		{
			name: "fail: no kid in header",
			jwk:  goodJWK,
			setup: func(
				req *http.Request, cfg *Auth0Config,
				createToken func(key *rsa.PrivateKey, kid string, expiresAt time.Time, aud string) string,
			) {
				token := createToken(privateKey, "", time.Now().Add(time.Hour), cfg.Audience)
				req.Header.Set("Authorization", "Bearer "+token)
			},
			wantCode:    http.StatusUnauthorized,
			errContains: "Invalid or missing JWT token",
		},
		{
			name:     "fail: empty jwks",
			jwksBody: `{"keys":[]}`,
			setup: func(
				req *http.Request, cfg *Auth0Config,
				createToken func(key *rsa.PrivateKey, kid string, expiresAt time.Time, aud string) string,
			) {
				token := createToken(privateKey, "test-kid", time.Now().Add(time.Hour), cfg.Audience)
				req.Header.Set("Authorization", "Bearer "+token)
			},
			wantCode:    http.StatusUnauthorized,
			errContains: "Invalid or missing JWT token",
		},
		{
			name: "success: valid token",
			jwk:  goodJWK,
			setup: func(
				req *http.Request, cfg *Auth0Config,
				createToken func(key *rsa.PrivateKey, kid string, expiresAt time.Time, aud string) string,
			) {
				token := createToken(privateKey, "test-kid", time.Now().Add(time.Hour), cfg.Audience)
				req.Header.Set("Authorization", "Bearer "+token)
			},
			wantCode: http.StatusOK,
		},
		{
			name: "fail: no authorization header",
			jwk:  goodJWK,
			setup: func(
				req *http.Request, cfg *Auth0Config,
				createToken func(key *rsa.PrivateKey, kid string, expiresAt time.Time, aud string) string,
			) {
			},
			wantCode:    http.StatusUnauthorized,
			errContains: "Invalid or missing JWT token",
		},
		{
			name: "fail: unexpected signing method",
			jwk:  goodJWK,
			setup: func(
				req *http.Request, cfg *Auth0Config,
				createToken func(key *rsa.PrivateKey, kid string, expiresAt time.Time, aud string) string,
			) {
				claims := jwt.MapClaims{"exp": time.Now().Add(time.Hour).Unix()}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				signedToken, _ := token.SignedString([]byte("secret"))
				req.Header.Set("Authorization", "Bearer "+signedToken)
			},
			wantCode:    http.StatusUnauthorized,
			errContains: "Invalid or missing JWT token",
		},
		{
			name:         "fail: jwks fetch error",
			jwk:          goodJWK,
			customDomain: "invalid-domain",
			setup: func(
				req *http.Request, cfg *Auth0Config,
				createToken func(key *rsa.PrivateKey, kid string, expiresAt time.Time, aud string) string,
			) {
				token := createToken(privateKey, "test-kid", time.Now().Add(time.Hour), cfg.Audience)
				req.Header.Set("Authorization", "Bearer "+token)
			},
			wantCode:    http.StatusUnauthorized,
			errContains: "Invalid or missing JWT token",
		},
		{
			name:     "fail: jwks decode error",
			jwksBody: "{",
			setup: func(
				req *http.Request, cfg *Auth0Config,
				createToken func(key *rsa.PrivateKey, kid string, expiresAt time.Time, aud string) string,
			) {
				token := createToken(privateKey, "test-kid", time.Now().Add(time.Hour), cfg.Audience)
				req.Header.Set("Authorization", "Bearer "+token)
			},
			wantCode:    http.StatusUnauthorized,
			errContains: "Invalid or missing JWT token",
		},
		{
			name: "fail: malformed modulus in jwk",
			jwk: JWK{
				Kty: "RSA", Kid: "test-kid", N: "-!-",
				E: base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes()),
			},
			setup: func(
				req *http.Request, cfg *Auth0Config,
				createToken func(key *rsa.PrivateKey, kid string, expiresAt time.Time, aud string) string,
			) {
				token := createToken(privateKey, "test-kid", time.Now().Add(time.Hour), cfg.Audience)
				req.Header.Set("Authorization", "Bearer "+token)
			},
			wantCode:    http.StatusUnauthorized,
			errContains: "Invalid or missing JWT token",
		},
		{
			name: "fail: malformed exponent in jwk",
			jwk:  JWK{Kty: "RSA", Kid: "test-kid", N: base64.RawURLEncoding.EncodeToString(privateKey.N.Bytes()), E: "-!-"},
			setup: func(
				req *http.Request, cfg *Auth0Config,
				createToken func(key *rsa.PrivateKey, kid string, expiresAt time.Time, aud string) string,
			) {
				token := createToken(privateKey, "test-kid", time.Now().Add(time.Hour), cfg.Audience)
				req.Header.Set("Authorization", "Bearer "+token)
			},
			wantCode:    http.StatusUnauthorized,
			errContains: "Invalid or missing JWT token",
		},
		{
			name: "fail: expired token",
			jwk:  goodJWK,
			setup: func(
				req *http.Request, cfg *Auth0Config,
				createToken func(key *rsa.PrivateKey, kid string, expiresAt time.Time, aud string) string,
			) {
				token := createToken(privateKey, "test-kid", time.Now().Add(-time.Hour), cfg.Audience)
				req.Header.Set("Authorization", "Bearer "+token)
			},
			wantCode:    http.StatusUnauthorized,
			errContains: "Invalid or missing JWT token",
		},
		{
			name: "fail: wrong audience",
			jwk:  goodJWK,
			setup: func(
				req *http.Request, cfg *Auth0Config,
				createToken func(key *rsa.PrivateKey, kid string, expiresAt time.Time, aud string) string,
			) {
				token := createToken(privateKey, "test-kid", time.Now().Add(time.Hour), "wrong-audience")
				req.Header.Set("Authorization", "Bearer "+token)
			},
			wantCode:    http.StatusUnauthorized,
			errContains: "Invalid or missing JWT token",
		},
		{
			name: "success: token in query parameter",
			jwk:  goodJWK,
			setup: func(
				req *http.Request, cfg *Auth0Config,
				createToken func(key *rsa.PrivateKey, kid string, expiresAt time.Time, aud string) string,
			) {
				token := createToken(privateKey, "test-kid", time.Now().Add(time.Hour), cfg.Audience)
				req.URL.RawQuery = "access_token=" + token
			},
			wantCode: http.StatusOK,
		},
		{
			name: "fail: malformed bearer prefix",
			jwk:  goodJWK,
			setup: func(
				req *http.Request, cfg *Auth0Config,
				createToken func(key *rsa.PrivateKey, kid string, expiresAt time.Time, aud string) string,
			) {
				token := createToken(privateKey, "test-kid", time.Now().Add(time.Hour), cfg.Audience)
				req.Header.Set("Authorization", "Basic "+token)
			},
			wantCode:    http.StatusUnauthorized,
			errContains: "Invalid or missing JWT token",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			jwksBody := tc.jwksBody
			if jwksBody == "" {
				jb, err := json.Marshal(JWKSet{Keys: []JWK{tc.jwk}})
				if err != nil {
					t.Fatalf("failed to marshal jwks: %v", err)
				}
				jwksBody = string(jb)
			}

			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(jwksBody))
			}))
			defer server.Close()

			originalClient := http.DefaultClient
			http.DefaultClient = server.Client()
			defer func() { http.DefaultClient = originalClient }()

			domain := strings.TrimPrefix(server.URL, "https://")
			if tc.customDomain != "" {
				domain = tc.customDomain
			}

			config := &Auth0Config{
				Domain:   domain,
				Audience: "test-audience",
				Issuer:   fmt.Sprintf("https://%s/", domain),
			}

			createToken := func(key *rsa.PrivateKey, kid string, expiresAt time.Time, aud string) string {
				claims := jwt.MapClaims{
					"sub": "auth0|123456789", "aud": aud, "iss": config.Issuer,
					"exp": jwt.NewNumericDate(expiresAt), "iat": jwt.NewNumericDate(time.Now()),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				if kid != "" {
					token.Header["kid"] = kid
				}
				signedToken, _ := token.SignedString(key)
				return signedToken
			}

			e := echo.New()
			middleware := JWTMiddleware(config)
			handler := middleware(func(c echo.Context) error { return c.String(http.StatusOK, "success") })

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			tc.setup(req, config, createToken)

			err := handler(c)

			if tc.wantCode == http.StatusOK {
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, rec.Code)
			} else {
				httpErr, ok := err.(*echo.HTTPError)
				assert.True(t, ok, "error is not an *echo.HTTPError")
				assert.Equal(t, tc.wantCode, httpErr.Code)
				if tc.errContains != "" {
					assert.Contains(t, httpErr.Message, tc.errContains)
				}
			}
		})
	}
}

func TestOptionalJWTMiddleware(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	goodJWK := JWK{
		Kty: "RSA",
		Kid: "test-kid",
		Use: "sig",
		N:   base64.RawURLEncoding.EncodeToString(privateKey.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(privateKey.PublicKey.E)).Bytes()),
	}
	jwks, err := json.Marshal(JWKSet{Keys: []JWK{goodJWK}})
	if err != nil {
		t.Fatalf("failed to marshal jwks: %v", err)
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jwks)
	}))
	defer server.Close()

	originalClient := http.DefaultClient
	http.DefaultClient = server.Client()
	defer func() { http.DefaultClient = originalClient }()

	domain := strings.TrimPrefix(server.URL, "https://")
	config := &Auth0Config{
		Domain:   domain,
		Audience: "test-audience",
		Issuer:   fmt.Sprintf("https://%s/", domain),
	}

	createToken := func(key *rsa.PrivateKey, kid string, expiresAt time.Time, aud string) string {
		claims := jwt.MapClaims{
			"sub": "auth0|123456789", "aud": aud, "iss": config.Issuer,
			"exp": jwt.NewNumericDate(expiresAt), "iat": jwt.NewNumericDate(time.Now()),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		if kid != "" {
			token.Header["kid"] = kid
		}
		signedToken, _ := token.SignedString(key)
		return signedToken
	}

	e := echo.New()
	middleware := OptionalJWTMiddleware(config)

	type testCase struct {
		name     string
		setup    func(req *http.Request)
		wantCode int
		wantErr  bool
	}

	cases := []testCase{
		{
			name:     "success: no authorization header",
			setup:    func(req *http.Request) {},
			wantCode: http.StatusOK,
		},
		{
			name: "success: valid token",
			setup: func(req *http.Request) {
				token := createToken(privateKey, "test-kid", time.Now().Add(time.Hour), config.Audience)
				req.Header.Set("Authorization", "Bearer "+token)
			},
			wantCode: http.StatusOK,
		},
		{
			name: "fail: invalid token",
			setup: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer invalid")
			},
			wantCode: http.StatusUnauthorized,
			wantErr:  true,
		},
		{
			name: "success: token in query parameter",
			setup: func(req *http.Request) {
				token := createToken(privateKey, "test-kid", time.Now().Add(time.Hour), config.Audience)
				req.URL.RawQuery = "access_token=" + token
			},
			wantCode: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			tc.setup(req)

			handler := middleware(func(c echo.Context) error {
				return c.String(http.StatusOK, "success")
			})

			err := handler(c)
			if tc.wantErr {
				assert.Error(t, err)
				httpErr, ok := err.(*echo.HTTPError)
				assert.True(t, ok)
				assert.Equal(t, tc.wantCode, httpErr.Code)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantCode, rec.Code)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	e := echo.New()

	type testCase struct {
		name        string
		token       interface{}
		wantUserID  string
		wantErr     bool
		errContains string
	}

	cases := []testCase{
		{
			name: "valid JWT token with sub claim",
			token: &jwt.Token{
				Claims: jwt.MapClaims{"sub": "auth0|123456789"},
			},
			wantUserID: "auth0|123456789",
		},
		{
			name:        "no JWT token in context",
			token:       nil,
			wantErr:     true,
			errContains: "no JWT token found in context",
		},
		{
			name: "invalid claims type",
			token: &jwt.Token{
				Claims: &jwt.RegisteredClaims{},
			},
			wantErr:     true,
			errContains: "invalid JWT claims",
		},
		{
			name: "missing sub claim",
			token: &jwt.Token{
				Claims: jwt.MapClaims{},
			},
			wantErr:     true,
			errContains: "sub claim not found",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tc.token != nil {
				c.Set("user", tc.token)
			}

			userID, err := GetUserID(c)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, userID)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUserID, userID)
			}
		})
	}
}

func TestGetUserIDOrDefault(t *testing.T) {
	e := echo.New()
	type testCase struct {
		name       string
		setup      func(c echo.Context)
		wantUserID string
	}

	cases := []testCase{
		{
			name: "success: user in context",
			setup: func(c echo.Context) {
				token := &jwt.Token{Claims: jwt.MapClaims{"sub": "auth0|realuser"}}
				c.Set("user", token)
			},
			wantUserID: "auth0|realuser",
		},
		{
			name:       "success: no user in context, returns default",
			setup:      func(c echo.Context) {},
			wantUserID: "auth0|testuser",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			tc.setup(c)
			userID, err := GetUserIDOrDefault(c)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantUserID, userID)
		})
	}
}

func TestGetUserEmail(t *testing.T) {
	e := echo.New()
	type testCase struct {
		name        string
		token       interface{}
		wantEmail   string
		wantErr     bool
		errContains string
	}

	cases := []testCase{
		{
			name: "valid JWT token with email claim",
			token: &jwt.Token{
				Claims: jwt.MapClaims{"email": "test@example.com"},
			},
			wantEmail: "test@example.com",
		},
		{
			name:        "no JWT token in context",
			token:       nil,
			wantErr:     true,
			errContains: "no JWT token found in context",
		},
		{
			name: "invalid claims type",
			token: &jwt.Token{
				Claims: &jwt.RegisteredClaims{},
			},
			wantErr:     true,
			errContains: "invalid JWT claims",
		},
		{
			name: "missing email claim",
			token: &jwt.Token{
				Claims: jwt.MapClaims{},
			},
			wantErr:     true,
			errContains: "email claim not found",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tc.token != nil {
				c.Set("user", tc.token)
			}

			email, err := GetUserEmail(c)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Empty(t, email)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantEmail, email)
			}
		})
	}
}

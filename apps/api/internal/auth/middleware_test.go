package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewAuth0Config(t *testing.T) {
	// Set environment variables
	os.Setenv("AUTH0_DOMAIN", "test-domain.auth0.com")
	os.Setenv("AUTH0_AUDIENCE", "test-audience")
	defer func() {
		os.Unsetenv("AUTH0_DOMAIN")
		os.Unsetenv("AUTH0_AUDIENCE")
	}()

	config := NewAuth0Config()

	assert.Equal(t, "test-domain.auth0.com", config.Domain)
	assert.Equal(t, "test-audience", config.Audience)
	assert.Equal(t, "https://test-domain.auth0.com/", config.Issuer)
}

func TestOptionalJWTMiddleware(t *testing.T) {
	config := &Auth0Config{
		Domain:   "test-domain.auth0.com",
		Audience: "test-audience",
		Issuer:   "https://test-domain.auth0.com/",
	}

	e := echo.New()
	middleware := OptionalJWTMiddleware(config)

	t.Run("no authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := middleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("empty authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := middleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("invalid token format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Basic invalid")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := middleware(func(c echo.Context) error {
			return c.String(http.StatusOK, "success")
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestGetUserID(t *testing.T) {
	e := echo.New()

	t.Run("valid JWT token with sub claim", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Create a mock JWT token with claims
		token := &jwt.Token{
			Claims: jwt.MapClaims{
				"sub":   "auth0|123456789",
				"email": "test@example.com",
			},
		}
		c.Set("user", token)

		userID, err := GetUserID(c)
		assert.NoError(t, err)
		assert.Equal(t, "auth0|123456789", userID)
	})

	t.Run("no JWT token in context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		userID, err := GetUserID(c)
		assert.Error(t, err)
		assert.Empty(t, userID)
		assert.Contains(t, err.Error(), "no JWT token found in context")
	})

	t.Run("invalid JWT token type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("user", "invalid-token")

		userID, err := GetUserID(c)
		assert.Error(t, err)
		assert.Empty(t, userID)
		assert.Contains(t, err.Error(), "no JWT token found in context")
	})

	t.Run("missing sub claim", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		token := &jwt.Token{
			Claims: jwt.MapClaims{
				"email": "test@example.com",
			},
		}
		c.Set("user", token)

		userID, err := GetUserID(c)
		assert.Error(t, err)
		assert.Empty(t, userID)
		assert.Contains(t, err.Error(), "sub claim not found")
	})
}

func TestGetUserEmail(t *testing.T) {
	e := echo.New()

	t.Run("valid JWT token with email claim", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		token := &jwt.Token{
			Claims: jwt.MapClaims{
				"sub":   "auth0|123456789",
				"email": "test@example.com",
			},
		}
		c.Set("user", token)

		email, err := GetUserEmail(c)
		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", email)
	})

	t.Run("no JWT token in context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		email, err := GetUserEmail(c)
		assert.Error(t, err)
		assert.Empty(t, email)
		assert.Contains(t, err.Error(), "no JWT token found in context")
	})

	t.Run("missing email claim", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		token := &jwt.Token{
			Claims: jwt.MapClaims{
				"sub": "auth0|123456789",
			},
		}
		c.Set("user", token)

		email, err := GetUserEmail(c)
		assert.Error(t, err)
		assert.Empty(t, email)
		assert.Contains(t, err.Error(), "email claim not found")
	})
}

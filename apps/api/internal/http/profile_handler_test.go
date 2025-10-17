package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/real-staging-ai/api/internal/logging"
	"github.com/real-staging-ai/api/internal/storage/queries"
	"github.com/real-staging-ai/api/internal/user"
)

func TestNewProfileHandler(t *testing.T) {
	t.Run("success: create new profile handler", func(t *testing.T) {
		service := &user.ProfileServiceMock{}
		repo := &user.RepositoryMock{}
		log := &logging.LoggerMock{
			ErrorFunc: func(ctx context.Context, msg string, keysAndValues ...any) {},
		}
		handler := NewProfileHandler(service, repo, log)
		assert.NotNil(t, handler)
		assert.Equal(t, service, handler.profileService)
		assert.Equal(t, log, handler.log)
	})
}

func TestProfileHandler_GetProfile(t *testing.T) {
	testCases := []struct {
		name           string
		auth0Sub       string
		setupMock      func(*user.ProfileServiceMock, *user.RepositoryMock)
		expectedStatus int
		validateResp   func(*testing.T, string)
	}{
		{
			name:     "success: get profile with user_id set",
			auth0Sub: "auth0|12345",
			setupMock: func(service *user.ProfileServiceMock, repo *user.RepositoryMock) {
				repo.GetByAuth0SubFunc = func(ctx context.Context, auth0Sub string) (*queries.GetUserByAuth0SubRow, error) {
					return &queries.GetUserByAuth0SubRow{}, nil
				}
				service.GetProfileByAuth0SubFunc = func(ctx context.Context, auth0Sub string) (*user.ProfileResponse, error) {
					email := "test@example.com"
					fullName := "Test User"
					return &user.ProfileResponse{
						ID:       "550e8400-e29b-41d4-a716-446655440000",
						Role:     "user",
						Email:    &email,
						FullName: &fullName,
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body string) {
				var resp user.ProfileResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", resp.ID)
				assert.Equal(t, "user", resp.Role)
				assert.NotNil(t, resp.Email)
				assert.Equal(t, "test@example.com", *resp.Email)
				assert.NotNil(t, resp.FullName)
				assert.Equal(t, "Test User", *resp.FullName)
			},
		},
		{
			name:     "success: create user when missing",
			auth0Sub: "auth0|67890",
			setupMock: func(service *user.ProfileServiceMock, repo *user.RepositoryMock) {
				created := false
				repo.GetByAuth0SubFunc = func(ctx context.Context, sub string) (*queries.GetUserByAuth0SubRow, error) {
					return nil, pgx.ErrNoRows
				}
				repo.CreateFunc = func(
					ctx context.Context,
					auth0Sub string,
					stripeCustomerID string,
					role string,
				) (*queries.CreateUserRow, error) {
					created = true
					return &queries.CreateUserRow{}, nil
				}
				service.GetProfileByAuth0SubFunc = func(ctx context.Context, auth0Sub string) (*user.ProfileResponse, error) {
					if !created {
						t.Fatalf("expected create to be called before profile lookup")
					}
					email := "created@example.com"
					return &user.ProfileResponse{ID: "550e8400-e29b-41d4-a716-446655440010", Role: "user", Email: &email}, nil
				}
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body string) {
				var resp user.ProfileResponse
				_ = json.Unmarshal([]byte(body), &resp)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440010", resp.ID)
			},
		},
		{
			name:     "success: get profile with default test user",
			auth0Sub: "",
			setupMock: func(service *user.ProfileServiceMock, repo *user.RepositoryMock) {
				repo.GetByAuth0SubFunc = func(ctx context.Context, sub string) (*queries.GetUserByAuth0SubRow, error) {
					// Verify it gets the default test user
					assert.Equal(t, "auth0|testuser", sub)
					return &queries.GetUserByAuth0SubRow{}, nil
				}
				service.GetProfileByAuth0SubFunc = func(ctx context.Context, auth0Sub string) (*user.ProfileResponse, error) {
					email := "testuser@example.com"
					return &user.ProfileResponse{
						ID:    "550e8400-e29b-41d4-a716-446655440001",
						Role:  "user",
						Email: &email,
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body string) {
				var resp user.ProfileResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440001", resp.ID)
			},
		},
		{
			name:     "fail: user not found in database",
			auth0Sub: "auth0|12345",
			setupMock: func(service *user.ProfileServiceMock, repo *user.RepositoryMock) {
				repo.GetByAuth0SubFunc = func(ctx context.Context, auth0Sub string) (*queries.GetUserByAuth0SubRow, error) {
					return &queries.GetUserByAuth0SubRow{}, nil
				}
				service.GetProfileByAuth0SubFunc = func(ctx context.Context, auth0Sub string) (*user.ProfileResponse, error) {
					return nil, errors.New("user not found")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			validateResp: func(t *testing.T, body string) {
				assert.Contains(t, body, "Failed to retrieve profile")
			},
		},
		{
			name:     "fail: service returns error",
			auth0Sub: "auth0|12345",
			setupMock: func(service *user.ProfileServiceMock, repo *user.RepositoryMock) {
				repo.GetByAuth0SubFunc = func(ctx context.Context, auth0Sub string) (*queries.GetUserByAuth0SubRow, error) {
					return &queries.GetUserByAuth0SubRow{}, nil
				}
				service.GetProfileByAuth0SubFunc = func(ctx context.Context, auth0Sub string) (*user.ProfileResponse, error) {
					return nil, errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			validateResp: func(t *testing.T, body string) {
				assert.Contains(t, body, "Failed to retrieve profile")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/user/profile", nil)
			if tc.auth0Sub != "" {
				req.Header.Set("X-Test-User", tc.auth0Sub)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			service := &user.ProfileServiceMock{}
			repo := &user.RepositoryMock{}
			tc.setupMock(service, repo)

			log := &logging.LoggerMock{
				ErrorFunc: func(ctx context.Context, msg string, keysAndValues ...any) {},
			}

			handler := NewProfileHandler(service, repo, log)

			err := handler.GetProfile(c)

			if tc.expectedStatus == http.StatusOK {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedStatus, rec.Code)
				tc.validateResp(t, rec.Body.String())
			} else {
				assert.Error(t, err)
				httpErr := &echo.HTTPError{}
				ok := errors.As(err, &httpErr)
				assert.True(t, ok)
				assert.Equal(t, tc.expectedStatus, httpErr.Code)
				tc.validateResp(t, httpErr.Message.(string))
			}
		})
	}
}

func TestProfileHandler_UpdateProfile(t *testing.T) {
	testCases := []struct {
		name           string
		auth0Sub       string
		requestBody    interface{}
		setupMock      func(*user.ProfileServiceMock, *user.RepositoryMock)
		expectedStatus int
		validateResp   func(*testing.T, string)
	}{
		{
			name:     "success: update profile with user_id set",
			auth0Sub: "auth0|12345",
			requestBody: map[string]interface{}{
				"email":     "updated@example.com",
				"full_name": "Updated Name",
			},
			setupMock: func(service *user.ProfileServiceMock, repo *user.RepositoryMock) {
				repo.GetByAuth0SubFunc = func(ctx context.Context, auth0Sub string) (*queries.GetUserByAuth0SubRow, error) {
					return &queries.GetUserByAuth0SubRow{}, nil
				}
				service.GetProfileByAuth0SubFunc = func(ctx context.Context, auth0Sub string) (*user.ProfileResponse, error) {
					email := "old@example.com"
					return &user.ProfileResponse{
						ID:    "550e8400-e29b-41d4-a716-446655440000",
						Role:  "user",
						Email: &email,
					}, nil
				}
				//nolint:lll // Test setup with long function signature
				service.UpdateProfileFunc = func(ctx context.Context, userID string, req *user.ProfileUpdateRequest) (*user.ProfileResponse, error) {
					email := "updated@example.com"
					fullName := "Updated Name"
					return &user.ProfileResponse{
						ID:       "550e8400-e29b-41d4-a716-446655440000",
						Role:     "user",
						Email:    &email,
						FullName: &fullName,
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body string) {
				var resp user.ProfileResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", resp.ID)
				assert.NotNil(t, resp.Email)
				assert.Equal(t, "updated@example.com", *resp.Email)
				assert.NotNil(t, resp.FullName)
				assert.Equal(t, "Updated Name", *resp.FullName)
			},
		},
		{
			name:     "success: update profile with default test user",
			auth0Sub: "",
			requestBody: map[string]interface{}{
				"email": "newemail@example.com",
			},
			setupMock: func(service *user.ProfileServiceMock, repo *user.RepositoryMock) {
				repo.GetByAuth0SubFunc = func(ctx context.Context, sub string) (*queries.GetUserByAuth0SubRow, error) {
					assert.Equal(t, "auth0|testuser", sub)
					return &queries.GetUserByAuth0SubRow{}, nil
				}
				service.GetProfileByAuth0SubFunc = func(ctx context.Context, auth0Sub string) (*user.ProfileResponse, error) {
					email := "old@example.com"
					return &user.ProfileResponse{
						ID:    "550e8400-e29b-41d4-a716-446655440001",
						Role:  "user",
						Email: &email,
					}, nil
				}
				//nolint:lll // Test setup with long function signature
				service.UpdateProfileFunc = func(ctx context.Context, userID string, req *user.ProfileUpdateRequest) (*user.ProfileResponse, error) {
					email := "newemail@example.com"
					return &user.ProfileResponse{
						ID:    "550e8400-e29b-41d4-a716-446655440001",
						Role:  "user",
						Email: &email,
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body string) {
				var resp user.ProfileResponse
				err := json.Unmarshal([]byte(body), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440001", resp.ID)
			},
		},
		{
			name:        "fail: invalid request body",
			auth0Sub:    "auth0|12345",
			requestBody: "invalid json",
			setupMock: func(service *user.ProfileServiceMock, repo *user.RepositoryMock) {
				repo.GetByAuth0SubFunc = func(ctx context.Context, auth0Sub string) (*queries.GetUserByAuth0SubRow, error) {
					return &queries.GetUserByAuth0SubRow{}, nil
				}
				service.GetProfileByAuth0SubFunc = func(ctx context.Context, auth0Sub string) (*user.ProfileResponse, error) {
					email := "test@example.com"
					return &user.ProfileResponse{
						ID:    "550e8400-e29b-41d4-a716-446655440000",
						Role:  "user",
						Email: &email,
					}, nil
				}
			},
			expectedStatus: http.StatusBadRequest,
			validateResp: func(t *testing.T, body string) {
				assert.Contains(t, body, "Invalid request body")
			},
		},
		{
			name:     "fail: update service returns error",
			auth0Sub: "auth0|12345",
			requestBody: map[string]interface{}{
				"email": "test@example.com",
			},
			setupMock: func(service *user.ProfileServiceMock, repo *user.RepositoryMock) {
				repo.GetByAuth0SubFunc = func(ctx context.Context, auth0Sub string) (*queries.GetUserByAuth0SubRow, error) {
					return &queries.GetUserByAuth0SubRow{}, nil
				}
				service.GetProfileByAuth0SubFunc = func(ctx context.Context, auth0Sub string) (*user.ProfileResponse, error) {
					email := "old@example.com"
					return &user.ProfileResponse{
						ID:    "550e8400-e29b-41d4-a716-446655440000",
						Role:  "user",
						Email: &email,
					}, nil
				}
				//nolint:lll // Test setup with long function signature
				service.UpdateProfileFunc = func(ctx context.Context, userID string, req *user.ProfileUpdateRequest) (*user.ProfileResponse, error) {
					return nil, errors.New("update failed")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			validateResp: func(t *testing.T, body string) {
				assert.Contains(t, body, "Failed to update profile")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()

			var reqBody []byte
			var err error
			if str, ok := tc.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, err = json.Marshal(tc.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/user/profile", bytes.NewReader(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			if tc.auth0Sub != "" {
				req.Header.Set("X-Test-User", tc.auth0Sub)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			service := &user.ProfileServiceMock{}
			repo := &user.RepositoryMock{}
			tc.setupMock(service, repo)

			log := &logging.LoggerMock{
				ErrorFunc: func(ctx context.Context, msg string, keysAndValues ...any) {},
			}
			handler := NewProfileHandler(service, repo, log)

			err = handler.UpdateProfile(c)

			if tc.expectedStatus == http.StatusOK {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedStatus, rec.Code)
				tc.validateResp(t, rec.Body.String())
			} else {
				assert.Error(t, err)
				httpErr := &echo.HTTPError{}
				ok := errors.As(err, &httpErr)
				assert.True(t, ok)
				assert.Equal(t, tc.expectedStatus, httpErr.Code)
				tc.validateResp(t, httpErr.Message.(string))
			}
		})
	}
}

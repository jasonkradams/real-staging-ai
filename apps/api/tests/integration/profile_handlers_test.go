//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httpLib "github.com/real-staging-ai/api/internal/http"
	"github.com/real-staging-ai/api/internal/image"
	"github.com/real-staging-ai/api/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProfile_Integration(t *testing.T) {
	// Setup
	db := SetupTestDatabase(t)
	defer db.Close()

	ctx := context.Background()
	TruncateAllTables(ctx, db.Pool())
	SeedDatabase(ctx, db.Pool())

	s3ServiceMock := SetupTestS3Service(t, ctx)
	imageServiceMock := &image.ServiceMock{}
	server := httpLib.NewTestServer(db, s3ServiceMock, imageServiceMock)

	testCases := []struct {
		name           string
		userHeader     string
		expectedStatus int
		validate       func(t *testing.T, response []byte)
	}{
		{
			name:           "success: get profile for seeded user",
			userHeader:     "auth0|testuser",
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, response []byte) {
				var profile user.ProfileResponse
				err := json.Unmarshal(response, &profile)
				require.NoError(t, err)
				assert.Equal(t, "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", profile.ID)
				assert.Equal(t, "user", profile.Role)
				assert.NotNil(t, profile.StripeCustomerID)
				assert.Equal(t, "cus_test", *profile.StripeCustomerID)
			},
		},
		{
			name:           "success: get profile with default test user",
			userHeader:     "",
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, response []byte) {
				var profile user.ProfileResponse
				err := json.Unmarshal(response, &profile)
				require.NoError(t, err)
				// Default test user should return the seeded user
				assert.Equal(t, "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", profile.ID)
			},
		},
		{
			name:           "success: create user on first access",
			userHeader:     "auth0|nonexistent",
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, response []byte) {
				var profile user.ProfileResponse
				err := json.Unmarshal(response, &profile)
				require.NoError(t, err)
				// Should have created a new user with default role
				assert.Equal(t, "user", profile.Role)
				assert.NotEmpty(t, profile.ID)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/user/profile", nil)
			if tc.userHeader != "" {
				req.Header.Set("X-Test-User", tc.userHeader)
			}
			rec := httptest.NewRecorder()

			// Run the request through the server
			server.ServeHTTP(rec, req)

			// Assertions
			assert.Equal(t, tc.expectedStatus, rec.Code)
			tc.validate(t, rec.Body.Bytes())
		})
	}
}

func TestUpdateProfile_Integration(t *testing.T) {
	// Setup
	db := SetupTestDatabase(t)
	defer db.Close()

	ctx := context.Background()
	TruncateAllTables(ctx, db.Pool())
	SeedDatabase(ctx, db.Pool())

	s3ServiceMock := SetupTestS3Service(t, ctx)
	imageServiceMock := &image.ServiceMock{}
	server := httpLib.NewTestServer(db, s3ServiceMock, imageServiceMock)

	testCases := []struct {
		name           string
		userHeader     string
		requestBody    interface{}
		expectedStatus int
		validate       func(t *testing.T, response []byte)
	}{
		{
			name:       "success: update profile with all fields",
			userHeader: "auth0|testuser",
			requestBody: map[string]interface{}{
				"email":             "updated@example.com",
				"full_name":         "Updated Name",
				"company_name":      "Updated Company",
				"phone":             "+1234567890",
				"profile_photo_url": "https://example.com/photo.jpg",
				"billing_address": map[string]string{
					"line1":       "123 Main St",
					"city":        "San Francisco",
					"state":       "CA",
					"postal_code": "94102",
					"country":     "US",
				},
				"preferences": map[string]interface{}{
					"email_notifications": true,
					"marketing_emails":    false,
					"default_room_type":   "living_room",
				},
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, response []byte) {
				var profile user.ProfileResponse
				err := json.Unmarshal(response, &profile)
				require.NoError(t, err)
				assert.Equal(t, "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", profile.ID)
				assert.NotNil(t, profile.Email)
				assert.Equal(t, "updated@example.com", *profile.Email)
				assert.NotNil(t, profile.FullName)
				assert.Equal(t, "Updated Name", *profile.FullName)
				assert.NotNil(t, profile.CompanyName)
				assert.Equal(t, "Updated Company", *profile.CompanyName)
				assert.NotNil(t, profile.Phone)
				assert.Equal(t, "+1234567890", *profile.Phone)
				assert.NotNil(t, profile.ProfilePhotoURL)
				assert.Equal(t, "https://example.com/photo.jpg", *profile.ProfilePhotoURL)
				assert.NotNil(t, profile.BillingAddress)
				assert.NotNil(t, profile.Preferences)
			},
		},
		{
			name:       "success: update profile with partial fields",
			userHeader: "auth0|testuser",
			requestBody: map[string]interface{}{
				"email":     "partial@example.com",
				"full_name": "Partial Update",
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, response []byte) {
				var profile user.ProfileResponse
				err := json.Unmarshal(response, &profile)
				require.NoError(t, err)
				assert.NotNil(t, profile.Email)
				assert.Equal(t, "partial@example.com", *profile.Email)
				assert.NotNil(t, profile.FullName)
				assert.Equal(t, "Partial Update", *profile.FullName)
			},
		},
		{
			name:       "success: update only JSON fields",
			userHeader: "auth0|testuser",
			requestBody: map[string]interface{}{
				"billing_address": map[string]string{
					"line1":       "456 Oak Ave",
					"city":        "Portland",
					"state":       "OR",
					"postal_code": "97201",
					"country":     "US",
				},
			},
			expectedStatus: http.StatusOK,
			validate: func(t *testing.T, response []byte) {
				var profile user.ProfileResponse
				err := json.Unmarshal(response, &profile)
				require.NoError(t, err)
				assert.NotNil(t, profile.BillingAddress)
				
				var addr map[string]string
				err = json.Unmarshal(profile.BillingAddress, &addr)
				require.NoError(t, err)
				assert.Equal(t, "456 Oak Ave", addr["line1"])
				assert.Equal(t, "Portland", addr["city"])
			},
		},
		{
			name:       "fail: invalid email (too short)",
			userHeader: "auth0|testuser",
			requestBody: map[string]interface{}{
				"email": "ab",
			},
			expectedStatus: http.StatusInternalServerError,
			validate: func(t *testing.T, response []byte) {
				assert.Contains(t, string(response), "Failed to update profile")
			},
		},
		{
			name:       "fail: invalid phone (too long)",
			userHeader: "auth0|testuser",
			requestBody: map[string]interface{}{
				"phone": "123456789012345678901",
			},
			expectedStatus: http.StatusInternalServerError,
			validate: func(t *testing.T, response []byte) {
				assert.Contains(t, string(response), "Failed to update profile")
			},
		},
		{
			name:           "fail: invalid JSON",
			userHeader:     "auth0|testuser",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			validate: func(t *testing.T, response []byte) {
				assert.Contains(t, string(response), "Invalid request body")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var reqBody []byte
			var err error
			if str, ok := tc.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, err = json.Marshal(tc.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPatch, "/api/v1/user/profile", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			if tc.userHeader != "" {
				req.Header.Set("X-Test-User", tc.userHeader)
			}
			rec := httptest.NewRecorder()

			// Run the request through the server
			server.ServeHTTP(rec, req)

			// Assertions
			assert.Equal(t, tc.expectedStatus, rec.Code)
			tc.validate(t, rec.Body.Bytes())
		})
	}
}

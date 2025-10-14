package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"

	"github.com/real-staging-ai/api/internal/storage/queries"
)

func TestNewDefaultProfileService(t *testing.T) {
	t.Run("success: create new default profile service", func(t *testing.T) {
		repo := &RepositoryMock{}
		service := NewDefaultProfileService(repo)
		assert.NotNil(t, service)
		assert.Equal(t, repo, service.repo)
	})
}

func TestDefaultProfileService_GetProfile(t *testing.T) {
	userID := uuid.New()
	stripeCustomerID := "cus_test123"
	email := "test@example.com"
	fullName := "Test User"
	companyName := "Test Company"
	phone := "+1234567890"
	photoURL := "https://example.com/photo.jpg"
	billingAddr := []byte(`{"street":"123 Main St"}`)
	prefs := []byte(`{"theme":"dark"}`)

	testCases := []struct {
		name        string
		userID      string
		setupMock   func(*RepositoryMock)
		expectErr   bool
		errContains string
		validate    func(*testing.T, *ProfileResponse)
	}{
		{
			name:   "success: get profile with all fields",
			userID: userID.String(),
			setupMock: func(repo *RepositoryMock) {
				repo.GetProfileByIDFunc = func(ctx context.Context, userIDStr string) (*queries.GetUserProfileByIDRow, error) {
					return &queries.GetUserProfileByIDRow{
						ID:               pgtype.UUID{Bytes: userID, Valid: true},
						StripeCustomerID: pgtype.Text{String: stripeCustomerID, Valid: true},
						Role:             "user",
						Email:            pgtype.Text{String: email, Valid: true},
						FullName:         pgtype.Text{String: fullName, Valid: true},
						CompanyName:      pgtype.Text{String: companyName, Valid: true},
						Phone:            pgtype.Text{String: phone, Valid: true},
						BillingAddress:   billingAddr,
						ProfilePhotoUrl:  pgtype.Text{String: photoURL, Valid: true},
						Preferences:      prefs,
						CreatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
						UpdatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
					}, nil
				}
			},
			expectErr: false,
			validate: func(t *testing.T, profile *ProfileResponse) {
				assert.Equal(t, userID.String(), profile.ID)
				assert.Equal(t, "user", profile.Role)
				assert.NotNil(t, profile.StripeCustomerID)
				assert.Equal(t, stripeCustomerID, *profile.StripeCustomerID)
				assert.NotNil(t, profile.Email)
				assert.Equal(t, email, *profile.Email)
				assert.NotNil(t, profile.FullName)
				assert.Equal(t, fullName, *profile.FullName)
				assert.NotNil(t, profile.CompanyName)
				assert.Equal(t, companyName, *profile.CompanyName)
				assert.NotNil(t, profile.Phone)
				assert.Equal(t, phone, *profile.Phone)
				assert.NotNil(t, profile.ProfilePhotoURL)
				assert.Equal(t, photoURL, *profile.ProfilePhotoURL)
				assert.JSONEq(t, string(billingAddr), string(profile.BillingAddress))
				assert.JSONEq(t, string(prefs), string(profile.Preferences))
			},
		},
		{
			name:   "success: get profile with minimal fields",
			userID: userID.String(),
			setupMock: func(repo *RepositoryMock) {
				repo.GetProfileByIDFunc = func(ctx context.Context, userIDStr string) (*queries.GetUserProfileByIDRow, error) {
					return &queries.GetUserProfileByIDRow{
						ID:               pgtype.UUID{Bytes: userID, Valid: true},
						Role:             "user",
						StripeCustomerID: pgtype.Text{Valid: false},
						Email:            pgtype.Text{Valid: false},
						FullName:         pgtype.Text{Valid: false},
						CompanyName:      pgtype.Text{Valid: false},
						Phone:            pgtype.Text{Valid: false},
						ProfilePhotoUrl:  pgtype.Text{Valid: false},
						CreatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
						UpdatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
					}, nil
				}
			},
			expectErr: false,
			validate: func(t *testing.T, profile *ProfileResponse) {
				assert.Equal(t, userID.String(), profile.ID)
				assert.Equal(t, "user", profile.Role)
				assert.Nil(t, profile.StripeCustomerID)
				assert.Nil(t, profile.Email)
				assert.Nil(t, profile.FullName)
				assert.Nil(t, profile.CompanyName)
				assert.Nil(t, profile.Phone)
				assert.Nil(t, profile.ProfilePhotoURL)
			},
		},
		{
			name:   "fail: user not found",
			userID: userID.String(),
			setupMock: func(repo *RepositoryMock) {
				repo.GetProfileByIDFunc = func(ctx context.Context, userIDStr string) (*queries.GetUserProfileByIDRow, error) {
					return nil, pgx.ErrNoRows
				}
			},
			expectErr:   true,
			errContains: "user not found",
		},
		{
			name:   "fail: database error",
			userID: userID.String(),
			setupMock: func(repo *RepositoryMock) {
				repo.GetProfileByIDFunc = func(ctx context.Context, userIDStr string) (*queries.GetUserProfileByIDRow, error) {
					return nil, errors.New("database connection error")
				}
			},
			expectErr:   true,
			errContains: "failed to get user profile",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &RepositoryMock{}
			tc.setupMock(repo)

			service := NewDefaultProfileService(repo)
			profile, err := service.GetProfile(context.Background(), tc.userID)

			if tc.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
				assert.Nil(t, profile)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, profile)
				tc.validate(t, profile)
			}
		})
	}
}

func TestDefaultProfileService_GetProfileByAuth0Sub(t *testing.T) {
	userID := uuid.New()
	auth0Sub := "auth0|12345"
	email := "test@example.com"

	testCases := []struct {
		name        string
		auth0Sub    string
		setupMock   func(*RepositoryMock)
		expectErr   bool
		errContains string
		validate    func(*testing.T, *ProfileResponse)
	}{
		{
			name:     "success: get profile by auth0 sub",
			auth0Sub: auth0Sub,
			setupMock: func(repo *RepositoryMock) {
				repo.GetProfileByAuth0SubFunc = func(
					ctx context.Context,
					sub string,
				) (*queries.GetUserProfileByAuth0SubRow, error) {
					return &queries.GetUserProfileByAuth0SubRow{
						ID:               pgtype.UUID{Bytes: userID, Valid: true},
						StripeCustomerID: pgtype.Text{Valid: false},
						Role:             "user",
						Email:            pgtype.Text{String: email, Valid: true},
						FullName:         pgtype.Text{Valid: false},
						CompanyName:      pgtype.Text{Valid: false},
						Phone:            pgtype.Text{Valid: false},
						ProfilePhotoUrl:  pgtype.Text{Valid: false},
						CreatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
						UpdatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
					}, nil
				}
			},
			expectErr: false,
			validate: func(t *testing.T, profile *ProfileResponse) {
				assert.Equal(t, userID.String(), profile.ID)
				assert.NotNil(t, profile.Email)
				assert.Equal(t, email, *profile.Email)
			},
		},
		{
			name:     "fail: user not found",
			auth0Sub: auth0Sub,
			setupMock: func(repo *RepositoryMock) {
				repo.GetProfileByAuth0SubFunc = func(
					ctx context.Context,
					sub string,
				) (*queries.GetUserProfileByAuth0SubRow, error) {
					return nil, pgx.ErrNoRows
				}
			},
			expectErr:   true,
			errContains: "user not found",
		},
		{
			name:     "fail: database error",
			auth0Sub: auth0Sub,
			setupMock: func(repo *RepositoryMock) {
				repo.GetProfileByAuth0SubFunc = func(
					ctx context.Context,
					sub string,
				) (*queries.GetUserProfileByAuth0SubRow, error) {
					return nil, errors.New("database error")
				}
			},
			expectErr:   true,
			errContains: "failed to get user profile",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &RepositoryMock{}
			tc.setupMock(repo)

			service := NewDefaultProfileService(repo)
			profile, err := service.GetProfileByAuth0Sub(context.Background(), tc.auth0Sub)

			if tc.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
				assert.Nil(t, profile)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, profile)
				tc.validate(t, profile)
			}
		})
	}
}

func TestDefaultProfileService_UpdateProfile(t *testing.T) {
	userID := uuid.New()
	email := "updated@example.com"
	fullName := "Updated Name"
	companyName := "Updated Company"
	phone := "+9876543210"
	photoURL := "https://example.com/new-photo.jpg"
	billingAddr := []byte(`{"street":"456 Oak Ave"}`)
	prefs := []byte(`{"theme":"light"}`)

	testCases := []struct {
		name        string
		userID      string
		req         *ProfileUpdateRequest
		setupMock   func(*RepositoryMock)
		expectErr   bool
		errContains string
		validate    func(*testing.T, *ProfileResponse)
	}{
		{
			name:   "success: update all fields",
			userID: userID.String(),
			req: &ProfileUpdateRequest{
				Email:           &email,
				FullName:        &fullName,
				CompanyName:     &companyName,
				Phone:           &phone,
				ProfilePhotoURL: &photoURL,
				BillingAddress:  billingAddr,
				Preferences:     prefs,
			},
			setupMock: func(repo *RepositoryMock) {
				repo.UpdateProfileFunc = func(
					ctx context.Context,
					userIDStr string,
					update *ProfileUpdate,
				) (*queries.UpdateUserProfileRow, error) {
					return &queries.UpdateUserProfileRow{
						ID:               pgtype.UUID{Bytes: userID, Valid: true},
						StripeCustomerID: pgtype.Text{Valid: false},
						Role:             "user",
						Email:            pgtype.Text{String: email, Valid: true},
						FullName:         pgtype.Text{String: fullName, Valid: true},
						CompanyName:      pgtype.Text{String: companyName, Valid: true},
						Phone:            pgtype.Text{String: phone, Valid: true},
						ProfilePhotoUrl:  pgtype.Text{String: photoURL, Valid: true},
						BillingAddress:   billingAddr,
						Preferences:      prefs,
						CreatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
						UpdatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
					}, nil
				}
			},
			expectErr: false,
			validate: func(t *testing.T, profile *ProfileResponse) {
				assert.Equal(t, userID.String(), profile.ID)
				assert.NotNil(t, profile.Email)
				assert.Equal(t, email, *profile.Email)
				assert.NotNil(t, profile.FullName)
				assert.Equal(t, fullName, *profile.FullName)
				assert.NotNil(t, profile.CompanyName)
				assert.Equal(t, companyName, *profile.CompanyName)
				assert.NotNil(t, profile.Phone)
				assert.Equal(t, phone, *profile.Phone)
				assert.NotNil(t, profile.ProfilePhotoURL)
				assert.Equal(t, photoURL, *profile.ProfilePhotoURL)
				assert.JSONEq(t, string(billingAddr), string(profile.BillingAddress))
				assert.JSONEq(t, string(prefs), string(profile.Preferences))
			},
		},
		{
			name:   "success: update partial fields",
			userID: userID.String(),
			req: &ProfileUpdateRequest{
				Email:    &email,
				FullName: &fullName,
			},
			setupMock: func(repo *RepositoryMock) {
				repo.UpdateProfileFunc = func(
					ctx context.Context,
					userIDStr string,
					update *ProfileUpdate,
				) (*queries.UpdateUserProfileRow, error) {
					return &queries.UpdateUserProfileRow{
						ID:               pgtype.UUID{Bytes: userID, Valid: true},
						StripeCustomerID: pgtype.Text{Valid: false},
						Role:             "user",
						Email:            pgtype.Text{String: email, Valid: true},
						FullName:         pgtype.Text{String: fullName, Valid: true},
						CompanyName:      pgtype.Text{Valid: false},
						Phone:            pgtype.Text{Valid: false},
						ProfilePhotoUrl:  pgtype.Text{Valid: false},
						CreatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
						UpdatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
					}, nil
				}
			},
			expectErr: false,
			validate: func(t *testing.T, profile *ProfileResponse) {
				assert.Equal(t, userID.String(), profile.ID)
				assert.NotNil(t, profile.Email)
				assert.Equal(t, email, *profile.Email)
				assert.NotNil(t, profile.FullName)
				assert.Equal(t, fullName, *profile.FullName)
			},
		},
		{
			name:   "fail: email too short",
			userID: userID.String(),
			req: &ProfileUpdateRequest{
				Email: stringPtr("ab"),
			},
			setupMock:   func(repo *RepositoryMock) {},
			expectErr:   true,
			errContains: "email must be between 3 and 255 characters",
		},
		{
			name:   "fail: email too long",
			userID: userID.String(),
			req: &ProfileUpdateRequest{
				Email: stringPtr(string(make([]byte, 256))),
			},
			setupMock:   func(repo *RepositoryMock) {},
			expectErr:   true,
			errContains: "email must be between 3 and 255 characters",
		},
		{
			name:   "fail: phone too long",
			userID: userID.String(),
			req: &ProfileUpdateRequest{
				Phone: stringPtr("123456789012345678901"),
			},
			setupMock:   func(repo *RepositoryMock) {},
			expectErr:   true,
			errContains: "phone number must be less than 20 characters",
		},
		{
			name:   "fail: full name too long",
			userID: userID.String(),
			req: &ProfileUpdateRequest{
				FullName: stringPtr(string(make([]byte, 101))),
			},
			setupMock:   func(repo *RepositoryMock) {},
			expectErr:   true,
			errContains: "full name must be less than 100 characters",
		},
		{
			name:   "fail: company name too long",
			userID: userID.String(),
			req: &ProfileUpdateRequest{
				CompanyName: stringPtr(string(make([]byte, 101))),
			},
			setupMock:   func(repo *RepositoryMock) {},
			expectErr:   true,
			errContains: "company name must be less than 100 characters",
		},
		{
			name:   "fail: user not found",
			userID: userID.String(),
			req: &ProfileUpdateRequest{
				Email: &email,
			},
			setupMock: func(repo *RepositoryMock) {
				repo.UpdateProfileFunc = func(
					ctx context.Context,
					userIDStr string,
					update *ProfileUpdate,
				) (*queries.UpdateUserProfileRow, error) {
					return nil, pgx.ErrNoRows
				}
			},
			expectErr:   true,
			errContains: "user not found",
		},
		{
			name:   "fail: database error",
			userID: userID.String(),
			req: &ProfileUpdateRequest{
				Email: &email,
			},
			setupMock: func(repo *RepositoryMock) {
				repo.UpdateProfileFunc = func(
					ctx context.Context,
					userIDStr string,
					update *ProfileUpdate,
				) (*queries.UpdateUserProfileRow, error) {
					return nil, errors.New("database error")
				}
			},
			expectErr:   true,
			errContains: "failed to update user profile",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &RepositoryMock{}
			tc.setupMock(repo)

			service := NewDefaultProfileService(repo)
			profile, err := service.UpdateProfile(context.Background(), tc.userID, tc.req)

			if tc.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
				assert.Nil(t, profile)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, profile)
				tc.validate(t, profile)
			}
		})
	}
}

//nolint:dupl
func TestConvertProfileRowToResponse(t *testing.T) {
	t.Run("success: convert GetUserProfileByIDRow", func(t *testing.T) {
		userID := uuid.New()
		email := "test@example.com"
		row := &queries.GetUserProfileByIDRow{
			ID:               pgtype.UUID{Bytes: userID, Valid: true},
			StripeCustomerID: pgtype.Text{Valid: false},
			Role:             "user",
			Email:            pgtype.Text{String: email, Valid: true},
			FullName:         pgtype.Text{Valid: false},
			CompanyName:      pgtype.Text{Valid: false},
			Phone:            pgtype.Text{Valid: false},
			ProfilePhotoUrl:  pgtype.Text{Valid: false},
			CreatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
			UpdatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}

		result := convertProfileRowToResponse(row)
		assert.NotNil(t, result)
		assert.Equal(t, userID.String(), result.ID)
		assert.Equal(t, "user", result.Role)
		assert.NotNil(t, result.Email)
		assert.Equal(t, email, *result.Email)
	})

	t.Run("success: convert GetUserProfileByAuth0SubRow", func(t *testing.T) {
		userID := uuid.New()
		email := "test@example.com"
		row := &queries.GetUserProfileByAuth0SubRow{
			ID:               pgtype.UUID{Bytes: userID, Valid: true},
			StripeCustomerID: pgtype.Text{Valid: false},
			Role:             "user",
			Email:            pgtype.Text{String: email, Valid: true},
			FullName:         pgtype.Text{Valid: false},
			CompanyName:      pgtype.Text{Valid: false},
			Phone:            pgtype.Text{Valid: false},
			ProfilePhotoUrl:  pgtype.Text{Valid: false},
			CreatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
			UpdatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}

		result := convertProfileRowToResponse(row)
		assert.NotNil(t, result)
		assert.Equal(t, userID.String(), result.ID)
		assert.Equal(t, "user", result.Role)
		assert.NotNil(t, result.Email)
		assert.Equal(t, email, *result.Email)
	})

	t.Run("fail: unsupported type returns nil", func(t *testing.T) {
		result := convertProfileRowToResponse("invalid type")
		assert.Nil(t, result)
	})
}

//nolint:dupl
func TestConvertUpdateProfileRowToResponse(t *testing.T) {
	t.Run("success: convert UpdateUserProfileRow", func(t *testing.T) {
		userID := uuid.New()
		email := "test@example.com"
		row := &queries.UpdateUserProfileRow{
			ID:               pgtype.UUID{Bytes: userID, Valid: true},
			StripeCustomerID: pgtype.Text{Valid: false},
			Role:             "user",
			Email:            pgtype.Text{String: email, Valid: true},
			FullName:         pgtype.Text{Valid: false},
			CompanyName:      pgtype.Text{Valid: false},
			Phone:            pgtype.Text{Valid: false},
			ProfilePhotoUrl:  pgtype.Text{Valid: false},
			CreatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
			UpdatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}

		result := convertUpdateProfileRowToResponse(row)
		assert.NotNil(t, result)
		assert.Equal(t, userID.String(), result.ID)
		assert.Equal(t, "user", result.Role)
		assert.NotNil(t, result.Email)
		assert.Equal(t, email, *result.Email)
	})
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

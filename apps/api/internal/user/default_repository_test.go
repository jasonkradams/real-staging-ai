package user

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"

	"github.com/real-staging-ai/api/internal/storage/queries"
)

// Use moq-generated QuerierMock from queries package for tests
type mockQuerier = queries.QuerierMock

func TestNewUserRepository(t *testing.T) {
	repo := NewDefaultRepository(nil)
	assert.NotNil(t, repo)
}

func TestDefaultRepository_Create(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	userID := uuid.New()

	type testCase struct {
		name        string
		auth0Sub    string
		stripeID    string
		role        string
		setupMock   func(mock *mockQuerier)
		wantErr     bool
		errContains string
	}

	cases := []testCase{
		{
			name:     "success: create user with stripe id",
			auth0Sub: "auth0|123",
			stripeID: "cus_123",
			role:     "user",
			setupMock: func(mock *mockQuerier) {
				mock.CreateUserFunc = func(ctx context.Context, arg queries.CreateUserParams) (*queries.User, error) {
					return &queries.User{
						ID:               pgtype.UUID{Bytes: userID, Valid: true},
						Auth0Sub:         "auth0|123",
						StripeCustomerID: pgtype.Text{String: "cus_123", Valid: true},
						Role:             "user",
						CreatedAt:        pgtype.Timestamptz{Time: now, Valid: true},
					}, nil
				}
			},
		},
		{
			name:     "fail: db error",
			auth0Sub: "auth0|123",
			stripeID: "cus_123",
			role:     "user",
			setupMock: func(mock *mockQuerier) {
				mock.CreateUserFunc = func(ctx context.Context, arg queries.CreateUserParams) (*queries.User, error) {
					return nil, fmt.Errorf("db error")
				}
			},
			wantErr:     true,
			errContains: "unable to create user",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockQuerier{}
			tc.setupMock(mock)

			repo := &DefaultRepository{queries: mock}
			_, err := repo.Create(ctx, tc.auth0Sub, tc.stripeID, tc.role)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	type testCase struct {
		name        string
		userID      string
		setupMock   func(mock *mockQuerier)
		wantErr     bool
		errContains string
	}

	cases := []testCase{
		{
			name:   "success: get user by id",
			userID: userID.String(),
			setupMock: func(mock *mockQuerier) {
				mock.GetUserByIDFunc = func(ctx context.Context, id pgtype.UUID) (*queries.User, error) {
					return &queries.User{ID: id}, nil
				}
			},
		},
		{
			name:        "fail: invalid uuid",
			userID:      "invalid",
			setupMock:   func(mock *mockQuerier) {},
			wantErr:     true,
			errContains: "invalid user ID format",
		},
		{
			name:   "fail: not found",
			userID: userID.String(),
			setupMock: func(mock *mockQuerier) {
				mock.GetUserByIDFunc = func(ctx context.Context, id pgtype.UUID) (*queries.User, error) {
					return nil, pgx.ErrNoRows
				}
			},
			wantErr:     true,
			errContains: pgx.ErrNoRows.Error(),
		},
		{
			name:   "fail: db error",
			userID: userID.String(),
			setupMock: func(mock *mockQuerier) {
				mock.GetUserByIDFunc = func(ctx context.Context, id pgtype.UUID) (*queries.User, error) {
					return nil, fmt.Errorf("db error")
				}
			},
			wantErr:     true,
			errContains: "unable to get user by ID",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockQuerier{}
			tc.setupMock(mock)

			repo := &DefaultRepository{queries: mock}
			_, err := repo.GetByID(ctx, tc.userID)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultRepository_GetByAuth0Sub(t *testing.T) {
	ctx := context.Background()
	type testCase struct {
		name        string
		auth0Sub    string
		setupMock   func(mock *mockQuerier)
		wantErr     bool
		errContains string
	}

	cases := []testCase{
		{
			name:     "success: get user by auth0 sub",
			auth0Sub: "auth0|123",
			setupMock: func(mock *mockQuerier) {
				mock.GetUserByAuth0SubFunc = func(ctx context.Context, auth0Sub string) (*queries.User, error) {
					return &queries.User{Auth0Sub: auth0Sub}, nil
				}
			},
		},
		{
			name:     "fail: not found",
			auth0Sub: "auth0|123",
			setupMock: func(mock *mockQuerier) {
				mock.GetUserByAuth0SubFunc = func(ctx context.Context, auth0Sub string) (*queries.User, error) {
					return nil, pgx.ErrNoRows
				}
			},
			wantErr:     true,
			errContains: pgx.ErrNoRows.Error(),
		},
		{
			name:     "fail: db error",
			auth0Sub: "auth0|123",
			setupMock: func(mock *mockQuerier) {
				mock.GetUserByAuth0SubFunc = func(ctx context.Context, auth0Sub string) (*queries.User, error) {
					return nil, fmt.Errorf("db error")
				}
			},
			wantErr:     true,
			errContains: "unable to get user by Auth0 sub",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockQuerier{}
			tc.setupMock(mock)

			repo := &DefaultRepository{queries: mock}
			_, err := repo.GetByAuth0Sub(ctx, tc.auth0Sub)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultRepository_GetByStripeCustomerID(t *testing.T) {
	ctx := context.Background()
	type testCase struct {
		name             string
		stripeCustomerID string
		setupMock        func(mock *mockQuerier)
		wantErr          bool
		errContains      string
	}

	cases := []testCase{
		{
			name:             "success: get user by stripe customer id",
			stripeCustomerID: "cus_123",
			setupMock: func(mock *mockQuerier) {
				mock.GetUserByStripeCustomerIDFunc = func(
					ctx context.Context,
					stripeCustomerID pgtype.Text,
				) (*queries.User, error) {
					return &queries.User{StripeCustomerID: stripeCustomerID}, nil
				}
			},
		},
		{
			name:             "fail: not found",
			stripeCustomerID: "cus_123",
			setupMock: func(mock *mockQuerier) {
				mock.GetUserByStripeCustomerIDFunc = func(
					ctx context.Context, stripeCustomerID pgtype.Text,
				) (*queries.User, error) {
					return nil, pgx.ErrNoRows
				}
			},
			wantErr:     true,
			errContains: pgx.ErrNoRows.Error(),
		},
		{
			name:             "fail: db error",
			stripeCustomerID: "cus_123",
			setupMock: func(mock *mockQuerier) {
				mock.GetUserByStripeCustomerIDFunc = func(
					ctx context.Context, stripeCustomerID pgtype.Text,
				) (*queries.User, error) {
					return nil, fmt.Errorf("db error")
				}
			},
			wantErr:     true,
			errContains: "unable to get user by Stripe customer ID",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockQuerier{}
			tc.setupMock(mock)

			repo := &DefaultRepository{queries: mock}
			_, err := repo.GetByStripeCustomerID(ctx, tc.stripeCustomerID)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultRepository_UpdateStripeCustomerID(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	type testCase struct {
		name             string
		userID           string
		stripeCustomerID string
		setupMock        func(mock *mockQuerier)
		wantErr          bool
		errContains      string
	}

	cases := []testCase{
		{
			name:             "success: update stripe customer id",
			userID:           userID.String(),
			stripeCustomerID: "cus_123",
			setupMock: func(mock *mockQuerier) {
				mock.UpdateUserStripeCustomerIDFunc = func(
					ctx context.Context,
					arg queries.UpdateUserStripeCustomerIDParams,
				) (*queries.User, error) {
					return &queries.User{}, nil
				}
			},
		},
		{
			name:        "fail: invalid id",
			userID:      "invalid",
			setupMock:   func(mock *mockQuerier) {},
			wantErr:     true,
			errContains: "invalid user ID format",
		},
		{
			name:   "fail: not found",
			userID: userID.String(),
			setupMock: func(mock *mockQuerier) {
				mock.UpdateUserStripeCustomerIDFunc = func(
					ctx context.Context,
					arg queries.UpdateUserStripeCustomerIDParams,
				) (*queries.User, error) {
					return nil, pgx.ErrNoRows
				}
			},
			wantErr:     true,
			errContains: pgx.ErrNoRows.Error(),
		},
		{
			name:   "fail: db error",
			userID: userID.String(),
			setupMock: func(mock *mockQuerier) {
				mock.UpdateUserStripeCustomerIDFunc = func(
					ctx context.Context,
					arg queries.UpdateUserStripeCustomerIDParams,
				) (*queries.User, error) {
					return nil, fmt.Errorf("db error")
				}
			},
			wantErr:     true,
			errContains: "unable to update user Stripe customer ID",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockQuerier{}
			tc.setupMock(mock)

			repo := &DefaultRepository{queries: mock}
			_, err := repo.UpdateStripeCustomerID(ctx, tc.userID, tc.stripeCustomerID)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultRepository_UpdateRole(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	type testCase struct {
		name        string
		userID      string
		role        string
		setupMock   func(mock *mockQuerier)
		wantErr     bool
		errContains string
	}

	cases := []testCase{
		{
			name:   "success: update role",
			userID: userID.String(),
			role:   "admin",
			setupMock: func(mock *mockQuerier) {
				mock.UpdateUserRoleFunc = func(ctx context.Context, arg queries.UpdateUserRoleParams) (*queries.User, error) {
					return &queries.User{}, nil
				}
			},
		},
		{
			name:        "fail: invalid id",
			userID:      "invalid",
			role:        "admin",
			setupMock:   func(mock *mockQuerier) {},
			wantErr:     true,
			errContains: "invalid user ID format",
		},
		{
			name:   "fail: not found",
			userID: userID.String(),
			role:   "admin",
			setupMock: func(mock *mockQuerier) {
				mock.UpdateUserRoleFunc = func(ctx context.Context, arg queries.UpdateUserRoleParams) (*queries.User, error) {
					return nil, pgx.ErrNoRows
				}
			},
			wantErr:     true,
			errContains: pgx.ErrNoRows.Error(),
		},
		{
			name:   "fail: db error",
			userID: userID.String(),
			role:   "admin",
			setupMock: func(mock *mockQuerier) {
				mock.UpdateUserRoleFunc = func(ctx context.Context, arg queries.UpdateUserRoleParams) (*queries.User, error) {
					return nil, fmt.Errorf("db error")
				}
			},
			wantErr:     true,
			errContains: "unable to update user role",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockQuerier{}
			tc.setupMock(mock)

			repo := &DefaultRepository{queries: mock}
			_, err := repo.UpdateRole(ctx, tc.userID, tc.role)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultRepository_Delete(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	type testCase struct {
		name        string
		userID      string
		setupMock   func(mock *mockQuerier)
		wantErr     bool
		errContains string
	}

	cases := []testCase{
		{
			name:   "success: delete user",
			userID: userID.String(),
			setupMock: func(mock *mockQuerier) {
				mock.DeleteUserFunc = func(ctx context.Context, id pgtype.UUID) error {
					return nil
				}
			},
		},
		{
			name:        "fail: invalid id",
			userID:      "invalid",
			setupMock:   func(mock *mockQuerier) {},
			wantErr:     true,
			errContains: "invalid user ID format",
		},
		{
			name:   "fail: db error",
			userID: userID.String(),
			setupMock: func(mock *mockQuerier) {
				mock.DeleteUserFunc = func(ctx context.Context, id pgtype.UUID) error {
					return fmt.Errorf("db error")
				}
			},
			wantErr:     true,
			errContains: "unable to delete user",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockQuerier{}
			tc.setupMock(mock)

			repo := &DefaultRepository{queries: mock}
			err := repo.Delete(ctx, tc.userID)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultRepository_List(t *testing.T) {
	ctx := context.Background()
	type testCase struct {
		name      string
		limit     int
		offset    int
		setupMock func(mock *mockQuerier)
		wantErr   bool
	}

	cases := []testCase{
		{
			name:   "success: list users",
			limit:  10,
			offset: 0,
			setupMock: func(mock *mockQuerier) {
				mock.ListUsersFunc = func(ctx context.Context, arg queries.ListUsersParams) ([]*queries.User, error) {
					return []*queries.User{}, nil
				}
			},
		},
		{
			name:   "fail: db error",
			limit:  10,
			offset: 0,
			setupMock: func(mock *mockQuerier) {
				mock.ListUsersFunc = func(ctx context.Context, arg queries.ListUsersParams) ([]*queries.User, error) {
					return nil, fmt.Errorf("db error")
				}
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockQuerier{}
			tc.setupMock(mock)

			repo := &DefaultRepository{queries: mock}
			_, err := repo.List(ctx, tc.limit, tc.offset)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultRepository_Count(t *testing.T) {
	ctx := context.Background()
	type testCase struct {
		name      string
		setupMock func(mock *mockQuerier)
		wantErr   bool
	}

	cases := []testCase{
		{
			name: "success: count users",
			setupMock: func(mock *mockQuerier) {
				mock.CountUsersFunc = func(ctx context.Context) (int64, error) {
					return 5, nil
				}
			},
		},
		{
			name: "fail: db error",
			setupMock: func(mock *mockQuerier) {
				mock.CountUsersFunc = func(ctx context.Context) (int64, error) {
					return 0, fmt.Errorf("db error")
				}
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockQuerier{}
			tc.setupMock(mock)

			repo := &DefaultRepository{queries: mock}
			_, err := repo.Count(ctx)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

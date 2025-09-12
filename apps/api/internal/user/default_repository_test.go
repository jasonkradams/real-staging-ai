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
	"github.com/virtual-staging-ai/api/internal/storage/queries"
)

// mockQuerier is a mock of the queries.Querier interface
type mockQuerier struct {
	createUserFunc                 func(ctx context.Context, arg queries.CreateUserParams) (*queries.User, error)
	getUserByIDFunc                func(ctx context.Context, id pgtype.UUID) (*queries.User, error)
	getUserByAuth0SubFunc          func(ctx context.Context, auth0Sub string) (*queries.User, error)
	getUserByStripeCustomerIDFunc  func(ctx context.Context, stripeCustomerID pgtype.Text) (*queries.User, error)
	updateUserStripeCustomerIDFunc func(ctx context.Context, arg queries.UpdateUserStripeCustomerIDParams) (*queries.User, error)
	updateUserRoleFunc             func(ctx context.Context, arg queries.UpdateUserRoleParams) (*queries.User, error)
	deleteUserFunc                 func(ctx context.Context, id pgtype.UUID) error
	listUsersFunc                  func(ctx context.Context, arg queries.ListUsersParams) ([]*queries.User, error)
	countUsersFunc                 func(ctx context.Context) (int64, error)
	// Other methods to satisfy the Querier interface
	completeJobFunc              func(ctx context.Context, id pgtype.UUID) (*queries.Job, error)
	countProjectsByUserIDFunc    func(ctx context.Context, userID pgtype.UUID) (int64, error)
	createImageFunc              func(ctx context.Context, arg queries.CreateImageParams) (*queries.Image, error)
	createJobFunc                func(ctx context.Context, arg queries.CreateJobParams) (*queries.Job, error)
	createProjectFunc            func(ctx context.Context, arg queries.CreateProjectParams) (*queries.CreateProjectRow, error)
	deleteImageFunc              func(ctx context.Context, id pgtype.UUID) error
	deleteImagesByProjectIDFunc  func(ctx context.Context, projectID pgtype.UUID) error
	deleteJobFunc                func(ctx context.Context, id pgtype.UUID) error
	deleteJobsByImageIDFunc      func(ctx context.Context, imageID pgtype.UUID) error
	deleteProjectFunc            func(ctx context.Context, id pgtype.UUID) error
	deleteProjectByUserIDFunc    func(ctx context.Context, arg queries.DeleteProjectByUserIDParams) error
	failJobFunc                  func(ctx context.Context, arg queries.FailJobParams) (*queries.Job, error)
	getAllProjectsFunc           func(ctx context.Context) ([]*queries.GetAllProjectsRow, error)
	getImageByIDFunc             func(ctx context.Context, id pgtype.UUID) (*queries.Image, error)
	getImagesByProjectIDFunc     func(ctx context.Context, projectID pgtype.UUID) ([]*queries.Image, error)
	getJobByIDFunc               func(ctx context.Context, id pgtype.UUID) (*queries.Job, error)
	getJobsByImageIDFunc         func(ctx context.Context, imageID pgtype.UUID) ([]*queries.Job, error)
	getPendingJobsFunc           func(ctx context.Context, limit int32) ([]*queries.Job, error)
	getProjectByIDFunc           func(ctx context.Context, id pgtype.UUID) (*queries.GetProjectByIDRow, error)
	getProjectsByUserIDFunc      func(ctx context.Context, userID pgtype.UUID) ([]*queries.GetProjectsByUserIDRow, error)
	startJobFunc                 func(ctx context.Context, id pgtype.UUID) (*queries.Job, error)
	updateImageStatusFunc        func(ctx context.Context, arg queries.UpdateImageStatusParams) (*queries.Image, error)
	updateImageWithErrorFunc     func(ctx context.Context, arg queries.UpdateImageWithErrorParams) (*queries.Image, error)
	updateImageWithStagedURLFunc func(ctx context.Context, arg queries.UpdateImageWithStagedURLParams) (*queries.Image, error)
	updateJobStatusFunc          func(ctx context.Context, arg queries.UpdateJobStatusParams) (*queries.Job, error)
	updateProjectFunc            func(ctx context.Context, arg queries.UpdateProjectParams) (*queries.UpdateProjectRow, error)
	updateProjectByUserIDFunc    func(ctx context.Context, arg queries.UpdateProjectByUserIDParams) (*queries.UpdateProjectByUserIDRow, error)
}

func (m *mockQuerier) CreateUser(ctx context.Context, arg queries.CreateUserParams) (*queries.User, error) {
	return m.createUserFunc(ctx, arg)
}
func (m *mockQuerier) GetUserByID(ctx context.Context, id pgtype.UUID) (*queries.User, error) {
	return m.getUserByIDFunc(ctx, id)
}
func (m *mockQuerier) GetUserByAuth0Sub(ctx context.Context, auth0Sub string) (*queries.User, error) {
	return m.getUserByAuth0SubFunc(ctx, auth0Sub)
}
func (m *mockQuerier) GetUserByStripeCustomerID(ctx context.Context, stripeCustomerID pgtype.Text) (*queries.User, error) {
	return m.getUserByStripeCustomerIDFunc(ctx, stripeCustomerID)
}
func (m *mockQuerier) UpdateUserStripeCustomerID(ctx context.Context, arg queries.UpdateUserStripeCustomerIDParams) (*queries.User, error) {
	return m.updateUserStripeCustomerIDFunc(ctx, arg)
}
func (m *mockQuerier) UpdateUserRole(ctx context.Context, arg queries.UpdateUserRoleParams) (*queries.User, error) {
	return m.updateUserRoleFunc(ctx, arg)
}
func (m *mockQuerier) DeleteUser(ctx context.Context, id pgtype.UUID) error {
	return m.deleteUserFunc(ctx, id)
}
func (m *mockQuerier) ListUsers(ctx context.Context, arg queries.ListUsersParams) ([]*queries.User, error) {
	return m.listUsersFunc(ctx, arg)
}
func (m *mockQuerier) CountUsers(ctx context.Context) (int64, error) {
	return m.countUsersFunc(ctx)
}
func (m *mockQuerier) CompleteJob(ctx context.Context, id pgtype.UUID) (*queries.Job, error) {
	return nil, nil
}
func (m *mockQuerier) CountProjectsByUserID(ctx context.Context, userID pgtype.UUID) (int64, error) {
	return 0, nil
}
func (m *mockQuerier) CreateImage(ctx context.Context, arg queries.CreateImageParams) (*queries.Image, error) {
	return nil, nil
}
func (m *mockQuerier) CreateJob(ctx context.Context, arg queries.CreateJobParams) (*queries.Job, error) {
	return nil, nil
}
func (m *mockQuerier) CreateProject(ctx context.Context, arg queries.CreateProjectParams) (*queries.CreateProjectRow, error) {
	return nil, nil
}
func (m *mockQuerier) DeleteImage(ctx context.Context, id pgtype.UUID) error {
	return nil
}
func (m *mockQuerier) DeleteImagesByProjectID(ctx context.Context, projectID pgtype.UUID) error {
	return nil
}
func (m *mockQuerier) DeleteJob(ctx context.Context, id pgtype.UUID) error {
	return nil
}
func (m *mockQuerier) DeleteJobsByImageID(ctx context.Context, imageID pgtype.UUID) error {
	return nil
}
func (m *mockQuerier) DeleteProject(ctx context.Context, id pgtype.UUID) error {
	return nil
}
func (m *mockQuerier) DeleteProjectByUserID(ctx context.Context, arg queries.DeleteProjectByUserIDParams) error {
	return nil
}
func (m *mockQuerier) FailJob(ctx context.Context, arg queries.FailJobParams) (*queries.Job, error) {
	return nil, nil
}
func (m *mockQuerier) GetAllProjects(ctx context.Context) ([]*queries.GetAllProjectsRow, error) {
	return nil, nil
}
func (m *mockQuerier) GetImageByID(ctx context.Context, id pgtype.UUID) (*queries.Image, error) {
	return nil, nil
}
func (m *mockQuerier) GetImagesByProjectID(ctx context.Context, projectID pgtype.UUID) ([]*queries.Image, error) {
	return nil, nil
}
func (m *mockQuerier) GetJobByID(ctx context.Context, id pgtype.UUID) (*queries.Job, error) {
	return nil, nil
}
func (m *mockQuerier) GetJobsByImageID(ctx context.Context, imageID pgtype.UUID) ([]*queries.Job, error) {
	return nil, nil
}
func (m *mockQuerier) GetPendingJobs(ctx context.Context, limit int32) ([]*queries.Job, error) {
	return nil, nil
}
func (m *mockQuerier) GetProjectByID(ctx context.Context, id pgtype.UUID) (*queries.GetProjectByIDRow, error) {
	return nil, nil
}
func (m *mockQuerier) GetProjectsByUserID(ctx context.Context, userID pgtype.UUID) ([]*queries.GetProjectsByUserIDRow, error) {
	return nil, nil
}
func (m *mockQuerier) StartJob(ctx context.Context, id pgtype.UUID) (*queries.Job, error) {
	return nil, nil
}
func (m *mockQuerier) UpdateImageStatus(ctx context.Context, arg queries.UpdateImageStatusParams) (*queries.Image, error) {
	return nil, nil
}
func (m *mockQuerier) UpdateImageWithError(ctx context.Context, arg queries.UpdateImageWithErrorParams) (*queries.Image, error) {
	return nil, nil
}
func (m *mockQuerier) UpdateImageWithStagedURL(ctx context.Context, arg queries.UpdateImageWithStagedURLParams) (*queries.Image, error) {
	return nil, nil
}
func (m *mockQuerier) UpdateJobStatus(ctx context.Context, arg queries.UpdateJobStatusParams) (*queries.Job, error) {
	return nil, nil
}
func (m *mockQuerier) UpdateProject(ctx context.Context, arg queries.UpdateProjectParams) (*queries.UpdateProjectRow, error) {
	return nil, nil
}
func (m *mockQuerier) UpdateProjectByUserID(ctx context.Context, arg queries.UpdateProjectByUserIDParams) (*queries.UpdateProjectByUserIDRow, error) {
	return nil, nil
}

func TestNewUserRepository(t *testing.T) {
	repo := NewUserRepository(nil)
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
				mock.createUserFunc = func(ctx context.Context, arg queries.CreateUserParams) (*queries.User, error) {
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
				mock.createUserFunc = func(ctx context.Context, arg queries.CreateUserParams) (*queries.User, error) {
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
				mock.getUserByIDFunc = func(ctx context.Context, id pgtype.UUID) (*queries.User, error) {
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
				mock.getUserByIDFunc = func(ctx context.Context, id pgtype.UUID) (*queries.User, error) {
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
				mock.getUserByIDFunc = func(ctx context.Context, id pgtype.UUID) (*queries.User, error) {
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
				mock.getUserByAuth0SubFunc = func(ctx context.Context, auth0Sub string) (*queries.User, error) {
					return &queries.User{Auth0Sub: auth0Sub}, nil
				}
			},
		},
		{
			name:     "fail: not found",
			auth0Sub: "auth0|123",
			setupMock: func(mock *mockQuerier) {
				mock.getUserByAuth0SubFunc = func(ctx context.Context, auth0Sub string) (*queries.User, error) {
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
				mock.getUserByAuth0SubFunc = func(ctx context.Context, auth0Sub string) (*queries.User, error) {
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
				mock.getUserByStripeCustomerIDFunc = func(ctx context.Context, stripeCustomerID pgtype.Text) (*queries.User, error) {
					return &queries.User{StripeCustomerID: stripeCustomerID}, nil
				}
			},
		},
		{
			name:             "fail: not found",
			stripeCustomerID: "cus_123",
			setupMock: func(mock *mockQuerier) {
				mock.getUserByStripeCustomerIDFunc = func(ctx context.Context, stripeCustomerID pgtype.Text) (*queries.User, error) {
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
				mock.getUserByStripeCustomerIDFunc = func(ctx context.Context, stripeCustomerID pgtype.Text) (*queries.User, error) {
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
				mock.updateUserStripeCustomerIDFunc = func(ctx context.Context, arg queries.UpdateUserStripeCustomerIDParams) (*queries.User, error) {
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
				mock.updateUserStripeCustomerIDFunc = func(ctx context.Context, arg queries.UpdateUserStripeCustomerIDParams) (*queries.User, error) {
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
				mock.updateUserStripeCustomerIDFunc = func(ctx context.Context, arg queries.UpdateUserStripeCustomerIDParams) (*queries.User, error) {
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
				mock.updateUserRoleFunc = func(ctx context.Context, arg queries.UpdateUserRoleParams) (*queries.User, error) {
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
				mock.updateUserRoleFunc = func(ctx context.Context, arg queries.UpdateUserRoleParams) (*queries.User, error) {
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
				mock.updateUserRoleFunc = func(ctx context.Context, arg queries.UpdateUserRoleParams) (*queries.User, error) {
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
				mock.deleteUserFunc = func(ctx context.Context, id pgtype.UUID) error {
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
				mock.deleteUserFunc = func(ctx context.Context, id pgtype.UUID) error {
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
				mock.listUsersFunc = func(ctx context.Context, arg queries.ListUsersParams) ([]*queries.User, error) {
					return []*queries.User{}, nil
				}
			},
		},
		{
			name:   "fail: db error",
			limit:  10,
			offset: 0,
			setupMock: func(mock *mockQuerier) {
				mock.listUsersFunc = func(ctx context.Context, arg queries.ListUsersParams) ([]*queries.User, error) {
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
				mock.countUsersFunc = func(ctx context.Context) (int64, error) {
					return 5, nil
				}
			},
		},
		{
			name: "fail: db error",
			setupMock: func(mock *mockQuerier) {
				mock.countUsersFunc = func(ctx context.Context) (int64, error) {
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

package user

import (
	"context"

	"github.com/real-staging-ai/api/internal/storage/queries"
)

//go:generate go run github.com/matryer/moq@v0.5.3 -out repository_mock.go . Repository

// Repository is an interface for user repository operations.
type Repository interface {
	Create(ctx context.Context, auth0Sub, stripeCustomerID, role string) (*queries.CreateUserRow, error)
	GetByID(ctx context.Context, userID string) (*queries.GetUserByIDRow, error)
	GetByAuth0Sub(ctx context.Context, auth0Sub string) (*queries.GetUserByAuth0SubRow, error)
	GetByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*queries.GetUserByStripeCustomerIDRow, error)
	UpdateStripeCustomerID(
		ctx context.Context, userID, stripeCustomerID string,
	) (*queries.UpdateUserStripeCustomerIDRow, error)
	UpdateRole(ctx context.Context, userID, role string) (*queries.UpdateUserRoleRow, error)
	Delete(ctx context.Context, userID string) error
	List(ctx context.Context, limit, offset int) ([]*queries.ListUsersRow, error)
	Count(ctx context.Context) (int64, error)

	// Profile operations
	GetProfileByID(ctx context.Context, userID string) (*queries.GetUserProfileByIDRow, error)
	GetProfileByAuth0Sub(ctx context.Context, auth0Sub string) (*queries.GetUserProfileByAuth0SubRow, error)
	UpdateProfile(ctx context.Context, userID string, profile *ProfileUpdate) (*queries.UpdateUserProfileRow, error)
}

// ProfileUpdate represents the fields that can be updated in a user profile.
type ProfileUpdate struct {
	Email           *string
	FullName        *string
	CompanyName     *string
	Phone           *string
	BillingAddress  []byte // JSON
	ProfilePhotoURL *string
	Preferences     []byte // JSON
}

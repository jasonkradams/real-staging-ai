package user

import (
	"context"

	"github.com/virtual-staging-ai/api/internal/storage/queries"
)

// Repository is an interface for user repository operations.
type Repository interface {
	Create(ctx context.Context, auth0Sub, stripeCustomerID, role string) (*queries.User, error)
	GetByID(ctx context.Context, userID string) (*queries.User, error)
	GetByAuth0Sub(ctx context.Context, auth0Sub string) (*queries.User, error)
	GetByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*queries.User, error)
	UpdateStripeCustomerID(ctx context.Context, userID, stripeCustomerID string) (*queries.User, error)
	UpdateRole(ctx context.Context, userID, role string) (*queries.User, error)
	Delete(ctx context.Context, userID string) error
	List(ctx context.Context, limit, offset int) ([]*queries.User, error)
	Count(ctx context.Context) (int64, error)
}

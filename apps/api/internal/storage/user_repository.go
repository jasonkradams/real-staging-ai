package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/virtual-staging-ai/api/internal/storage/queries"
)

// UserRepositoryImpl handles the database operations for users using sqlc-generated queries.
type UserRepositoryImpl struct {
	queries *queries.Queries
}

// Ensure UserRepositoryImpl implements UserRepository interface.
var _ UserRepository = (*UserRepositoryImpl)(nil)

// NewUserRepository creates a new UserRepositoryImpl instance.
func NewUserRepository(db *DB) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		queries: queries.New(db.pool),
	}
}

// CreateUser creates a new user in the database.
func (r *UserRepositoryImpl) CreateUser(ctx context.Context, auth0Sub, stripeCustomerID, role string) (*queries.User, error) {
	var stripeCustomerIDType pgtype.Text
	if stripeCustomerID != "" {
		stripeCustomerIDType = pgtype.Text{String: stripeCustomerID, Valid: true}
	}

	params := queries.CreateUserParams{
		Auth0Sub:         auth0Sub,
		StripeCustomerID: stripeCustomerIDType,
		Role:             role,
	}

	user, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("unable to create user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by their ID.
func (r *UserRepositoryImpl) GetUserByID(ctx context.Context, userID string) (*queries.User, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	userUUIDType := pgtype.UUID{}
	err = userUUIDType.Scan(userUUID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to convert user ID to pgtype.UUID: %w", err)
	}

	user, err := r.queries.GetUserByID(ctx, userUUIDType)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("unable to get user by ID: %w", err)
	}

	return user, nil
}

// GetUserByAuth0Sub retrieves a user by their Auth0 subject ID.
func (r *UserRepositoryImpl) GetUserByAuth0Sub(ctx context.Context, auth0Sub string) (*queries.User, error) {
	user, err := r.queries.GetUserByAuth0Sub(ctx, auth0Sub)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("unable to get user by Auth0 sub: %w", err)
	}

	return user, nil
}

// GetUserByStripeCustomerID retrieves a user by their Stripe customer ID.
func (r *UserRepositoryImpl) GetUserByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*queries.User, error) {
	stripeCustomerIDType := pgtype.Text{String: stripeCustomerID, Valid: true}

	user, err := r.queries.GetUserByStripeCustomerID(ctx, stripeCustomerIDType)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("unable to get user by Stripe customer ID: %w", err)
	}

	return user, nil
}

// UpdateUserStripeCustomerID updates a user's Stripe customer ID.
func (r *UserRepositoryImpl) UpdateUserStripeCustomerID(ctx context.Context, userID, stripeCustomerID string) (*queries.User, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	userUUIDType := pgtype.UUID{}
	err = userUUIDType.Scan(userUUID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to convert user ID to pgtype.UUID: %w", err)
	}

	stripeCustomerIDType := pgtype.Text{String: stripeCustomerID, Valid: true}

	params := queries.UpdateUserStripeCustomerIDParams{
		ID:               userUUIDType,
		StripeCustomerID: stripeCustomerIDType,
	}

	user, err := r.queries.UpdateUserStripeCustomerID(ctx, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("unable to update user Stripe customer ID: %w", err)
	}

	return user, nil
}

// UpdateUserRole updates a user's role.
func (r *UserRepositoryImpl) UpdateUserRole(ctx context.Context, userID, role string) (*queries.User, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	userUUIDType := pgtype.UUID{}
	err = userUUIDType.Scan(userUUID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to convert user ID to pgtype.UUID: %w", err)
	}

	params := queries.UpdateUserRoleParams{
		ID:   userUUIDType,
		Role: role,
	}

	user, err := r.queries.UpdateUserRole(ctx, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("unable to update user role: %w", err)
	}

	return user, nil
}

// DeleteUser deletes a user from the database.
func (r *UserRepositoryImpl) DeleteUser(ctx context.Context, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	userUUIDType := pgtype.UUID{}
	err = userUUIDType.Scan(userUUID.String())
	if err != nil {
		return fmt.Errorf("failed to convert user ID to pgtype.UUID: %w", err)
	}

	err = r.queries.DeleteUser(ctx, userUUIDType)
	if err != nil {
		return fmt.Errorf("unable to delete user: %w", err)
	}

	return nil
}

// ListUsers retrieves a paginated list of users.
func (r *UserRepositoryImpl) ListUsers(ctx context.Context, limit, offset int) ([]*queries.User, error) {
	params := queries.ListUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	users, err := r.queries.ListUsers(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("unable to list users: %w", err)
	}

	return users, nil
}

// CountUsers returns the total number of users.
func (r *UserRepositoryImpl) CountUsers(ctx context.Context) (int64, error) {
	count, err := r.queries.CountUsers(ctx)
	if err != nil {
		return 0, fmt.Errorf("unable to count users: %w", err)
	}

	return count, nil
}

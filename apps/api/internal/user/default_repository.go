package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/real-staging-ai/api/internal/storage"
	"github.com/real-staging-ai/api/internal/storage/queries"
)

// DefaultRepository handles the database operations for users using sqlc-generated queries.
type DefaultRepository struct {
	queries queries.Querier
}

// Ensure DefaultRepository implements UserRepository interface.
var _ Repository = (*DefaultRepository)(nil)

// NewDefaultRepository creates a new DefaultRepository instance.
func NewDefaultRepository(db storage.Database) *DefaultRepository {
	return &DefaultRepository{
		queries: queries.New(db),
	}
}

// Create creates a new user in the database.
func (r *DefaultRepository) Create(
	ctx context.Context, auth0Sub, stripeCustomerID, role string,
) (*queries.CreateUserRow, error) {
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

// GetByID retrieves a user by their ID.
func (r *DefaultRepository) GetByID(ctx context.Context, userID string) (*queries.GetUserByIDRow, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	userUUIDType := pgtype.UUID{Bytes: userUUID, Valid: true}

	user, err := r.queries.GetUserByID(ctx, userUUIDType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("unable to get user by ID: %w", err)
	}

	return user, nil
}

// GetByAuth0Sub retrieves a user by their Auth0 subject ID.
func (r *DefaultRepository) GetByAuth0Sub(ctx context.Context, auth0Sub string) (*queries.GetUserByAuth0SubRow, error) {
	user, err := r.queries.GetUserByAuth0Sub(ctx, auth0Sub)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("unable to get user by Auth0 sub: %w", err)
	}

	return user, nil
}

// GetByStripeCustomerID retrieves a user by their Stripe customer ID.
func (r *DefaultRepository) GetByStripeCustomerID(
	ctx context.Context, stripeCustomerID string,
) (*queries.GetUserByStripeCustomerIDRow, error) {
	stripeCustomerIDType := pgtype.Text{String: stripeCustomerID, Valid: true}

	user, err := r.queries.GetUserByStripeCustomerID(ctx, stripeCustomerIDType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("unable to get user by Stripe customer ID: %w", err)
	}

	return user, nil
}

// UpdateStripeCustomerID updates a user's Stripe customer ID.
func (r *DefaultRepository) UpdateStripeCustomerID(
	ctx context.Context, userID, stripeCustomerID string,
) (*queries.UpdateUserStripeCustomerIDRow, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	userUUIDType := pgtype.UUID{Bytes: userUUID, Valid: true}

	stripeCustomerIDType := pgtype.Text{String: stripeCustomerID, Valid: true}

	params := queries.UpdateUserStripeCustomerIDParams{
		ID:               userUUIDType,
		StripeCustomerID: stripeCustomerIDType,
	}

	user, err := r.queries.UpdateUserStripeCustomerID(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("unable to update user Stripe customer ID: %w", err)
	}

	return user, nil
}

// UpdateRole updates a user's role.
func (r *DefaultRepository) UpdateRole(ctx context.Context, userID, role string) (*queries.UpdateUserRoleRow, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	userUUIDType := pgtype.UUID{Bytes: userUUID, Valid: true}

	params := queries.UpdateUserRoleParams{
		ID:   userUUIDType,
		Role: role,
	}

	user, err := r.queries.UpdateUserRole(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("unable to update user role: %w", err)
	}

	return user, nil
}

// Delete deletes a user from the database.
func (r *DefaultRepository) Delete(ctx context.Context, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	userUUIDType := pgtype.UUID{Bytes: userUUID, Valid: true}

	err = r.queries.DeleteUser(ctx, userUUIDType)
	if err != nil {
		return fmt.Errorf("unable to delete user: %w", err)
	}

	return nil
}

// List retrieves a paginated list of users.
func (r *DefaultRepository) List(ctx context.Context, limit, offset int) ([]*queries.ListUsersRow, error) {
	params := queries.ListUsersParams{
		Limit:  int32(limit),  // #nosec G115 -- Limit/offset are validated by caller
		Offset: int32(offset), // #nosec G115 -- Limit/offset are validated by caller
	}

	users, err := r.queries.ListUsers(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("unable to list users: %w", err)
	}

	return users, nil
}

// Count returns the total number of users.
func (r *DefaultRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.queries.CountUsers(ctx)
	if err != nil {
		return 0, fmt.Errorf("unable to count users: %w", err)
	}

	return count, nil
}

// GetProfileByID retrieves a full user profile by user ID.
func (r *DefaultRepository) GetProfileByID(ctx context.Context, userID string) (*queries.GetUserProfileByIDRow, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	userUUIDType := pgtype.UUID{Bytes: userUUID, Valid: true}

	profile, err := r.queries.GetUserProfileByID(ctx, userUUIDType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("unable to get user profile by ID: %w", err)
	}

	return profile, nil
}

// GetProfileByAuth0Sub retrieves a full user profile by Auth0 subject ID.
func (r *DefaultRepository) GetProfileByAuth0Sub(
	ctx context.Context, auth0Sub string,
) (*queries.GetUserProfileByAuth0SubRow, error) {
	profile, err := r.queries.GetUserProfileByAuth0Sub(ctx, auth0Sub)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("unable to get user profile by Auth0 sub: %w", err)
	}

	return profile, nil
}

// UpdateProfile updates a user's profile information.
func (r *DefaultRepository) UpdateProfile(
	ctx context.Context, userID string, profile *ProfileUpdate,
) (*queries.UpdateUserProfileRow, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	userUUIDType := pgtype.UUID{Bytes: userUUID, Valid: true}

	// Convert optional fields to pgtype
	var emailType, fullNameType, companyNameType, phoneType, profilePhotoURLType pgtype.Text
	var billingAddressType, preferencesType []byte

	if profile.Email != nil {
		emailType = pgtype.Text{String: *profile.Email, Valid: true}
	}
	if profile.FullName != nil {
		fullNameType = pgtype.Text{String: *profile.FullName, Valid: true}
	}
	if profile.CompanyName != nil {
		companyNameType = pgtype.Text{String: *profile.CompanyName, Valid: true}
	}
	if profile.Phone != nil {
		phoneType = pgtype.Text{String: *profile.Phone, Valid: true}
	}
	if profile.ProfilePhotoURL != nil {
		profilePhotoURLType = pgtype.Text{String: *profile.ProfilePhotoURL, Valid: true}
	}
	if profile.BillingAddress != nil {
		billingAddressType = profile.BillingAddress
	}
	if profile.Preferences != nil {
		preferencesType = profile.Preferences
	}

	params := queries.UpdateUserProfileParams{
		ID:              userUUIDType,
		Email:           emailType,
		FullName:        fullNameType,
		CompanyName:     companyNameType,
		Phone:           phoneType,
		BillingAddress:  billingAddressType,
		ProfilePhotoUrl: profilePhotoURLType,
		Preferences:     preferencesType,
	}

	updated, err := r.queries.UpdateUserProfile(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("unable to update user profile: %w", err)
	}

	return updated, nil
}

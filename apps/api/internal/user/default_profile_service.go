package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/real-staging-ai/api/internal/storage/queries"
)

// DefaultProfileService implements the ProfileService interface.
type DefaultProfileService struct {
	repo Repository
}

// Ensure DefaultProfileService implements ProfileService.
var _ ProfileService = (*DefaultProfileService)(nil)

// NewDefaultProfileService creates a new DefaultProfileService instance.
func NewDefaultProfileService(repo Repository) *DefaultProfileService {
	return &DefaultProfileService{
		repo: repo,
	}
}

// GetProfile retrieves a user's complete profile by user ID.
func (s *DefaultProfileService) GetProfile(ctx context.Context, userID string) (*ProfileResponse, error) {
	profile, err := s.repo.GetProfileByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return convertProfileRowToResponse(profile), nil
}

// GetProfileByAuth0Sub retrieves a user's complete profile by Auth0 subject.
func (s *DefaultProfileService) GetProfileByAuth0Sub(ctx context.Context, auth0Sub string) (*ProfileResponse, error) {
	profile, err := s.repo.GetProfileByAuth0Sub(ctx, auth0Sub)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return convertProfileRowToResponse(profile), nil
}

// UpdateProfile updates a user's profile information.
func (s *DefaultProfileService) UpdateProfile(
	ctx context.Context, userID string, req *ProfileUpdateRequest,
) (*ProfileResponse, error) {
	// Validate request
	if err := validateProfileUpdate(req); err != nil {
		return nil, fmt.Errorf("invalid profile update request: %w", err)
	}

	// Convert request to repository ProfileUpdate
	profileUpdate := &ProfileUpdate{
		Email:           req.Email,
		FullName:        req.FullName,
		CompanyName:     req.CompanyName,
		Phone:           req.Phone,
		ProfilePhotoURL: req.ProfilePhotoURL,
	}

	// Handle JSON fields
	if len(req.BillingAddress) > 0 {
		profileUpdate.BillingAddress = req.BillingAddress
	}
	if len(req.Preferences) > 0 {
		profileUpdate.Preferences = req.Preferences
	}

	// Update profile
	updated, err := s.repo.UpdateProfile(ctx, userID, profileUpdate)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	return convertUpdateProfileRowToResponse(updated), nil
}

// validateProfileUpdate validates a profile update request.
func validateProfileUpdate(req *ProfileUpdateRequest) error {
	// Email validation if provided
	if req.Email != nil && *req.Email != "" {
		// Basic email format check
		if len(*req.Email) < 3 || len(*req.Email) > 255 {
			return fmt.Errorf("email must be between 3 and 255 characters")
		}
	}

	// Phone validation if provided
	if req.Phone != nil && *req.Phone != "" {
		if len(*req.Phone) > 20 {
			return fmt.Errorf("phone number must be less than 20 characters")
		}
	}

	// Name validations
	if req.FullName != nil && len(*req.FullName) > 100 {
		return fmt.Errorf("full name must be less than 100 characters")
	}
	if req.CompanyName != nil && len(*req.CompanyName) > 100 {
		return fmt.Errorf("company name must be less than 100 characters")
	}

	return nil
}

// convertProfileRowToResponse converts a profile query result to ProfileResponse.
func convertProfileRowToResponse(profile interface{}) *ProfileResponse {
	// Handle both GetUserProfileByIDRow and GetUserProfileByAuth0SubRow
	switch p := profile.(type) {
	case *queries.GetUserProfileByIDRow:
		return mapRowToProfileResponse(
			p.ID, p.StripeCustomerID, p.Role, p.Email, p.FullName,
			p.CompanyName, p.Phone, p.BillingAddress, p.ProfilePhotoUrl,
			p.Preferences, p.CreatedAt, p.UpdatedAt,
		)
	case *queries.GetUserProfileByAuth0SubRow:
		return mapRowToProfileResponse(
			p.ID, p.StripeCustomerID, p.Role, p.Email, p.FullName,
			p.CompanyName, p.Phone, p.BillingAddress, p.ProfilePhotoUrl,
			p.Preferences, p.CreatedAt, p.UpdatedAt,
		)
	default:
		return nil
	}
}

// convertUpdateProfileRowToResponse converts UpdateUserProfileRow to ProfileResponse.
func convertUpdateProfileRowToResponse(profile *queries.UpdateUserProfileRow) *ProfileResponse {
	return mapRowToProfileResponse(
		profile.ID, profile.StripeCustomerID, profile.Role, profile.Email, profile.FullName,
		profile.CompanyName, profile.Phone, profile.BillingAddress, profile.ProfilePhotoUrl,
		profile.Preferences, profile.CreatedAt, profile.UpdatedAt,
	)
}

// mapRowToProfileResponse maps database row fields to ProfileResponse.
func mapRowToProfileResponse(
	id pgtype.UUID,
	stripeCustomerID pgtype.Text,
	role string,
	email, fullName, companyName, phone pgtype.Text,
	billingAddress []byte,
	profilePhotoUrl pgtype.Text,
	preferences []byte,
	createdAt, updatedAt pgtype.Timestamptz,
) *ProfileResponse {
	// Convert UUID to string
	uuidBytes := id.Bytes
	idUUID := uuid.UUID(uuidBytes)

	resp := &ProfileResponse{
		ID:   idUUID.String(),
		Role: role,
	}

	// Handle optional text fields
	if stripeCustomerID.Valid && stripeCustomerID.String != "" {
		resp.StripeCustomerID = &stripeCustomerID.String
	}
	if email.Valid {
		resp.Email = &email.String
	}
	if fullName.Valid {
		resp.FullName = &fullName.String
	}
	if companyName.Valid {
		resp.CompanyName = &companyName.String
	}
	if phone.Valid {
		resp.Phone = &phone.String
	}
	if profilePhotoUrl.Valid {
		resp.ProfilePhotoURL = &profilePhotoUrl.String
	}

	// Handle JSON fields
	if len(billingAddress) > 0 {
		resp.BillingAddress = billingAddress
	}
	if len(preferences) > 0 {
		resp.Preferences = preferences
	}

	// Handle timestamps
	if createdAt.Valid {
		resp.CreatedAt = createdAt.Time.String()
	}
	if updatedAt.Valid {
		resp.UpdatedAt = updatedAt.Time.String()
	}

	return resp
}

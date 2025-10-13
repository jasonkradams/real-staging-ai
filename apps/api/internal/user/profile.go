package user

import (
	"context"
	"encoding/json"
)

//go:generate go run github.com/matryer/moq@v0.5.3 -out profile_mock.go . ProfileService

// ProfileService defines the service-layer contract for user profile operations.
type ProfileService interface {
	// GetProfile retrieves a user's complete profile by user ID.
	GetProfile(ctx context.Context, userID string) (*ProfileResponse, error)

	// GetProfileByAuth0Sub retrieves a user's complete profile by Auth0 subject.
	GetProfileByAuth0Sub(ctx context.Context, auth0Sub string) (*ProfileResponse, error)

	// UpdateProfile updates a user's profile information.
	UpdateProfile(ctx context.Context, userID string, req *ProfileUpdateRequest) (*ProfileResponse, error)
}

// ProfileResponse represents a user's complete profile information.
type ProfileResponse struct {
	ID               string          `json:"id"`
	Email            *string         `json:"email,omitempty"`
	FullName         *string         `json:"full_name,omitempty"`
	CompanyName      *string         `json:"company_name,omitempty"`
	Phone            *string         `json:"phone,omitempty"`
	BillingAddress   json.RawMessage `json:"billing_address,omitempty"`
	ProfilePhotoURL  *string         `json:"profile_photo_url,omitempty"`
	Preferences      json.RawMessage `json:"preferences,omitempty"`
	Role             string          `json:"role"`
	StripeCustomerID *string         `json:"stripe_customer_id,omitempty"`
	CreatedAt        string          `json:"created_at"`
	UpdatedAt        string          `json:"updated_at"`
}

// ProfileUpdateRequest represents the request body for updating a user profile.
type ProfileUpdateRequest struct {
	Email           *string         `json:"email,omitempty"`
	FullName        *string         `json:"full_name,omitempty"`
	CompanyName     *string         `json:"company_name,omitempty"`
	Phone           *string         `json:"phone,omitempty"`
	BillingAddress  json.RawMessage `json:"billing_address,omitempty"`
	ProfilePhotoURL *string         `json:"profile_photo_url,omitempty"`
	Preferences     json.RawMessage `json:"preferences,omitempty"`
}

// BillingAddress represents a user's billing address.
type BillingAddress struct {
	Line1      string `json:"line1,omitempty"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city,omitempty"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	Country    string `json:"country,omitempty"`
}

// Preferences represents user preferences.
type Preferences struct {
	EmailNotifications bool   `json:"email_notifications"`
	MarketingEmails    bool   `json:"marketing_emails"`
	DefaultRoomType    string `json:"default_room_type,omitempty"`
	DefaultStyle       string `json:"default_style,omitempty"`
}

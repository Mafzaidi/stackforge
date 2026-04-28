package serializer

import (
	"time"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

// UserProfileResponse is the API representation of a user profile.
type UserProfileResponse struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id"`
	EncryptionKeyHash  *string   `json:"encryption_key_hash,omitempty"`
	MasterPasswordHash *string   `json:"master_password_hash,omitempty"`
	PasswordSalt       *string   `json:"password_salt,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// UserProfileCreatedResponse is a minimal response after profile creation.
type UserProfileCreatedResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FromUserProfile converts a domain user profile entity to a full API response.
func FromUserProfile(e *entity.UserProfiles) UserProfileResponse {
	return UserProfileResponse{
		ID:        e.ID,
		UserID:    e.UserID,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

// FromUserProfileCreated converts a domain entity to a minimal creation response.
func FromUserProfileCreated(e *entity.UserProfiles) UserProfileCreatedResponse {
	return UserProfileCreatedResponse{
		ID:        e.ID,
		UserID:    e.UserID,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

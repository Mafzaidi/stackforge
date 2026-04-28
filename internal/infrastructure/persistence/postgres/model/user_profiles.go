package model

import (
	"time"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

// UserProfiles is the GORM model for the stackforge_service.user_profiles table.
type UserProfiles struct {
	ID                 string    `gorm:"column:id;primaryKey"`
	UserID             string    `gorm:"column:user_id"`
	EncryptionKeyHash  *string   `gorm:"column:encryption_key_hash"`
	MasterPasswordHash *string   `gorm:"column:master_password_hash"`
	PasswordSalt       *string   `gorm:"column:password_salt"`
	CreatedAt          time.Time `gorm:"column:created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at"`
}

// TableName returns the fully qualified table name with schema prefix.
func (UserProfiles) TableName() string {
	return "stackforge_service.user_profiles"
}

// UserProfilesFromEntity converts a domain entity to a GORM model.
func UserProfilesFromEntity(e *entity.UserProfiles) *UserProfiles {
	return &UserProfiles{
		ID:                 e.ID,
		UserID:             e.UserID,
		EncryptionKeyHash:  e.EncryptionKeyHash,
		MasterPasswordHash: e.MasterPasswordHash,
		PasswordSalt:       e.PasswordSalt,
		CreatedAt:          e.CreatedAt,
		UpdatedAt:          e.UpdatedAt,
	}
}

// ToEntity converts the GORM model back to a domain entity.
func (m *UserProfiles) ToEntity() *entity.UserProfiles {
	return &entity.UserProfiles{
		ID:                 m.ID,
		UserID:             m.UserID,
		EncryptionKeyHash:  m.EncryptionKeyHash,
		MasterPasswordHash: m.MasterPasswordHash,
		PasswordSalt:       m.PasswordSalt,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
}

package entity

import "time"

type UserProfiles struct {
	ID                 string
	UserID             string
	EncryptionKeyHash  *string
	MasterPasswordHash *string
	PasswordSalt       *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

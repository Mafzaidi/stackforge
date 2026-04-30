package entity

import "time"

// Credential represents a credential item in the domain.
type Credential struct {
	ID                string
	UserID            string
	VaultID           string
	CategoryID        string
	Title             string
	SiteUrl           string
	FaviconUrl        *string
	UsernameEncrypted string
	PasswordEncrypted string
	NotesEncrypted    *string
	IsFavorite        bool
	PasswordStrength  int64
	LastUsedAt        *time.Time
	PasswordChangedAt *time.Time
	ExpiresAt         *time.Time
	AutoLogin         bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Vault             *Vault
	Category          *MasterData
	Tags              []*Tag
}

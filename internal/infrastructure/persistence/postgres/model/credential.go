package model

import (
	"time"

	"github.com/mafzaidi/stackforge/internal/domain/entity"
)

// Credential is the GORM model for the stackforge_service.credentials table.
type Credential struct {
	ID                string     `gorm:"column:id;primaryKey"`
	UserID            string     `gorm:"column:user_id"`
	VaultID           string     `gorm:"column:vault_id"`
	CategoryID        string     `gorm:"column:category_id"`
	Title             string     `gorm:"column:title"`
	SiteUrl           string     `gorm:"column:site_url"`
	FaviconUrl        *string    `gorm:"column:favicon_url"`
	UsernameEncrypted string     `gorm:"column:username_encrypted"`
	PasswordEncrypted string     `gorm:"column:password_encrypted"`
	NotesEncrypted    *string    `gorm:"column:notes_encrypted"`
	IsFavorite        bool       `gorm:"column:is_favorite"`
	PasswordStrength  int64      `gorm:"column:password_strength"`
	LastUsedAt        *time.Time `gorm:"column:last_used_at"`
	PasswordChangedAt *time.Time `gorm:"column:password_changed_at"`
	ExpiresAt         *time.Time `gorm:"column:expires_at"`
	AutoLogin         bool       `gorm:"column:auto_login"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at"`
}

// TableName returns the fully qualified table name with schema prefix.
func (Credential) TableName() string {
	return "stackforge_service.credentials"
}

// CredentialFromEntity converts a domain entity to a GORM model.
func CredentialFromEntity(e *entity.Credential) *Credential {
	m := &Credential{
		ID:                e.ID,
		UserID:            e.UserID,
		Title:             e.Title,
		SiteUrl:           e.SiteUrl,
		FaviconUrl:        e.FaviconUrl,
		UsernameEncrypted: e.UsernameEncrypted,
		PasswordEncrypted: e.PasswordEncrypted,
		IsFavorite:        e.IsFavorite,
		PasswordStrength:  e.PasswordStrength,
		AutoLogin:         e.AutoLogin,
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
	}

	if e.VaultID != "" {
		m.VaultID = e.VaultID
	}
	if e.CategoryID != "" {
		m.CategoryID = e.CategoryID
	}
	if e.NotesEncrypted != nil {
		m.NotesEncrypted = e.NotesEncrypted
	}
	if e.LastUsedAt != nil && !e.LastUsedAt.IsZero() {
		m.LastUsedAt = e.LastUsedAt
	}
	if e.PasswordChangedAt != nil && !e.PasswordChangedAt.IsZero() {
		m.PasswordChangedAt = e.PasswordChangedAt
	}
	if e.ExpiresAt != nil && !e.ExpiresAt.IsZero() {
		m.ExpiresAt = e.ExpiresAt
	}
	return m
}

// ToEntity converts the GORM model back to a domain entity.
func (m *Credential) ToEntity() *entity.Credential {
	e := &entity.Credential{
		ID:                m.ID,
		UserID:            m.UserID,
		Title:             m.Title,
		SiteUrl:           m.SiteUrl,
		FaviconUrl:        m.FaviconUrl,
		UsernameEncrypted: m.UsernameEncrypted,
		PasswordEncrypted: m.PasswordEncrypted,
		IsFavorite:        m.IsFavorite,
		AutoLogin:         m.AutoLogin,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
	if m.VaultID != "" {
		e.VaultID = m.VaultID
	}
	if m.CategoryID != "" {
		e.CategoryID = m.CategoryID
	}
	if m.NotesEncrypted != nil {
		e.NotesEncrypted = m.NotesEncrypted
	}
	if m.PasswordStrength != 0 {
		e.PasswordStrength = m.PasswordStrength
	}
	if m.LastUsedAt != nil {
		e.LastUsedAt = m.LastUsedAt
	}
	if m.PasswordChangedAt != nil {
		e.PasswordChangedAt = m.PasswordChangedAt
	}
	if m.ExpiresAt != nil {
		e.ExpiresAt = m.ExpiresAt
	}
	return e
}

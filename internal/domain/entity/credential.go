package entity

import "time"

// Credential represents a credential item in the domain.
type Credential struct {
	ID        string
	Name      string
	Username  string
	Secret    string // encrypted in real impl
	CreatedAt time.Time
	UpdatedAt time.Time
}

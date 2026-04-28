package entity

import "time"

type Vault struct {
	ID          string
	UserID      string
	Name        string
	Description string
	Icon        string
	IsDefault   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

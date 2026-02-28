package entity

import "time"

// Todo represents a single todo item in the domain.
type Todo struct {
	ID          string
	Title       string
	Description string
	Done        bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

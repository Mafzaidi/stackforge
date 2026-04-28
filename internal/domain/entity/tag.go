package entity

import "time"

type Tag struct {
	ID        string
	UserID    string
	Name      string
	Color     *string
	Module    string
	RefID     string
	CreatedAt time.Time
	IsActive  bool
}

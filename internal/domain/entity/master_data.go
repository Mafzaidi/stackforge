package entity

import "time"

type MasterData struct {
	ID          string
	Module      string
	Type        string
	Name        string
	Description string
	Icon        string
	Color       string
	SortOrder   string
	IsActive    bool
	Metadata    map[string]interface{}
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

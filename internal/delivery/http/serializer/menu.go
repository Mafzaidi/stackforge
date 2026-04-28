package serializer

import "github.com/mafzaidi/stackforge/internal/domain/entity"

// MenuItemResponse is the API representation of a menu item.
type MenuItemResponse struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Icon      string `json:"icon"`
	SortOrder string `json:"sort_order"`
}

// FromMenuItem converts a domain menu item entity to an API response.
func FromMenuItem(e entity.MenuItem) MenuItemResponse {
	return MenuItemResponse{
		ID:        e.ID,
		Title:     e.Title,
		URL:       e.URL,
		Icon:      e.Icon,
		SortOrder: e.SortOrder,
	}
}

// FromMenuItemList converts a slice of menu item entities.
func FromMenuItemList(entities []entity.MenuItem) []MenuItemResponse {
	result := make([]MenuItemResponse, len(entities))
	for i, e := range entities {
		result[i] = FromMenuItem(e)
	}
	return result
}

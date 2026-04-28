package entity

// MenuItem represents a navigation menu item returned to the frontend.
// Data is fetched from MasterData and mapped to this response struct.
type MenuItem struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Icon      string `json:"icon"`
	SortOrder string `json:"sort_order"`
}

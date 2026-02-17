package list

// --- Query & Result ---

type Query struct {
	UserID string
	Limit  int
	Offset int
}

type Result struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Slug      string                 `json:"slug"`
	Settings  map[string]interface{} `json:"settings"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

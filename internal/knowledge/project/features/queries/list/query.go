package list

type Query struct {
	WorkspaceID string
	Limit       int
	Offset      int
}

type Result struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

package listtypes

type Query struct {
	WorkspaceID string
	Limit       int
	Offset      int
}

type Result struct {
	ID            string `json:"id"`
	WorkspaceID   string `json:"workspace_id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	IsDirectional bool   `json:"is_directional"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

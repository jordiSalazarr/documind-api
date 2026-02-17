package get

type Query struct {
	ID string
}

type Result struct {
	ID          string                 `json:"id"`
	WorkspaceID string                 `json:"workspace_id"`
	Name        string                 `json:"name"`
	Slug        string                 `json:"slug"`
	Description *string                `json:"description"`
	Icon        *string                `json:"icon"`
	Fields      map[string]interface{} `json:"fields"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

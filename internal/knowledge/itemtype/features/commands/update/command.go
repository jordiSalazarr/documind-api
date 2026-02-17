package update

type Command struct {
	ID          string
	Name        string
	Description *string
	Icon        *string
	Fields      map[string]interface{}
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

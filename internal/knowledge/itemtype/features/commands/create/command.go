package create

import (
	"strings"
)

type Command struct {
	WorkspaceID string
	Name        string
	Slug        string
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

func generateSlug(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}

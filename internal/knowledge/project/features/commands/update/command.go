package update

import "errors"

var ErrProjectNotFound = errors.New("project not found")

type Command struct {
	ID          string
	Name        string
	Description string
	UpdatedBy   string
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

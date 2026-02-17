package get

import "errors"

var ErrWorkspaceNotFound = errors.New("workspace not found")

// --- Query & Result ---

type Query struct {
	ID string
}

type Result struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Slug      string                 `json:"slug"`
	Settings  map[string]interface{} `json:"settings"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

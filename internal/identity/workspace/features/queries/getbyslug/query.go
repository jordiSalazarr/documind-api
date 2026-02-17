package getbyslug

import "errors"

var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
	ErrInvalidSlug       = errors.New("invalid slug")
)

// --- Query & Result ---

type Query struct {
	Slug string
}

type Result struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Slug      string                 `json:"slug"`
	Settings  map[string]interface{} `json:"settings"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

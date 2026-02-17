package update

import "errors"

var ErrWorkspaceNotFound = errors.New("workspace not found")

// --- Command & Result ---

type Command struct {
	ID       string
	Name     string
	Settings map[string]interface{}
}

type Result struct {
	ID        string
	Name      string
	Slug      string
	Settings  map[string]interface{}
	CreatedAt string
	UpdatedAt string
}

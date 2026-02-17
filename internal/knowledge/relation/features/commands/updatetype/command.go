package updatetype

import (
	"errors"
)

var (
	ErrRelationTypeNotFound = errors.New("relation type not found")
	ErrRelationTypeExists   = errors.New("relation type with this slug already exists")
	ErrInvalidSlug          = errors.New("invalid slug")
)

type Command struct {
	ID            string
	Name          *string
	Slug          *string
	IsDirectional *bool
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

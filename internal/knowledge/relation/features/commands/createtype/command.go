package createtype

import (
	"errors"
)

var (
	ErrInvalidRelationTypeName = errors.New("invalid relation type name")
	ErrInvalidSlug             = errors.New("invalid slug")
	ErrRelationTypeExists      = errors.New("relation type with this slug already exists")
)

type Command struct {
	WorkspaceID   string
	Name          string
	Slug          string
	IsDirectional bool
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

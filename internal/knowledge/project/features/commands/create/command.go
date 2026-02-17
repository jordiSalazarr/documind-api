package create

import (
	"errors"
	"strings"
)

var (
	ErrInvalidProjectName = errors.New("invalid project name")
	ErrInvalidSlug        = errors.New("invalid slug")
	ErrProjectExists      = errors.New("project with this slug already exists in workspace")
)

type Command struct {
	WorkspaceID string
	Name        string
	Slug        string
	Description string
	CreatedBy   string
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

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug)
	return slug
}

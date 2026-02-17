package create

import (
	"errors"
	"strings"
)

var (
	ErrInvalidWorkspaceName = errors.New("invalid workspace name")
	ErrInvalidSlug          = errors.New("invalid slug")
	ErrWorkspaceExists      = errors.New("workspace with this slug already exists")
)

// --- Command & Result ---

type Command struct {
	Name      string
	Slug      string
	UserID    string
	UserEmail string
}

type Result struct {
	ID        string
	Name      string
	Slug      string
	Settings  map[string]interface{}
	CreatedAt string
	UpdatedAt string
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

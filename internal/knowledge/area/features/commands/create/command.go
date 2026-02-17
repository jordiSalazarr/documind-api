package create

import (
	"errors"
	"strings"
)

var (
	ErrInvalidAreaName = errors.New("invalid area name")
	ErrInvalidSlug     = errors.New("invalid slug")
	ErrAreaExists      = errors.New("area with this slug already exists in project")
	ErrProjectNotFound = errors.New("project not found")
)

type Command struct {
	ProjectID   string
	Name        string
	Slug        string
	Description string
}

type Result struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
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

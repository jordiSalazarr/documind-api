package get

import "errors"

var ErrAreaNotFound = errors.New("area not found")

type Query struct {
	ID string
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

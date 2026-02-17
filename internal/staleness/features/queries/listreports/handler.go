package listreports

import (
	"context"
	"database/sql"
)

// Handler handles listing staleness reports.
type Handler struct {
	repo Repository
}

// NewHandler creates a new list reports handler.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

// Handle lists staleness reports with optional filtering.
func (h *Handler) Handle(ctx context.Context, q Query) ([]*Result, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
	if q.Limit > 100 {
		q.Limit = 100
	}
	return h.repo.List(ctx, q)
}

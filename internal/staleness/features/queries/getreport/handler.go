package getreport

import (
	"context"
	"database/sql"
)

// Handler handles getting a staleness report by ID.
type Handler struct {
	repo Repository
}

// NewHandler creates a new get report handler.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

// Handle retrieves a staleness report with all its items.
func (h *Handler) Handle(ctx context.Context, q Query) (*Result, error) {
	return h.repo.GetByID(ctx, q.ID, q.WorkspaceID)
}

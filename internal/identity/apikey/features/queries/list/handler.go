package list

import (
	"context"
	"database/sql"
)

// Handler handles listing API keys.
type Handler struct {
	repo Repository
}

// NewHandler creates a new list API keys handler.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

// Handle lists all non-revoked API keys for a workspace.
func (h *Handler) Handle(ctx context.Context, q Query) ([]*Result, error) {
	return h.repo.ListByWorkspace(ctx, q.WorkspaceID)
}

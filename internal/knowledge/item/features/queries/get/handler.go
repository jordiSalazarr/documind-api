package get

import (
	"context"
	"database/sql"
)

// Handler retrieves an item and its latest version.
type Handler struct {
	repo Repository
}

// NewHandler constructs a Handler.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

// Handle executes the get-item query.
func (h *Handler) Handle(ctx context.Context, q Query) (*Result, error) {
	item, err := h.repo.GetByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	version, err := h.repo.GetLatestVersion(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	return &Result{Item: item, Version: version}, nil
}

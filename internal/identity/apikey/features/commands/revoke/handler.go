package revoke

import (
	"context"
	"database/sql"
)

// Handler handles API key revocation.
type Handler struct {
	repo Repository
}

// NewHandler creates a new revoke API key handler.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

// Handle revokes an API key.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	found, err := h.repo.Revoke(ctx, cmd.ID, cmd.WorkspaceID)
	if err != nil {
		return err
	}
	if !found {
		return ErrKeyNotFound
	}
	return nil
}

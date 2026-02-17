package delete

import (
	"context"
	"database/sql"
)

// Handler orchestrates item deletion.
type Handler struct {
	repo Repository
}

// NewHandler constructs a Handler.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

// Handle executes the delete-item command.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	return h.repo.Delete(ctx, cmd.ID)
}

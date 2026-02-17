package delete

import (
	"context"
	"database/sql"
	"errors"
)

type Handler struct{ repo Repository }

func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	exists, err := h.repo.Exists(ctx, cmd.ID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("item type not found")
	}
	return h.repo.Delete(ctx, cmd.ID)
}

package delete

import (
	"context"
	"database/sql"

	shared "documind.jordi.org/internal/shared/domain"
)

type Handler struct{ repo Repository }

func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	id := shared.ID(cmd.ID)
	ws, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if ws == nil {
		return ErrWorkspaceNotFound
	}
	return h.repo.Delete(ctx, ws.ID)
}

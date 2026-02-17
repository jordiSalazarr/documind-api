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
	project, err := h.repo.GetByID(ctx, shared.ID(cmd.ID))
	if err != nil {
		return err
	}
	if project == nil {
		return ErrProjectNotFound
	}
	return h.repo.Delete(ctx, project.ID)
}

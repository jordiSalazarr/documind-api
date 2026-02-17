package update

import (
	"context"
	"database/sql"

	shared "documind.jordi.org/internal/shared/domain"
)

type Handler struct{ repo Repository }

func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	id := shared.ID(cmd.ID)
	ws, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, ErrWorkspaceNotFound
	}

	if cmd.Name != "" {
		ws.UpdateName(cmd.Name)
	}
	if cmd.Settings != nil {
		ws.UpdateSettings(cmd.Settings)
	}

	if err := h.repo.Update(ctx, ws); err != nil {
		return nil, err
	}

	return &Result{
		ID:        ws.ID.String(),
		Name:      ws.Name,
		Slug:      ws.Slug.String(),
		Settings:  ws.Settings,
		CreatedAt: ws.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: ws.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

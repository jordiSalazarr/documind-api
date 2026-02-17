package getbyslug

import (
	"context"
	"database/sql"

	shared "documind.jordi.org/internal/shared/domain"
)

type Handler struct{ repo Repository }

func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*Result, error) {
	slugVO, err := shared.NewSlug(q.Slug)
	if err != nil {
		return nil, ErrInvalidSlug
	}

	ws, err := h.repo.GetBySlug(ctx, slugVO)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, ErrWorkspaceNotFound
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

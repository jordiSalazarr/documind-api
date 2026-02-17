package get

import (
	"context"
	"database/sql"

	shared "documind.jordi.org/internal/shared/domain"
	wsdomain "documind.jordi.org/internal/identity/workspace/domain"
)

type Handler struct{ repo Repository }

func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*Result, error) {
	ws, err := h.repo.GetByID(ctx, shared.ID(q.ID))
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, ErrWorkspaceNotFound
	}
	return toResult(ws), nil
}

func toResult(ws *wsdomain.Workspace) *Result {
	return &Result{
		ID:        ws.ID.String(),
		Name:      ws.Name,
		Slug:      ws.Slug.String(),
		Settings:  ws.Settings,
		CreatedAt: ws.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: ws.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

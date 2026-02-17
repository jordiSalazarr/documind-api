package get

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
	project, err := h.repo.GetByID(ctx, shared.ID(q.ID))
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}
	return &Result{
		ID: project.ID.String(), WorkspaceID: project.WorkspaceID.String(),
		Name: project.Name, Slug: project.Slug.String(), Description: project.Description,
		CreatedAt: project.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: project.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

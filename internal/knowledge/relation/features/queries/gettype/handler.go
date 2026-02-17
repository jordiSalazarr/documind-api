package gettype

import (
	"context"

	shared "documind.jordi.org/internal/shared/domain"
)

type Handler struct{ repo Repository }

func NewHandler(repo Repository) *Handler { return &Handler{repo: repo} }

func (h *Handler) Handle(ctx context.Context, q Query) (*Result, error) {
	rt, err := h.repo.GetByID(ctx, shared.ID(q.ID))
	if err != nil {
		return nil, err
	}
	if rt == nil {
		return nil, ErrRelationTypeNotFound
	}
	return &Result{
		ID: rt.ID.String(), WorkspaceID: rt.WorkspaceID.String(),
		Name: rt.Name, Slug: rt.Slug.String(), IsDirectional: rt.IsDirectional,
		CreatedAt: rt.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: rt.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

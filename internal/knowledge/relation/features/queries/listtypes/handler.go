package listtypes

import (
	"context"

	shared "documind.jordi.org/internal/shared/domain"
)

const (
	defaultLimit = 50
	maxLimit     = 100
)

type Handler struct{ repo Repository }

func NewHandler(repo Repository) *Handler { return &Handler{repo: repo} }

func (h *Handler) Handle(ctx context.Context, q Query) ([]*Result, error) {
	q.Limit = clampLimit(q.Limit)
	q.Offset = clampOffset(q.Offset)

	types, err := h.repo.ListByWorkspace(ctx, shared.ID(q.WorkspaceID), q.Limit, q.Offset)
	if err != nil {
		return nil, err
	}
	results := make([]*Result, len(types))
	for i, rt := range types {
		results[i] = &Result{
			ID: rt.ID.String(), WorkspaceID: rt.WorkspaceID.String(),
			Name: rt.Name, Slug: rt.Slug.String(), IsDirectional: rt.IsDirectional,
			CreatedAt: rt.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: rt.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	return results, nil
}

func clampLimit(v int) int {
	if v <= 0 {
		return defaultLimit
	}
	if v > maxLimit {
		return maxLimit
	}
	return v
}

func clampOffset(v int) int {
	if v < 0 {
		return 0
	}
	return v
}

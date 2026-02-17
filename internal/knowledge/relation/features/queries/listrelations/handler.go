package listrelations

import (
	"context"

	shared "documind.jordi.org/internal/shared/domain"
)

const (
	defaultLimit = 50
	maxLimit     = 200
)

type Handler struct{ repo Repository }

func NewHandler(repo Repository) *Handler { return &Handler{repo: repo} }

func (h *Handler) Handle(ctx context.Context, q Query) ([]*Result, error) {
	q.Limit = clampLimit(q.Limit)
	q.Offset = clampOffset(q.Offset)

	relations, err := h.repo.ListByItem(ctx, shared.ID(q.ItemID), q.Limit, q.Offset)
	if err != nil {
		return nil, err
	}
	results := make([]*Result, len(relations))
	for i, rel := range relations {
		r := &Result{
			ID: rel.ID.String(), WorkspaceID: rel.WorkspaceID.String(),
			FromItemID: rel.FromItemID.String(), ToItemID: rel.ToItemID.String(),
			RelationTypeID: rel.RelationTypeID.String(),
			CreatedAt: rel.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedBy: rel.CreatedBy.String(),
		}
		if rel.RelationType != nil {
			r.RelationType = &RelationTypeResult{
				ID: rel.RelationType.ID.String(), WorkspaceID: rel.RelationType.WorkspaceID.String(),
				Name: rel.RelationType.Name, Slug: rel.RelationType.Slug.String(),
				IsDirectional: rel.RelationType.IsDirectional,
				CreatedAt: rel.RelationType.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt: rel.RelationType.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
		}
		results[i] = r
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

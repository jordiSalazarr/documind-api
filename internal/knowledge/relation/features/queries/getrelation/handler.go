package getrelation

import (
	"context"

	shared "documind.jordi.org/internal/shared/domain"
)

type Handler struct{ repo Repository }

func NewHandler(repo Repository) *Handler { return &Handler{repo: repo} }

func (h *Handler) Handle(ctx context.Context, q Query) (*Result, error) {
	rel, err := h.repo.GetByID(ctx, shared.ID(q.ID))
	if err != nil {
		return nil, err
	}
	if rel == nil {
		return nil, ErrRelationNotFound
	}
	result := &Result{
		ID: rel.ID.String(), WorkspaceID: rel.WorkspaceID.String(),
		FromItemID: rel.FromItemID.String(), ToItemID: rel.ToItemID.String(),
		RelationTypeID: rel.RelationTypeID.String(),
		CreatedAt: rel.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedBy: rel.CreatedBy.String(),
	}
	if rel.RelationType != nil {
		result.RelationType = &RelationTypeResult{
			ID: rel.RelationType.ID.String(), WorkspaceID: rel.RelationType.WorkspaceID.String(),
			Name: rel.RelationType.Name, Slug: rel.RelationType.Slug.String(),
			IsDirectional: rel.RelationType.IsDirectional,
			CreatedAt: rel.RelationType.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: rel.RelationType.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	return result, nil
}

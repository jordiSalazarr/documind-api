package createrelation

import (
	"context"
	"fmt"

	shared "documind.jordi.org/internal/shared/domain"
	reldomain "documind.jordi.org/internal/knowledge/domain"
)

type Handler struct {
	typeRepo TypeRepository
	relRepo  RelationRepository
}

func NewHandler(typeRepo TypeRepository, relRepo RelationRepository) *Handler {
	return &Handler{typeRepo: typeRepo, relRepo: relRepo}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	relationTypeID := shared.ID(cmd.RelationTypeID)
	rt, err := h.typeRepo.GetByID(ctx, relationTypeID)
	if err != nil {
		return nil, fmt.Errorf("relation type not found: %w", err)
	}
	if rt == nil {
		return nil, ErrRelationTypeNotFound
	}
	fromItemID := shared.ID(cmd.FromItemID)
	toItemID := shared.ID(cmd.ToItemID)
	exists, err := h.relRepo.Exists(ctx, fromItemID, toItemID, relationTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to check relation existence: %w", err)
	}
	if exists {
		return nil, ErrRelationAlreadyExists
	}
	relation, err := reldomain.NewRelation(shared.ID(cmd.WorkspaceID), fromItemID, toItemID, relationTypeID, shared.ID(cmd.CreatedBy))
	if err != nil {
		return nil, err
	}
	if err := h.relRepo.Create(ctx, relation); err != nil {
		return nil, fmt.Errorf("failed to save relation: %w", err)
	}
	saved, err := h.relRepo.GetByID(ctx, relation.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch saved relation: %w", err)
	}
	result := &Result{
		ID: saved.ID.String(), WorkspaceID: saved.WorkspaceID.String(),
		FromItemID: saved.FromItemID.String(), ToItemID: saved.ToItemID.String(),
		RelationTypeID: saved.RelationTypeID.String(),
		CreatedAt: saved.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CreatedBy: saved.CreatedBy.String(),
	}
	if saved.RelationType != nil {
		result.RelationType = &RelationTypeResult{
			ID: saved.RelationType.ID.String(), WorkspaceID: saved.RelationType.WorkspaceID.String(),
			Name: saved.RelationType.Name, Slug: saved.RelationType.Slug.String(),
			IsDirectional: saved.RelationType.IsDirectional,
			CreatedAt: saved.RelationType.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: saved.RelationType.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	return result, nil
}

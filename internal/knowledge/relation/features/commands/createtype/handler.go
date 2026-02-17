package createtype

import (
	"context"
	"fmt"

	shared "documind.jordi.org/internal/shared/domain"
	reldomain "documind.jordi.org/internal/knowledge/domain"
)

type Handler struct{ repo Repository }

func NewHandler(repo Repository) *Handler { return &Handler{repo: repo} }

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	if cmd.Name == "" {
		return nil, ErrInvalidRelationTypeName
	}
	slug, err := shared.NewSlug(cmd.Slug)
	if err != nil {
		return nil, ErrInvalidSlug
	}
	wsID := shared.ID(cmd.WorkspaceID)
	exists, err := h.repo.Exists(ctx, wsID, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to check relation type existence: %w", err)
	}
	if exists {
		return nil, ErrRelationTypeExists
	}
	rt := reldomain.NewRelationType(wsID, cmd.Name, slug, cmd.IsDirectional)
	if err := h.repo.Create(ctx, rt); err != nil {
		return nil, fmt.Errorf("failed to create relation type: %w", err)
	}
	return &Result{
		ID: rt.ID.String(), WorkspaceID: rt.WorkspaceID.String(),
		Name: rt.Name, Slug: rt.Slug.String(), IsDirectional: rt.IsDirectional,
		CreatedAt: rt.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: rt.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

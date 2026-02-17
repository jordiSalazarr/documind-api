package updatetype

import (
	"context"
	"fmt"

	shared "documind.jordi.org/internal/shared/domain"
)

type Handler struct{ repo Repository }

func NewHandler(repo Repository) *Handler { return &Handler{repo: repo} }

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	id := shared.ID(cmd.ID)
	rt, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if rt == nil {
		return nil, ErrRelationTypeNotFound
	}
	if cmd.Name != nil {
		rt.Name = *cmd.Name
	}
	if cmd.Slug != nil {
		slug, err := shared.NewSlug(*cmd.Slug)
		if err != nil {
			return nil, ErrInvalidSlug
		}
		exists, err := h.repo.Exists(ctx, rt.WorkspaceID, slug)
		if err != nil {
			return nil, fmt.Errorf("failed to check relation type existence: %w", err)
		}
		if exists {
			existing, err := h.repo.GetByID(ctx, id)
			if err != nil {
				return nil, err
			}
			if existing.Slug != slug {
				return nil, ErrRelationTypeExists
			}
		}
		rt.Slug = slug
	}
	if cmd.IsDirectional != nil {
		rt.IsDirectional = *cmd.IsDirectional
	}
	rt.Update()
	if err := h.repo.Update(ctx, rt); err != nil {
		return nil, fmt.Errorf("failed to update relation type: %w", err)
	}
	return &Result{
		ID: rt.ID.String(), WorkspaceID: rt.WorkspaceID.String(),
		Name: rt.Name, Slug: rt.Slug.String(), IsDirectional: rt.IsDirectional,
		CreatedAt: rt.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: rt.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

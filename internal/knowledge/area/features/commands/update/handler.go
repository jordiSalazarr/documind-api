package update

import (
	"context"
	"database/sql"

	shared "documind.jordi.org/internal/shared/domain"
)

type Handler struct{ repo Repository }

func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	area, err := h.repo.GetByID(ctx, shared.ID(cmd.ID))
	if err != nil {
		return nil, err
	}
	if area == nil {
		return nil, ErrAreaNotFound
	}
	if cmd.Name != "" {
		area.UpdateName(cmd.Name)
	}
	if cmd.Description != "" {
		area.UpdateDescription(cmd.Description)
	}
	if err := h.repo.Update(ctx, area); err != nil {
		return nil, err
	}
	return &Result{
		ID: area.ID.String(), ProjectID: area.ProjectID.String(),
		Name: area.Name, Slug: area.Slug.String(), Description: area.Description,
		CreatedAt: area.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: area.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

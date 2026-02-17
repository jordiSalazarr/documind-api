package create

import (
	"context"
	"database/sql"

	svcdomain "documind.jordi.org/internal/knowledge/domain"
	shared "documind.jordi.org/internal/shared/domain"
)

type Handler struct {
	repo Repository
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	if cmd.Name == "" {
		return nil, ErrInvalidAreaName
	}
	projectID := shared.ID(cmd.ProjectID)
	exists, err := h.repo.ProjectExists(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrProjectNotFound
	}
	if cmd.Slug == "" {
		cmd.Slug = generateSlug(cmd.Name)
	}
	slug, err := shared.NewSlug(cmd.Slug)
	if err != nil {
		return nil, ErrInvalidSlug
	}
	existing, err := h.repo.GetBySlug(ctx, projectID, slug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAreaExists
	}
	area := svcdomain.NewArea(projectID, cmd.Name, slug, cmd.Description)
	if err := h.repo.Insert(ctx, area); err != nil {
		return nil, err
	}
	return toResult(area), nil
}

func toResult(a *svcdomain.Area) *Result {
	return &Result{
		ID: a.ID.String(), ProjectID: a.ProjectID.String(),
		Name: a.Name, Slug: a.Slug.String(), Description: a.Description,
		CreatedAt: a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: a.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

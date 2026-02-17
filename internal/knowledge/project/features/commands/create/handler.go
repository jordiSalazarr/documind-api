package create

import (
	"context"
	"database/sql"

	projdomain "documind.jordi.org/internal/knowledge/domain"
	shared "documind.jordi.org/internal/shared/domain"
)

type Handler struct{ repo Repository }

func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	if cmd.Name == "" {
		return nil, ErrInvalidProjectName
	}
	if cmd.Slug == "" {
		cmd.Slug = generateSlug(cmd.Name)
	}
	slug, err := shared.NewSlug(cmd.Slug)
	if err != nil {
		return nil, ErrInvalidSlug
	}
	wsID := shared.ID(cmd.WorkspaceID)
	existing, err := h.repo.GetBySlug(ctx, wsID, slug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrProjectExists
	}
	project := projdomain.NewProject(wsID, cmd.Name, slug, cmd.Description, shared.ID(cmd.CreatedBy))
	if err := h.repo.Insert(ctx, project); err != nil {
		return nil, err
	}
	return toResult(project), nil
}

func toResult(p *projdomain.Project) *Result {
	return &Result{
		ID:          p.ID.String(),
		WorkspaceID: p.WorkspaceID.String(),
		Name:        p.Name,
		Slug:        p.Slug.String(),
		Description: p.Description,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

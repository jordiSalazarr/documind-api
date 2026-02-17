package create

import (
	"context"
	"database/sql"

	shared "documind.jordi.org/internal/shared/domain"
	wsdomain "documind.jordi.org/internal/identity/workspace/domain"
)

type Handler struct{ repo Repository }

func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	if cmd.Name == "" {
		return nil, ErrInvalidWorkspaceName
	}
	if cmd.Slug == "" {
		cmd.Slug = generateSlug(cmd.Name)
	}

	slug, err := shared.NewSlug(cmd.Slug)
	if err != nil {
		return nil, ErrInvalidSlug
	}

	existing, err := h.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrWorkspaceExists
	}

	ws := wsdomain.NewWorkspace(cmd.Name, slug)
	if err := h.repo.Insert(ctx, ws); err != nil {
		return nil, err
	}

	// Auto-add creator as admin member
	if cmd.UserID != "" {
		if err := h.repo.InsertMember(ctx, ws.ID.String(), cmd.UserID, cmd.UserEmail, "admin"); err != nil {
			return nil, err
		}
	}

	return &Result{
		ID:        ws.ID.String(),
		Name:      ws.Name,
		Slug:      ws.Slug.String(),
		Settings:  ws.Settings,
		CreatedAt: ws.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: ws.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

package create

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"documind.jordi.org/internal/knowledge/domain"
	shareddomain "documind.jordi.org/internal/shared/domain"
)

type Handler struct{ repo Repository }

func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	slugStr := strings.TrimSpace(cmd.Slug)
	if slugStr == "" {
		slugStr = generateSlug(name)
	}
	slug, err := shareddomain.NewSlug(slugStr)
	if err != nil {
		return nil, err
	}
	existing, err := h.repo.GetBySlug(ctx, cmd.WorkspaceID, slugStr)
	if err == nil && existing != nil {
		return nil, errors.New("an item type with this slug already exists in the workspace")
	}
	itemType := domain.NewItemType(shareddomain.ID(cmd.WorkspaceID), name, slug, cmd.Description, cmd.Icon, cmd.Fields)
	if err := h.repo.Insert(ctx, itemType); err != nil {
		return nil, err
	}
	return toResponse(itemType), nil
}

func toResponse(it *domain.ItemType) *Result {
	return &Result{
		ID: string(it.ID), WorkspaceID: string(it.WorkspaceID),
		Name: it.Name, Slug: string(it.Slug),
		Description: it.Description, Icon: it.Icon, Fields: it.Fields,
		CreatedAt: it.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: it.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

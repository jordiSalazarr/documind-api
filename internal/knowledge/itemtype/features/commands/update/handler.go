package update

import (
	"context"
	"database/sql"
	"errors"
	"strings"
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
	itemType, err := h.repo.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	itemType.Update(name, cmd.Description, cmd.Icon, cmd.Fields)
	if err := h.repo.Update(ctx, itemType); err != nil {
		return nil, err
	}
	return &Result{
		ID: string(itemType.ID), WorkspaceID: string(itemType.WorkspaceID),
		Name: itemType.Name, Slug: string(itemType.Slug),
		Description: itemType.Description, Icon: itemType.Icon, Fields: itemType.Fields,
		CreatedAt: itemType.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: itemType.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

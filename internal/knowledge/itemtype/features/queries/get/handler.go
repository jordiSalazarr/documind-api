package get

import (
	"context"
	"database/sql"
)

type Handler struct{ repo Repository }

func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

func (h *Handler) Handle(ctx context.Context, q Query) (*Result, error) {
	itemType, err := h.repo.GetByID(ctx, q.ID)
	if err != nil {
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

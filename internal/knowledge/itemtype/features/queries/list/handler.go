package list

import (
	"context"
	"database/sql"
)

const (
	defaultLimit = 50
	maxLimit     = 100
)

type Handler struct{ repo Repository }

func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

func (h *Handler) Handle(ctx context.Context, q Query) ([]*Result, error) {
	q.Limit = clampLimit(q.Limit)
	q.Offset = clampOffset(q.Offset)

	itemTypes, err := h.repo.ListByWorkspace(ctx, q.WorkspaceID, q.Limit, q.Offset)
	if err != nil {
		return nil, err
	}
	results := make([]*Result, len(itemTypes))
	for i, it := range itemTypes {
		results[i] = &Result{
			ID: string(it.ID), WorkspaceID: string(it.WorkspaceID),
			Name: it.Name, Slug: string(it.Slug),
			Description: it.Description, Icon: it.Icon, Fields: it.Fields,
			CreatedAt: it.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: it.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	return results, nil
}

func clampLimit(v int) int {
	if v <= 0 {
		return defaultLimit
	}
	if v > maxLimit {
		return maxLimit
	}
	return v
}

func clampOffset(v int) int {
	if v < 0 {
		return 0
	}
	return v
}

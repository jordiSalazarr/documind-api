package list

import (
	"context"
	"database/sql"
)

type Handler struct{ repo Repository }

func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

func (h *Handler) Handle(ctx context.Context, q Query) ([]*Result, error) {
	limit := q.Limit
	offset := q.Offset
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	workspaces, err := h.repo.ListByUser(ctx, q.UserID, limit, offset)
	if err != nil {
		return nil, err
	}

	results := make([]*Result, len(workspaces))
	for i, ws := range workspaces {
		results[i] = &Result{
			ID:        ws.ID.String(),
			Name:      ws.Name,
			Slug:      ws.Slug.String(),
			Settings:  ws.Settings,
			CreatedAt: ws.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: ws.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	return results, nil
}

package list

import (
	"context"
	"database/sql"

	shared "documind.jordi.org/internal/shared/domain"
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

	areas, err := h.repo.ListByProject(ctx, shared.ID(q.ProjectID), q.Limit, q.Offset)
	if err != nil {
		return nil, err
	}
	results := make([]*Result, len(areas))
	for i, a := range areas {
		results[i] = &Result{
			ID: a.ID.String(), ProjectID: a.ProjectID.String(),
			Name: a.Name, Slug: a.Slug.String(), Description: a.Description,
			CreatedAt: a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: a.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
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

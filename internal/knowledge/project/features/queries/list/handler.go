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

	projects, err := h.repo.ListByWorkspace(ctx, shared.ID(q.WorkspaceID), q.Limit, q.Offset)
	if err != nil {
		return nil, err
	}
	results := make([]*Result, len(projects))
	for i, p := range projects {
		results[i] = &Result{
			ID: p.ID.String(), WorkspaceID: p.WorkspaceID.String(),
			Name: p.Name, Slug: p.Slug.String(), Description: p.Description,
			CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
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

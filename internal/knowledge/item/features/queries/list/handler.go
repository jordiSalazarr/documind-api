package list

import (
	"context"
	"database/sql"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// Handler retrieves items by project or service.
type Handler struct {
	repo Repository
}

// NewHandler constructs a Handler.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

// HandleByProject lists items belonging to a project.
func (h *Handler) HandleByProject(ctx context.Context, q ByProjectQuery) (*ListResult, error) {
	q.Limit = clampLimit(q.Limit)
	q.Offset = clampOffset(q.Offset)

	items, err := h.repo.ListByProject(ctx, q.ProjectID, q.Limit, q.Offset)
	if err != nil {
		return nil, err
	}

	count, err := h.repo.CountByProject(ctx, q.ProjectID)
	if err != nil {
		return nil, err
	}

	return &ListResult{Items: items, TotalCount: count}, nil
}

// HandleByService lists items belonging to a service.
func (h *Handler) HandleByService(ctx context.Context, q ByServiceQuery) (*ListResult, error) {
	q.Limit = clampLimit(q.Limit)
	q.Offset = clampOffset(q.Offset)

	items, err := h.repo.ListByService(ctx, q.ServiceID, q.Limit, q.Offset)
	if err != nil {
		return nil, err
	}

	count, err := h.repo.CountByService(ctx, q.ServiceID)
	if err != nil {
		return nil, err
	}

	return &ListResult{Items: items, TotalCount: count}, nil
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

package listversions

import (
	"context"
	"database/sql"

	itemdomain "documind.jordi.org/internal/knowledge/domain"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// Handler retrieves all versions of an item.
type Handler struct {
	repo Repository
}

// NewHandler constructs a Handler.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

// Handle executes the list-item-versions query.
func (h *Handler) Handle(ctx context.Context, q Query) ([]*itemdomain.ItemVersion, error) {
	q.Limit = clampLimit(q.Limit)
	q.Offset = clampOffset(q.Offset)
	return h.repo.ListByItemID(ctx, q.ItemID, q.Limit, q.Offset)
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

package getversion

import (
	"context"
	"database/sql"

	itemdomain "documind.jordi.org/internal/knowledge/domain"
)

// Handler retrieves a specific version of an item.
type Handler struct {
	repo Repository
}

// NewHandler constructs a Handler.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

// Handle executes the get-item-version query.
func (h *Handler) Handle(ctx context.Context, q Query) (*itemdomain.ItemVersion, error) {
	return h.repo.GetVersion(ctx, q.ItemID, q.Version)
}

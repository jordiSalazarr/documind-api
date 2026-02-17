package publish

import (
	"context"
	"database/sql"

	itemdomain "documind.jordi.org/internal/knowledge/domain"
)

// Handler orchestrates item publishing.
type Handler struct {
	repo Repository
}

// NewHandler constructs a Handler.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

// Handle executes the publish-item command.
func (h *Handler) Handle(ctx context.Context, cmd Command) (*itemdomain.Item, error) {
	item, err := h.repo.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	item.Publish(cmd.UpdatedBy)

	if err := h.repo.Update(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

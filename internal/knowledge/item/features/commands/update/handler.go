package update

import (
	"context"
	"database/sql"

	itemdomain "documind.jordi.org/internal/knowledge/domain"
)

// Handler orchestrates item metadata updates.
type Handler struct {
	repo Repository
}

// NewHandler constructs a Handler.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

// Handle executes the update-item command.
func (h *Handler) Handle(ctx context.Context, cmd Command) (*itemdomain.Item, error) {
	item, err := h.repo.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	if cmd.ProjectID != nil {
		item.ProjectID = *cmd.ProjectID
	}

	if cmd.ServiceID != nil {
		item.SetService(*cmd.ServiceID)
	}

	if cmd.OwnerUserID != nil {
		item.OwnerUserID = *cmd.OwnerUserID
	}

	item.Audit.Update(cmd.UpdatedBy)
	item.Timestamp.Update()

	if err := h.repo.Update(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

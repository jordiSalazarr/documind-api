package updatereport

import (
	"context"
	"database/sql"
)

// Handler handles updating a staleness report status.
type Handler struct {
	repo Repository
}

// NewHandler creates a new update report handler.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

// Handle updates the status of a staleness report.
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	if cmd.Status != "dismissed" && cmd.Status != "resolved" {
		return ErrInvalidStatus
	}

	found, err := h.repo.UpdateStatus(ctx, cmd.ID, cmd.WorkspaceID, cmd.Status)
	if err != nil {
		return err
	}
	if !found {
		return ErrReportNotFound
	}
	return nil
}

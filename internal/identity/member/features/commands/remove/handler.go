package remove

import (
	"context"
	"database/sql"
)

type Handler struct {
	repo Repository
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	role, err := h.repo.GetMemberRole(ctx, cmd.WorkspaceID, cmd.UserID)
	if err != nil {
		return err
	}

	if role == "admin" {
		count, err := h.repo.CountAdmins(ctx, cmd.WorkspaceID)
		if err != nil {
			return err
		}
		if count <= 1 {
			return ErrLastAdmin
		}
	}

	return h.repo.Delete(ctx, cmd.WorkspaceID, cmd.UserID)
}

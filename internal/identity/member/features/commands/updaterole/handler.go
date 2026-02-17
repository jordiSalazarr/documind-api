package updaterole

import (
	"context"
	"database/sql"
)

var validRoles = map[string]bool{"reader": true, "editor": true, "admin": true}

type Handler struct {
	repo Repository
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	if !validRoles[cmd.NewRole] {
		return nil, ErrInvalidRole
	}

	currentRole, err := h.repo.GetMemberRole(ctx, cmd.WorkspaceID, cmd.UserID)
	if err != nil {
		return nil, err
	}

	// Prevent demoting the last admin
	if currentRole == "admin" && cmd.NewRole != "admin" {
		count, err := h.repo.CountAdmins(ctx, cmd.WorkspaceID)
		if err != nil {
			return nil, err
		}
		if count <= 1 {
			return nil, ErrLastAdmin
		}
	}

	if err := h.repo.UpdateRole(ctx, cmd.WorkspaceID, cmd.UserID, cmd.NewRole); err != nil {
		return nil, err
	}

	return &Result{
		UserID: cmd.UserID,
		Role:   cmd.NewRole,
	}, nil
}

package invite

import (
	"context"
	"database/sql"
	"strings"
)

var validRoles = map[string]bool{"reader": true, "editor": true, "admin": true}

type Handler struct {
	repo Repository
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{repo: newPostgresRepo(db)}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	cmd.Email = strings.TrimSpace(strings.ToLower(cmd.Email))

	if cmd.Email == "" || !strings.Contains(cmd.Email, "@") {
		return nil, ErrInvalidEmail
	}
	if !validRoles[cmd.Role] {
		return nil, ErrInvalidRole
	}

	userID, err := h.repo.FindUserByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, err
	}

	memberID, err := h.repo.InsertMember(ctx, cmd.WorkspaceID, userID, cmd.Email, cmd.Role, cmd.InvitedBy)
	if err != nil {
		return nil, err
	}

	return &Result{
		ID:     memberID,
		UserID: userID,
		Email:  cmd.Email,
		Role:   cmd.Role,
	}, nil
}

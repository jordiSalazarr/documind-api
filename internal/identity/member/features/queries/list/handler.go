package list

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

func (h *Handler) Handle(ctx context.Context, q Query) ([]MemberResult, error) {
	return h.repo.ListByWorkspace(ctx, q.WorkspaceID)
}

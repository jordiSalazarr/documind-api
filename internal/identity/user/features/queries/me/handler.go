package me

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

func (h *Handler) Handle(ctx context.Context, q Query) (*Result, error) {
	return h.repo.GetByID(ctx, q.UserID)
}

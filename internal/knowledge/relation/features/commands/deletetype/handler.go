package deletetype

import (
	"context"

	shared "documind.jordi.org/internal/shared/domain"
)

type Handler struct{ repo Repository }

func NewHandler(repo Repository) *Handler { return &Handler{repo: repo} }

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	id := shared.ID(cmd.ID)
	rt, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if rt == nil {
		return ErrRelationTypeNotFound
	}
	return h.repo.Delete(ctx, id)
}

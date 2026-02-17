package deleterelation

import (
	"context"

	shared "documind.jordi.org/internal/shared/domain"
)

type Handler struct{ repo Repository }

func NewHandler(repo Repository) *Handler { return &Handler{repo: repo} }

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	id := shared.ID(cmd.ID)
	rel, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if rel == nil {
		return ErrRelationNotFound
	}
	return h.repo.Delete(ctx, id)
}

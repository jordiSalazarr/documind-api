package deleteconversation

import (
	"context"

	sharedDomain "documind.jordi.org/internal/shared/domain"
)

// Handler handles conversation deletion
type Handler struct {
	deleter ConversationDeleter
}

// NewHandler creates a new handler
func NewHandler(deleter ConversationDeleter) *Handler {
	return &Handler{deleter: deleter}
}

// Handle soft-deletes a conversation
func (h *Handler) Handle(ctx context.Context, id sharedDomain.ID) error {
	return h.deleter.Delete(id)
}

package getmessages

import (
	"context"

	chatDomain "documind.jordi.org/internal/conversation/domain"
	sharedDomain "documind.jordi.org/internal/shared/domain"
)

// Handler handles retrieving messages for a conversation
type Handler struct {
	lister MessageLister
}

// NewHandler creates a new handler
func NewHandler(lister MessageLister) *Handler {
	return &Handler{lister: lister}
}

// Handle retrieves messages for a conversation
func (h *Handler) Handle(ctx context.Context, conversationID sharedDomain.ID, limit, offset int) ([]*chatDomain.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return h.lister.ListByConversationID(conversationID, limit, offset)
}

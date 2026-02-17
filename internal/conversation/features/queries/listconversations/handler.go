package listconversations

import (
	"context"

	chatDomain "documind.jordi.org/internal/conversation/domain"
	sharedDomain "documind.jordi.org/internal/shared/domain"
)

// Handler handles listing conversations for a service
type Handler struct {
	lister ConversationLister
}

// NewHandler creates a new handler
func NewHandler(lister ConversationLister) *Handler {
	return &Handler{lister: lister}
}

// Handle lists conversations for a service
func (h *Handler) Handle(ctx context.Context, serviceID sharedDomain.ID, limit, offset int) ([]*chatDomain.Conversation, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return h.lister.ListByServiceID(serviceID, limit, offset)
}

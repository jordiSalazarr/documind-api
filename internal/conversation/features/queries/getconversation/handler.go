package getconversation

import (
	"context"

	chatDomain "documind.jordi.org/internal/conversation/domain"
	sharedDomain "documind.jordi.org/internal/shared/domain"
)

// Handler handles retrieving a conversation by ID
type Handler struct {
	reader ConversationReader
}

// NewHandler creates a new handler
func NewHandler(reader ConversationReader) *Handler {
	return &Handler{reader: reader}
}

// Handle retrieves a conversation by ID
func (h *Handler) Handle(ctx context.Context, id sharedDomain.ID) (*chatDomain.Conversation, error) {
	return h.reader.GetByID(id)
}

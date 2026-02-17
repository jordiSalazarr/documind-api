package createconversation

import (
	"context"

	chatDomain "documind.jordi.org/internal/conversation/domain"
)

// Handler handles conversation creation
type Handler struct {
	writer ConversationWriter
}

// NewHandler creates a new handler
func NewHandler(writer ConversationWriter) *Handler {
	return &Handler{writer: writer}
}

// Handle creates a new conversation
func (h *Handler) Handle(ctx context.Context, cmd Command) (*chatDomain.Conversation, error) {
	conversation := chatDomain.NewConversation(
		cmd.WorkspaceID,
		cmd.ServiceID,
		cmd.UserID,
	)

	if err := h.writer.Create(conversation); err != nil {
		return nil, err
	}

	return conversation, nil
}

package createconversation

import (
	chatDomain "documind.jordi.org/internal/conversation/domain"
)

// ConversationWriter defines the write interface needed for conversation creation
type ConversationWriter interface {
	Create(conversation *chatDomain.Conversation) error
}

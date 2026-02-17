package getconversation

import (
	chatDomain "documind.jordi.org/internal/conversation/domain"
	sharedDomain "documind.jordi.org/internal/shared/domain"
)

// ConversationReader defines the read interface for retrieving conversations
type ConversationReader interface {
	GetByID(id sharedDomain.ID) (*chatDomain.Conversation, error)
}

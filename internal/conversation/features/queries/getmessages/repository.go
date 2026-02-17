package getmessages

import (
	chatDomain "documind.jordi.org/internal/conversation/domain"
	sharedDomain "documind.jordi.org/internal/shared/domain"
)

// MessageLister defines the read interface for listing messages
type MessageLister interface {
	ListByConversationID(conversationID sharedDomain.ID, limit, offset int) ([]*chatDomain.Message, error)
}

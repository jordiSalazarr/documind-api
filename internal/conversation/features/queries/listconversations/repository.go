package listconversations

import (
	chatDomain "documind.jordi.org/internal/conversation/domain"
	sharedDomain "documind.jordi.org/internal/shared/domain"
)

// ConversationLister defines the read interface for listing conversations
type ConversationLister interface {
	ListByServiceID(serviceID sharedDomain.ID, limit, offset int) ([]*chatDomain.Conversation, error)
}

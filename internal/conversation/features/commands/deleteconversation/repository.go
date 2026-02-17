package deleteconversation

import (
	sharedDomain "documind.jordi.org/internal/shared/domain"
)

// ConversationDeleter defines the interface for deleting conversations
type ConversationDeleter interface {
	Delete(id sharedDomain.ID) error
}

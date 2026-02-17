package createconversation

import (
	sharedDomain "documind.jordi.org/internal/shared/domain"
)

// Command contains parameters for creating a conversation
type Command struct {
	WorkspaceID sharedDomain.ID
	ServiceID   sharedDomain.ID
	UserID      sharedDomain.ID
}

package chat

import (
	"time"

	"documind.jordi.org/internal/domain/common"
)

// Conversation represents a chat session with a service
type Conversation struct {
	ID           common.ID
	WorkspaceID  common.ID
	ServiceID    common.ID
	UserID       common.ID
	Title        string
	Summary      string     // Rolling summary of conversation history
	SummaryUpTo  *common.ID // ID of the last message included in summary
	MessageCount int        // Total message count for triggering summary updates
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

// NewConversation creates a new conversation
func NewConversation(
	workspaceID common.ID,
	serviceID common.ID,
	userID common.ID,
) *Conversation {
	now := time.Now()
	return &Conversation{
		ID:          common.NewID(),
		WorkspaceID: workspaceID,
		ServiceID:   serviceID,
		UserID:      userID,
		Title:       "New Conversation",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// SetTitle updates the conversation title
func (c *Conversation) SetTitle(title string) {
	c.Title = title
	c.UpdatedAt = time.Now()
}

// GenerateTitleFromQuery sets the title based on the first user query
func (c *Conversation) GenerateTitleFromQuery(query string) {
	// Truncate to first 50 characters and add ellipsis if needed
	title := query
	if len(title) > 50 {
		title = title[:47] + "..."
	}
	c.SetTitle(title)
}

// Touch updates the UpdatedAt timestamp
func (c *Conversation) Touch() {
	c.UpdatedAt = time.Now()
}

// UpdateSummary updates the conversation summary
func (c *Conversation) UpdateSummary(summary string, lastMessageID common.ID) {
	c.Summary = summary
	c.SummaryUpTo = &lastMessageID
	c.UpdatedAt = time.Now()
}

// IncrementMessageCount increments the message counter
func (c *Conversation) IncrementMessageCount() {
	c.MessageCount++
	c.UpdatedAt = time.Now()
}

// NeedsSummaryUpdate checks if the conversation needs a summary update
// Returns true if there are more than `threshold` messages since last summary
func (c *Conversation) NeedsSummaryUpdate(threshold int) bool {
	// No summary yet and we have enough messages
	if c.SummaryUpTo == nil && c.MessageCount >= threshold {
		return true
	}
	// This is a simplified check - the service will do the actual count
	return false
}

// SoftDelete marks the conversation as deleted
func (c *Conversation) SoftDelete() {
	now := time.Now()
	c.DeletedAt = &now
}

// IsDeleted checks if the conversation is soft-deleted
func (c *Conversation) IsDeleted() bool {
	return c.DeletedAt != nil
}

// ConversationRepository defines the interface for conversation persistence
type ConversationRepository interface {
	// Create creates a new conversation
	Create(conversation *Conversation) error

	// GetByID retrieves a conversation by ID
	GetByID(id common.ID) (*Conversation, error)

	// ListByServiceID retrieves conversations for a service
	ListByServiceID(serviceID common.ID, limit, offset int) ([]*Conversation, error)

	// ListByUserID retrieves conversations for a user within a workspace
	ListByUserID(workspaceID, userID common.ID, limit, offset int) ([]*Conversation, error)

	// Update updates a conversation
	Update(conversation *Conversation) error

	// Delete soft-deletes a conversation
	Delete(id common.ID) error
}

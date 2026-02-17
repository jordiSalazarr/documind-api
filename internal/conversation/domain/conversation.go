package domain

import (
	"time"

	sharedDomain "documind.jordi.org/internal/shared/domain"
)

// Conversation represents a chat session with a service
type Conversation struct {
	ID           sharedDomain.ID
	WorkspaceID  sharedDomain.ID
	ServiceID    sharedDomain.ID
	UserID       sharedDomain.ID
	Title        string
	Summary      string          // Rolling summary of conversation history
	SummaryUpTo  *sharedDomain.ID // ID of the last message included in summary
	MessageCount int             // Total message count for triggering summary updates
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

// NewConversation creates a new conversation
func NewConversation(
	workspaceID sharedDomain.ID,
	serviceID sharedDomain.ID,
	userID sharedDomain.ID,
) *Conversation {
	now := time.Now()
	return &Conversation{
		ID:          sharedDomain.NewID(),
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
func (c *Conversation) UpdateSummary(summary string, lastMessageID sharedDomain.ID) {
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

// ConversationWriteRepository defines the write interface for conversation persistence
type ConversationWriteRepository interface {
	// Create creates a new conversation
	Create(conversation *Conversation) error

	// Update updates a conversation
	Update(conversation *Conversation) error

	// Delete soft-deletes a conversation
	Delete(id sharedDomain.ID) error
}

// ConversationReadRepository defines the read interface for conversation persistence
type ConversationReadRepository interface {
	// GetByID retrieves a conversation by ID
	GetByID(id sharedDomain.ID) (*Conversation, error)

	// ListByServiceID retrieves conversations for a service
	ListByServiceID(serviceID sharedDomain.ID, limit, offset int) ([]*Conversation, error)

	// ListByUserID retrieves conversations for a user within a workspace
	ListByUserID(workspaceID, userID sharedDomain.ID, limit, offset int) ([]*Conversation, error)
}

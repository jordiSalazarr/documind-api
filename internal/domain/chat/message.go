package chat

import (
	"time"

	"documind.jordi.org/internal/domain/common"
)

// MessageRole represents who sent the message
type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
)

// Source represents a document chunk used as context for the response
type Source struct {
	ChunkID   common.ID `json:"chunk_id"`
	ItemID    common.ID `json:"item_id"`
	ItemTitle string    `json:"item_title"`
	Snippet   string    `json:"snippet"`
	Score     float64   `json:"score"`
}

// Message represents a single message in a conversation
type Message struct {
	ID             common.ID
	ConversationID common.ID
	WorkspaceID    common.ID
	Role           MessageRole
	Content        string
	Sources        []Source
	TokenCount     int    // Total tokens (prompt + completion)
	Model          string // LLM model used (e.g., gpt-4o-mini)
	LatencyMs      int    // Response generation time
	CreatedAt      time.Time
}

// NewUserMessage creates a new user message
func NewUserMessage(
	conversationID common.ID,
	workspaceID common.ID,
	content string,
) *Message {
	return &Message{
		ID:             common.NewID(),
		ConversationID: conversationID,
		WorkspaceID:    workspaceID,
		Role:           MessageRoleUser,
		Content:        content,
		Sources:        []Source{},
		CreatedAt:      time.Now(),
	}
}

// NewAssistantMessage creates a new assistant (AI) message
func NewAssistantMessage(
	conversationID common.ID,
	workspaceID common.ID,
	content string,
	sources []Source,
	model string,
	tokenCount int,
	latencyMs int,
) *Message {
	return &Message{
		ID:             common.NewID(),
		ConversationID: conversationID,
		WorkspaceID:    workspaceID,
		Role:           MessageRoleAssistant,
		Content:        content,
		Sources:        sources,
		TokenCount:     tokenCount,
		Model:          model,
		LatencyMs:      latencyMs,
		CreatedAt:      time.Now(),
	}
}

// NewSystemMessage creates a new system message
func NewSystemMessage(
	conversationID common.ID,
	workspaceID common.ID,
	content string,
) *Message {
	return &Message{
		ID:             common.NewID(),
		ConversationID: conversationID,
		WorkspaceID:    workspaceID,
		Role:           MessageRoleSystem,
		Content:        content,
		Sources:        []Source{},
		CreatedAt:      time.Now(),
	}
}

// IsFromUser checks if the message is from a user
func (m *Message) IsFromUser() bool {
	return m.Role == MessageRoleUser
}

// IsFromAssistant checks if the message is from the assistant
func (m *Message) IsFromAssistant() bool {
	return m.Role == MessageRoleAssistant
}

// HasSources checks if the message has source citations
func (m *Message) HasSources() bool {
	return len(m.Sources) > 0
}

// MessageRepository defines the interface for message persistence
type MessageRepository interface {
	// Create creates a new message
	Create(message *Message) error

	// GetByID retrieves a message by ID
	GetByID(id common.ID) (*Message, error)

	// ListByConversationID retrieves messages for a conversation (ordered by created_at)
	ListByConversationID(conversationID common.ID, limit, offset int) ([]*Message, error)

	// GetLastMessages retrieves the last N messages for context window
	GetLastMessages(conversationID common.ID, limit int) ([]*Message, error)

	// CountByConversationID counts messages in a conversation
	CountByConversationID(conversationID common.ID) (int, error)
}

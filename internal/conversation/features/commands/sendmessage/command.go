package sendmessage

import (
	"context"

	chatDomain "documind.jordi.org/internal/conversation/domain"
	sharedDomain "documind.jordi.org/internal/shared/domain"
	"documind.jordi.org/internal/shared/infrastructure/openai"
)

// Retriever defines the interface for retrieving context documents.
type Retriever interface {
	Retrieve(ctx context.Context, input RetrievalInput) (*RetrievalOutput, error)
}

// RetrievalInput contains parameters for retrieval
type RetrievalInput struct {
	Query       string
	WorkspaceID sharedDomain.ID
	MaxResults  int
	MaxTokens   int
}

// RetrievalOutput contains retrieval results
type RetrievalOutput struct {
	Results []*RetrievalResult
}

// RetrievalResult represents a single retrieval result with chunk information
type RetrievalResult struct {
	ChunkID       sharedDomain.ID
	ItemVersionID sharedDomain.ID
	Heading       string
	Content       string
	TokenCount    int
	FinalScore    float64
}

// ConversationReader defines the read interface for conversations
type ConversationReader interface {
	GetByID(id sharedDomain.ID) (*chatDomain.Conversation, error)
}

// ConversationWriter defines the write interface for conversations
type ConversationWriter interface {
	Update(conversation *chatDomain.Conversation) error
}

// MessageReader defines the read interface for messages
type MessageReader interface {
	GetLastMessages(conversationID sharedDomain.ID, limit int) ([]*chatDomain.Message, error)
}

// MessageWriter defines the write interface for messages
type MessageWriter interface {
	Create(message *chatDomain.Message) error
}

// ChatClient defines the interface for LLM chat operations
type ChatClient interface {
	CreateCompletion(ctx context.Context, messages []openai.ChatMessage, maxTokens int) (*openai.ChatResponse, error)
	CreateCompletionStream(ctx context.Context, messages []openai.ChatMessage, maxTokens int, handler openai.StreamHandler) (*openai.ChatUsage, error)
	GetModel() string
}

// ConversationContext contains the prepared context for a conversation.
// This mirrors the chat.ConversationContext type to avoid circular imports.
type ConversationContext struct {
	Summary             string
	RecentMessages      []*chatDomain.Message
	NeedsSummaryUpdate  bool
	MessagesToSummarize []*chatDomain.Message
}

// MemoryService defines the interface for conversation memory operations.
// This is satisfied by chat.MemoryService in the parent package.
type MemoryService interface {
	GetConversationContext(ctx context.Context, conversation *chatDomain.Conversation, excludeMessageID sharedDomain.ID) (*ConversationContext, error)
	EnhanceQueryWithContext(ctx context.Context, query string, conversationContext *ConversationContext) (string, error)
	BuildMessagesForLLM(systemPrompt string, contextPrompt string, conversationContext *ConversationContext, currentQuery string) []openai.ChatMessage
	GenerateSummary(ctx context.Context, conversation *chatDomain.Conversation, existingSummary string) (string, sharedDomain.ID, error)
}

// RetrievalContext represents context from retrieved documents
type RetrievalContext struct {
	SourceIndex int
	ItemTitle   string
	Content     string
	Snippet     string
}

// Command contains parameters for sending a message
type Command struct {
	ConversationID sharedDomain.ID
	WorkspaceID    sharedDomain.ID
	ServiceID      sharedDomain.ID
	Query          string
	UserID         sharedDomain.ID
}

// Result contains the result of sending a message
type Result struct {
	UserMessage      *chatDomain.Message
	AssistantMessage *chatDomain.Message
	Sources          []chatDomain.Source
}

// Config holds configuration for the send message handler
type Config struct {
	MaxContextTokens   int  // Max tokens for context (default: 4000)
	MaxChunks          int  // Max chunks to retrieve (default: 5)
	EnableMemory       bool // Enable conversation memory with summarization (default: true)
	EnableQueryEnhance bool // Enable context-aware query enhancement (default: true)
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		MaxContextTokens:   4000,
		MaxChunks:          5,
		EnableMemory:       true,
		EnableQueryEnhance: true,
	}
}

// StreamCallback is a callback for streaming responses
type StreamCallback func(token string, done bool) error

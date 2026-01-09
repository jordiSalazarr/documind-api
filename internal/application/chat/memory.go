package chat

import (
	"context"
	"fmt"
	"strings"

	"documind.jordi.org/internal/domain/chat"
	"documind.jordi.org/internal/domain/common"
	"documind.jordi.org/internal/infrastructure/openai"
)

// MemoryConfig holds configuration for conversation memory
type MemoryConfig struct {
	// SlidingWindowSize is the number of recent messages to keep verbatim
	SlidingWindowSize int
	// SummaryThreshold is the number of messages before triggering a summary update
	SummaryThreshold int
	// MaxSummaryTokens is the max tokens for the summary
	MaxSummaryTokens int
}

// DefaultMemoryConfig returns sensible defaults for memory configuration
func DefaultMemoryConfig() MemoryConfig {
	return MemoryConfig{
		SlidingWindowSize: 6,  // Keep last 3 exchanges (6 messages)
		SummaryThreshold:  10, // Summarize after 10 messages
		MaxSummaryTokens:  500,
	}
}

// MemoryService handles conversation memory and summarization
type MemoryService struct {
	messageRepo chat.MessageRepository
	chatClient  *openai.ChatClient
	config      MemoryConfig
}

// NewMemoryService creates a new memory service
func NewMemoryService(
	messageRepo chat.MessageRepository,
	chatClient *openai.ChatClient,
	config MemoryConfig,
) *MemoryService {
	return &MemoryService{
		messageRepo: messageRepo,
		chatClient:  chatClient,
		config:      config,
	}
}

// ConversationContext contains the prepared context for a conversation
type ConversationContext struct {
	// Summary of older messages (if any)
	Summary string
	// Recent messages within the sliding window
	RecentMessages []*chat.Message
	// Whether a summary update is needed after this interaction
	NeedsSummaryUpdate bool
	// Messages that should be included in the next summary
	MessagesToSummarize []*chat.Message
}

// GetConversationContext retrieves the conversation context with sliding window + summary
func (s *MemoryService) GetConversationContext(
	ctx context.Context,
	conversation *chat.Conversation,
	excludeMessageID common.ID,
) (*ConversationContext, error) {
	// Get total message count
	totalMessages, err := s.messageRepo.CountByConversationID(conversation.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to count messages: %w", err)
	}

	result := &ConversationContext{
		Summary: conversation.Summary,
	}

	// Get recent messages for the sliding window
	recentMessages, err := s.messageRepo.GetLastMessages(conversation.ID, s.config.SlidingWindowSize+1)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent messages: %w", err)
	}

	// Filter out the excluded message (usually the current user message being processed)
	filtered := make([]*chat.Message, 0, len(recentMessages))
	for _, msg := range recentMessages {
		if msg.ID != excludeMessageID {
			filtered = append(filtered, msg)
		}
	}
	result.RecentMessages = filtered

	// Check if we need to update the summary
	// We need a summary update if:
	// 1. Total messages exceed threshold AND
	// 2. Either no summary exists OR there are messages not yet summarized
	if totalMessages >= s.config.SummaryThreshold {
		if conversation.Summary == "" {
			result.NeedsSummaryUpdate = true
		} else if conversation.SummaryUpTo != nil {
			// Count messages after the last summarized message
			// For simplicity, we'll trigger update when we have more than SlidingWindowSize unsummarized
			unsummarizedCount := totalMessages - s.config.SlidingWindowSize
			if unsummarizedCount > s.config.SummaryThreshold/2 {
				result.NeedsSummaryUpdate = true
			}
		}
	}

	return result, nil
}

// GenerateSummary creates or updates the conversation summary
func (s *MemoryService) GenerateSummary(
	ctx context.Context,
	conversation *chat.Conversation,
	existingSummary string,
) (string, common.ID, error) {
	// Get all messages except the most recent ones (which will stay in sliding window)
	allMessages, err := s.messageRepo.ListByConversationID(conversation.ID, 100, 0)
	if err != nil {
		return "", "", fmt.Errorf("failed to get messages: %w", err)
	}

	if len(allMessages) <= s.config.SlidingWindowSize {
		// Not enough messages to summarize
		return existingSummary, "", nil
	}

	// Messages to summarize (all except the sliding window)
	messagesToSummarize := allMessages[:len(allMessages)-s.config.SlidingWindowSize]
	if len(messagesToSummarize) == 0 {
		return existingSummary, "", nil
	}

	lastMessageID := messagesToSummarize[len(messagesToSummarize)-1].ID

	// Build the conversation text to summarize
	var conversationText strings.Builder
	for _, msg := range messagesToSummarize {
		role := "User"
		if msg.Role == chat.MessageRoleAssistant {
			role = "Assistant"
		}
		conversationText.WriteString(fmt.Sprintf("%s: %s\n\n", role, msg.Content))
	}

	// Build summary prompt
	prompt := buildSummaryPrompt(existingSummary, conversationText.String())

	messages := []openai.ChatMessage{
		{
			Role:    "system",
			Content: SummarySystemPrompt,
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	resp, err := s.chatClient.CreateCompletion(ctx, messages, s.config.MaxSummaryTokens)
	if err != nil {
		return existingSummary, "", fmt.Errorf("failed to generate summary: %w", err)
	}

	if len(resp.Choices) == 0 {
		return existingSummary, "", nil
	}

	return resp.Choices[0].Message.Content, lastMessageID, nil
}

// EnhanceQueryWithContext uses conversation context to improve the retrieval query
func (s *MemoryService) EnhanceQueryWithContext(
	ctx context.Context,
	query string,
	conversationContext *ConversationContext,
) (string, error) {
	// If no context, return original query
	if conversationContext.Summary == "" && len(conversationContext.RecentMessages) == 0 {
		return query, nil
	}

	// Build context for query enhancement
	var contextBuilder strings.Builder

	if conversationContext.Summary != "" {
		contextBuilder.WriteString("Conversation summary:\n")
		contextBuilder.WriteString(conversationContext.Summary)
		contextBuilder.WriteString("\n\n")
	}

	if len(conversationContext.RecentMessages) > 0 {
		contextBuilder.WriteString("Recent exchanges:\n")
		for _, msg := range conversationContext.RecentMessages {
			role := "User"
			if msg.Role == chat.MessageRoleAssistant {
				role = "Assistant"
			}
			// Truncate long messages for context
			content := msg.Content
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			contextBuilder.WriteString(fmt.Sprintf("%s: %s\n", role, content))
		}
	}

	prompt := fmt.Sprintf(`Given this conversation context:
%s

The user now asks: "%s"

Rewrite this query to be self-contained and optimized for searching a document database.
Include relevant context from the conversation that helps clarify what the user is looking for.
Only output the enhanced query, nothing else. Keep it concise (under 100 words).`,
		contextBuilder.String(), query)

	messages := []openai.ChatMessage{
		{
			Role:    "system",
			Content: "You are a query enhancement assistant. Your job is to rewrite user queries to be more effective for document retrieval by incorporating relevant conversation context.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	resp, err := s.chatClient.CreateCompletion(ctx, messages, 150)
	if err != nil {
		// Fall back to original query on error
		return query, nil
	}

	if len(resp.Choices) == 0 {
		return query, nil
	}

	enhancedQuery := strings.TrimSpace(resp.Choices[0].Message.Content)
	if enhancedQuery == "" {
		return query, nil
	}
	fmt.Printf("enhanced query: %s", enhancedQuery)
	return enhancedQuery, nil
}

// BuildMessagesForLLM builds the message list for the LLM including memory context
func (s *MemoryService) BuildMessagesForLLM(
	systemPrompt string,
	contextPrompt string,
	conversationContext *ConversationContext,
	currentQuery string,
) []openai.ChatMessage {
	messages := []openai.ChatMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	// Add summary as context if available
	if conversationContext.Summary != "" {
		messages = append(messages, openai.ChatMessage{
			Role:    "system",
			Content: fmt.Sprintf("Previous conversation summary:\n%s", conversationContext.Summary),
		})
	}

	// Add document context
	if contextPrompt != "" {
		messages = append(messages, openai.ChatMessage{
			Role:    "system",
			Content: contextPrompt,
		})
	}

	// Add recent messages from sliding window
	for _, msg := range conversationContext.RecentMessages {
		messages = append(messages, openai.ChatMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	// Add current query
	messages = append(messages, openai.ChatMessage{
		Role:    "user",
		Content: currentQuery,
	})

	return messages
}

// SummarySystemPrompt is the system prompt for generating conversation summaries
const SummarySystemPrompt = `You are a conversation summarizer. Your job is to create concise summaries of conversations that capture:
1. The main topics discussed
2. Key questions asked by the user
3. Important information or answers provided
4. Any ongoing context that would be needed to understand follow-up questions

Keep summaries factual and concise. Focus on information that would help understand future questions in the conversation.`

// buildSummaryPrompt builds the prompt for summary generation
func buildSummaryPrompt(existingSummary, newConversation string) string {
	if existingSummary == "" {
		return fmt.Sprintf(`Summarize this conversation concisely:

%s

Provide a summary that captures the key topics, questions, and information exchanged.`, newConversation)
	}

	return fmt.Sprintf(`Update this existing conversation summary with new messages:

EXISTING SUMMARY:
%s

NEW MESSAGES TO INCORPORATE:
%s

Provide an updated summary that incorporates the new information while keeping it concise.`, existingSummary, newConversation)
}

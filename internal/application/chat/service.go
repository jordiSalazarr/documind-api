package chat

import (
	"context"
	"errors"
	"time"

	"documind.jordi.org/internal/application/retrieval"
	"documind.jordi.org/internal/domain/chat"
	"documind.jordi.org/internal/domain/common"
	"documind.jordi.org/internal/domain/knowledge"
	"documind.jordi.org/internal/infrastructure/openai"
)

// Service handles chat operations and RAG
type Service struct {
	conversationRepo   chat.ConversationRepository
	messageRepo        chat.MessageRepository
	retrievalService   *retrieval.Service
	chatClient         *openai.ChatClient
	memoryService      *MemoryService
	maxContextTokens   int
	maxChunks          int
	useHybridRetrieval bool
	enableMemory       bool
	enableQueryEnhance bool
}

// ServiceConfig holds configuration for the chat service
type ServiceConfig struct {
	MaxContextTokens   int  // Max tokens for context (default: 4000)
	MaxChunks          int  // Max chunks to retrieve (default: 5)
	UseHybridRetrieval bool // Use hybrid retrieval (default: true)
	EnableMemory       bool // Enable conversation memory with summarization (default: true)
	EnableQueryEnhance bool // Enable context-aware query enhancement (default: true)
}

// NewService creates a new chat service
func NewService(
	conversationRepo chat.ConversationRepository,
	messageRepo chat.MessageRepository,
	retrievalService *retrieval.Service,
	chatClient *openai.ChatClient,
	config ServiceConfig,
) *Service {
	if config.MaxContextTokens <= 0 {
		config.MaxContextTokens = 4000
	}
	if config.MaxChunks <= 0 {
		config.MaxChunks = 5
	}

	// Create memory service
	memoryService := NewMemoryService(
		messageRepo,
		chatClient,
		DefaultMemoryConfig(),
	)

	return &Service{
		conversationRepo:   conversationRepo,
		messageRepo:        messageRepo,
		retrievalService:   retrievalService,
		chatClient:         chatClient,
		memoryService:      memoryService,
		maxContextTokens:   config.MaxContextTokens,
		maxChunks:          config.MaxChunks,
		useHybridRetrieval: config.UseHybridRetrieval,
		enableMemory:       config.EnableMemory,
		enableQueryEnhance: config.EnableQueryEnhance,
	}
}

// CreateConversationInput contains parameters for creating a conversation
type CreateConversationInput struct {
	WorkspaceID common.ID
	ServiceID   common.ID
	UserID      common.ID
}

// CreateConversation creates a new chat conversation
func (s *Service) CreateConversation(ctx context.Context, input CreateConversationInput) (*chat.Conversation, error) {
	conversation := chat.NewConversation(
		input.WorkspaceID,
		input.ServiceID,
		input.UserID,
	)

	if err := s.conversationRepo.Create(conversation); err != nil {
		return nil, err
	}

	return conversation, nil
}

// GetConversation retrieves a conversation by ID
func (s *Service) GetConversation(ctx context.Context, id common.ID) (*chat.Conversation, error) {
	return s.conversationRepo.GetByID(id)
}

// ListConversations lists conversations for a service
func (s *Service) ListConversations(ctx context.Context, serviceID common.ID, limit, offset int) ([]*chat.Conversation, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.conversationRepo.ListByServiceID(serviceID, limit, offset)
}

// GetMessages retrieves messages for a conversation
func (s *Service) GetMessages(ctx context.Context, conversationID common.ID, limit, offset int) ([]*chat.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return s.messageRepo.ListByConversationID(conversationID, limit, offset)
}

// SendMessageInput contains parameters for sending a message
type SendMessageInput struct {
	ConversationID common.ID
	WorkspaceID    common.ID
	ServiceID      common.ID
	Query          string
	UserID         common.ID
}

// SendMessageResult contains the result of sending a message
type SendMessageResult struct {
	UserMessage      *chat.Message
	AssistantMessage *chat.Message
	Sources          []chat.Source
}

// SendMessage sends a user message and generates an AI response
func (s *Service) SendMessage(ctx context.Context, input SendMessageInput) (*SendMessageResult, error) {
	if input.Query == "" {
		return nil, errors.New("query cannot be empty")
	}

	start := time.Now()

	// Create user message
	userMessage := chat.NewUserMessage(
		input.ConversationID,
		input.WorkspaceID,
		input.Query,
	)
	if err := s.messageRepo.Create(userMessage); err != nil {
		return nil, err
	}

	// Get conversation and update title if first message
	conversation, err := s.conversationRepo.GetByID(input.ConversationID)
	if err != nil {
		return nil, err
	}

	// Increment message count
	conversation.IncrementMessageCount()

	// Update title from first query if still default
	if conversation.Title == "New Conversation" {
		conversation.GenerateTitleFromQuery(input.Query)
	}

	// Get conversation context (summary + sliding window)
	var conversationContext *ConversationContext
	if s.enableMemory {
		conversationContext, err = s.memoryService.GetConversationContext(ctx, conversation, userMessage.ID)
		if err != nil {
			// Fall back to empty context on error
			conversationContext = &ConversationContext{}
		}
	} else {
		conversationContext = &ConversationContext{}
	}

	// Determine the query to use for retrieval
	retrievalQuery := input.Query
	if s.enableQueryEnhance && (conversationContext.Summary != "" || len(conversationContext.RecentMessages) > 0) {
		enhancedQuery, err := s.memoryService.EnhanceQueryWithContext(ctx, input.Query, conversationContext)
		if err == nil && enhancedQuery != "" {
			retrievalQuery = enhancedQuery
		}
	}

	// Retrieve relevant chunks using hybrid retrieval with potentially enhanced query
	retrievalResult, err := s.retrievalService.Retrieve(ctx, retrieval.RetrieveInput{
		Query:       retrievalQuery,
		WorkspaceID: input.WorkspaceID,
		MaxResults:  s.maxChunks,
		MaxTokens:   s.maxContextTokens,
	})

	var retrievedResults []*knowledge.RetrievalResult
	if err != nil {
		retrievedResults = []*knowledge.RetrievalResult{}
	} else {
		retrievedResults = retrievalResult.Results
	}

	// Build context and sources from hybrid retrieval results
	contexts := make([]RetrievalContext, 0, len(retrievedResults))
	sources := make([]chat.Source, 0, len(retrievedResults))
	totalTokens := 0

	for i, result := range retrievedResults {
		if totalTokens >= s.maxContextTokens {
			break
		}

		chunk := result.Chunk
		// Truncate if needed
		content := chunk.Content
		if totalTokens+chunk.TokenCount > s.maxContextTokens {
			remainingTokens := s.maxContextTokens - totalTokens
			content = TruncateContent(content, remainingTokens)
		}

		contexts = append(contexts, RetrievalContext{
			SourceIndex: i + 1,
			ItemTitle:   chunk.Heading,
			Content:     content,
			Snippet:     BuildSourceSnippet(content, 200),
		})

		sources = append(sources, chat.Source{
			ChunkID:   chunk.ID,
			ItemID:    chunk.ItemVersionID,
			ItemTitle: chunk.Heading,
			Snippet:   BuildSourceSnippet(chunk.Content, 200),
			Score:     result.FinalScore, // Use the hybrid retrieval score
		})

		totalTokens += chunk.TokenCount
	}

	// Build messages for LLM using memory service
	contextPrompt := BuildContextPrompt(contexts)
	var messages []openai.ChatMessage
	if s.enableMemory {
		messages = s.memoryService.BuildMessagesForLLM(
			SystemPrompt,
			contextPrompt,
			conversationContext,
			input.Query,
		)
	} else {
		// Legacy behavior: get last 10 messages
		history, _ := s.messageRepo.GetLastMessages(input.ConversationID, 10)
		historyMessages := make([]openai.ChatMessage, 0, len(history))
		for _, msg := range history {
			if msg.ID == userMessage.ID {
				continue
			}
			historyMessages = append(historyMessages, openai.ChatMessage{
				Role:    string(msg.Role),
				Content: msg.Content,
			})
		}
		messages = openai.BuildRAGMessages(
			SystemPrompt,
			contextPrompt,
			input.Query,
			historyMessages,
		)
	}

	// Generate response
	var responseContent string
	var tokenCount int

	if len(retrievedResults) == 0 {
		// No context found, return helpful message
		responseContent = NoContextResponse
		tokenCount = 0
	} else {
		// Call LLM
		resp, err := s.chatClient.CreateCompletion(ctx, messages, 2000)
		if err != nil {
			return nil, err
		}

		if len(resp.Choices) > 0 {
			responseContent = resp.Choices[0].Message.Content
		} else {
			responseContent = "I was unable to generate a response. Please try again."
		}
		tokenCount = resp.Usage.TotalTokens
	}

	latencyMs := int(time.Since(start).Milliseconds())

	// Create assistant message
	assistantMessage := chat.NewAssistantMessage(
		input.ConversationID,
		input.WorkspaceID,
		responseContent,
		sources,
		s.chatClient.GetModel(),
		tokenCount,
		latencyMs,
	)
	if err := s.messageRepo.Create(assistantMessage); err != nil {
		return nil, err
	}

	// Increment message count for assistant message
	conversation.IncrementMessageCount()

	// Check if we need to update the summary (async, non-blocking)
	if s.enableMemory && conversationContext.NeedsSummaryUpdate {
		go func() {
			bgCtx := context.Background()
			summary, lastMsgID, err := s.memoryService.GenerateSummary(bgCtx, conversation, conversation.Summary)
			if err == nil && summary != "" {
				conversation.UpdateSummary(summary, lastMsgID)
				s.conversationRepo.Update(conversation)
			}
		}()
	}

	// Update conversation
	conversation.Touch()
	if err := s.conversationRepo.Update(conversation); err != nil {
		// Non-fatal
	}

	return &SendMessageResult{
		UserMessage:      userMessage,
		AssistantMessage: assistantMessage,
		Sources:          sources,
	}, nil
}

// StreamHandler is a callback for streaming responses
type StreamHandler func(token string, done bool) error

// SendMessageStream sends a message with streaming response
func (s *Service) SendMessageStream(
	ctx context.Context,
	input SendMessageInput,
	handler StreamHandler,
) (*SendMessageResult, error) {
	if input.Query == "" {
		return nil, errors.New("query cannot be empty")
	}

	start := time.Now()

	// Create user message
	userMessage := chat.NewUserMessage(
		input.ConversationID,
		input.WorkspaceID,
		input.Query,
	)
	if err := s.messageRepo.Create(userMessage); err != nil {
		return nil, err
	}

	// Get conversation and update title if first message
	conversation, err := s.conversationRepo.GetByID(input.ConversationID)
	if err != nil {
		return nil, err
	}

	// Increment message count
	conversation.IncrementMessageCount()

	if conversation.Title == "New Conversation" {
		conversation.GenerateTitleFromQuery(input.Query)
	}

	// Get conversation context (summary + sliding window)
	var conversationContext *ConversationContext
	if s.enableMemory {
		conversationContext, err = s.memoryService.GetConversationContext(ctx, conversation, userMessage.ID)
		if err != nil {
			conversationContext = &ConversationContext{}
		}
	} else {
		conversationContext = &ConversationContext{}
	}

	// Determine the query to use for retrieval
	retrievalQuery := input.Query
	if s.enableQueryEnhance && (conversationContext.Summary != "" || len(conversationContext.RecentMessages) > 0) {
		enhancedQuery, err := s.memoryService.EnhanceQueryWithContext(ctx, input.Query, conversationContext)
		if err == nil && enhancedQuery != "" {
			retrievalQuery = enhancedQuery
		}
	}

	// Retrieve relevant chunks using hybrid retrieval
	retrievalResult, err := s.retrievalService.Retrieve(ctx, retrieval.RetrieveInput{
		Query:       retrievalQuery,
		WorkspaceID: input.WorkspaceID,
		MaxResults:  s.maxChunks,
		MaxTokens:   s.maxContextTokens,
	})

	var retrievedResults []*knowledge.RetrievalResult
	if err != nil {
		retrievedResults = []*knowledge.RetrievalResult{}
	} else {
		retrievedResults = retrievalResult.Results
	}

	// Build context and sources from hybrid retrieval results
	contexts := make([]RetrievalContext, 0, len(retrievedResults))
	sources := make([]chat.Source, 0, len(retrievedResults))
	totalTokens := 0

	for i, result := range retrievedResults {
		if totalTokens >= s.maxContextTokens {
			break
		}

		chunk := result.Chunk
		content := chunk.Content
		if totalTokens+chunk.TokenCount > s.maxContextTokens {
			remainingTokens := s.maxContextTokens - totalTokens
			content = TruncateContent(content, remainingTokens)
		}

		contexts = append(contexts, RetrievalContext{
			SourceIndex: i + 1,
			ItemTitle:   chunk.Heading,
			Content:     content,
		})

		sources = append(sources, chat.Source{
			ChunkID:   chunk.ID,
			ItemID:    chunk.ItemVersionID,
			ItemTitle: chunk.Heading,
			Snippet:   BuildSourceSnippet(chunk.Content, 200),
			Score:     result.FinalScore,
		})

		totalTokens += chunk.TokenCount
	}

	var responseContent string
	var tokenCount int

	if len(retrievedResults) == 0 {
		responseContent = NoContextResponse
		handler(responseContent, true)
	} else {
		contextPrompt := BuildContextPrompt(contexts)

		// Build messages using memory service
		var messages []openai.ChatMessage
		if s.enableMemory {
			messages = s.memoryService.BuildMessagesForLLM(
				SystemPrompt,
				contextPrompt,
				conversationContext,
				input.Query,
			)
		} else {
			history, _ := s.messageRepo.GetLastMessages(input.ConversationID, 10)
			historyMessages := make([]openai.ChatMessage, 0, len(history))
			for _, msg := range history {
				if msg.ID == userMessage.ID {
					continue
				}
				historyMessages = append(historyMessages, openai.ChatMessage{
					Role:    string(msg.Role),
					Content: msg.Content,
				})
			}
			messages = openai.BuildRAGMessages(
				SystemPrompt,
				contextPrompt,
				input.Query,
				historyMessages,
			)
		}

		var fullResponse string
		usage, err := s.chatClient.CreateCompletionStream(ctx, messages, 2000, func(token string, done bool) error {
			fullResponse += token
			return handler(token, done)
		})
		if err != nil {
			return nil, err
		}

		responseContent = fullResponse
		if usage != nil {
			tokenCount = usage.TotalTokens
		}
	}

	latencyMs := int(time.Since(start).Milliseconds())

	// Create assistant message
	assistantMessage := chat.NewAssistantMessage(
		input.ConversationID,
		input.WorkspaceID,
		responseContent,
		sources,
		s.chatClient.GetModel(),
		tokenCount,
		latencyMs,
	)
	if err := s.messageRepo.Create(assistantMessage); err != nil {
		return nil, err
	}

	// Increment message count for assistant message
	conversation.IncrementMessageCount()

	// Check if we need to update the summary (async, non-blocking)
	if s.enableMemory && conversationContext.NeedsSummaryUpdate {
		go func() {
			bgCtx := context.Background()
			summary, lastMsgID, err := s.memoryService.GenerateSummary(bgCtx, conversation, conversation.Summary)
			if err == nil && summary != "" {
				conversation.UpdateSummary(summary, lastMsgID)
				s.conversationRepo.Update(conversation)
			}
		}()
	}

	conversation.Touch()
	s.conversationRepo.Update(conversation)

	return &SendMessageResult{
		UserMessage:      userMessage,
		AssistantMessage: assistantMessage,
		Sources:          sources,
	}, nil
}

// DeleteConversation soft-deletes a conversation
func (s *Service) DeleteConversation(ctx context.Context, id common.ID) error {
	return s.conversationRepo.Delete(id)
}

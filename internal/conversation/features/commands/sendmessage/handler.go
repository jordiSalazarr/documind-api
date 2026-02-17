package sendmessage

import (
	"context"
	"errors"
	"time"

	chatDomain "documind.jordi.org/internal/conversation/domain"
	"documind.jordi.org/internal/shared/infrastructure/openai"
)

// lowCoverageDisclaimer is appended when citation coverage falls below 50%.
const lowCoverageDisclaimer = "\n\n---\n*Note: Parts of this answer may not be directly supported by the available documents.*"

// CitationCoverageChecker checks citation coverage in LLM responses.
type CitationCoverageChecker interface {
	CheckCoverage(answer string, sourceCount int) CitationCoverageResult
}

// CitationCoverageResult contains the result of a citation coverage check.
type CitationCoverageResult struct {
	CoverageRatio    float64
	TotalSentences   int
	CitedSentences   int
	UncitedSentences int
	CitedSources     []int
}

// Handler handles sending a user message and generating an AI response
type Handler struct {
	convReader      ConversationReader
	convWriter      ConversationWriter
	msgReader       MessageReader
	msgWriter       MessageWriter
	retriever       Retriever
	chatClient      ChatClient
	memoryService   MemoryService
	citationChecker CitationCoverageChecker
	config          Config
}

// NewHandler creates a new send message handler
func NewHandler(
	convReader ConversationReader,
	convWriter ConversationWriter,
	msgReader MessageReader,
	msgWriter MessageWriter,
	retriever Retriever,
	chatClient ChatClient,
	memoryService MemoryService,
	config Config,
) *Handler {
	if config.MaxContextTokens <= 0 {
		config.MaxContextTokens = 4000
	}
	if config.MaxChunks <= 0 {
		config.MaxChunks = 5
	}

	return &Handler{
		convReader:    convReader,
		convWriter:    convWriter,
		msgReader:     msgReader,
		msgWriter:     msgWriter,
		retriever:     retriever,
		chatClient:    chatClient,
		memoryService: memoryService,
		config:        config,
	}
}

// SetCitationChecker sets an optional citation coverage checker.
func (h *Handler) SetCitationChecker(checker CitationCoverageChecker) {
	h.citationChecker = checker
}

// Handle sends a user message and generates a blocking AI response
func (h *Handler) Handle(ctx context.Context, cmd Command) (*Result, error) {
	if cmd.Query == "" {
		return nil, errors.New("query cannot be empty")
	}

	start := time.Now()

	// Create user message
	userMessage := chatDomain.NewUserMessage(
		cmd.ConversationID,
		cmd.WorkspaceID,
		cmd.Query,
	)
	if err := h.msgWriter.Create(userMessage); err != nil {
		return nil, err
	}

	// Get conversation and update title if first message
	conversation, err := h.convReader.GetByID(cmd.ConversationID)
	if err != nil {
		return nil, err
	}

	// Increment message count
	conversation.IncrementMessageCount()

	// Update title from first query if still default
	if conversation.Title == "New Conversation" {
		conversation.GenerateTitleFromQuery(cmd.Query)
	}

	// Get conversation context (summary + sliding window)
	var conversationContext *ConversationContext
	if h.config.EnableMemory {
		conversationContext, err = h.memoryService.GetConversationContext(ctx, conversation, userMessage.ID)
		if err != nil {
			// Fall back to empty context on error
			conversationContext = &ConversationContext{}
		}
	} else {
		conversationContext = &ConversationContext{}
	}

	// Determine the query to use for retrieval
	retrievalQuery := cmd.Query
	if h.config.EnableQueryEnhance && (conversationContext.Summary != "" || len(conversationContext.RecentMessages) > 0) {
		enhancedQuery, err := h.memoryService.EnhanceQueryWithContext(ctx, cmd.Query, conversationContext)
		if err == nil && enhancedQuery != "" {
			retrievalQuery = enhancedQuery
		}
	}

	// Retrieve relevant chunks using hybrid retrieval with potentially enhanced query
	retrievalResult, err := h.retriever.Retrieve(ctx, RetrievalInput{
		Query:       retrievalQuery,
		WorkspaceID: cmd.WorkspaceID,
		MaxResults:  h.config.MaxChunks,
		MaxTokens:   h.config.MaxContextTokens,
	})

	var retrievedResults []*RetrievalResult
	if err != nil {
		retrievedResults = []*RetrievalResult{}
	} else {
		retrievedResults = retrievalResult.Results
	}

	// Build context and sources from hybrid retrieval results
	contexts := make([]RetrievalContext, 0, len(retrievedResults))
	sources := make([]chatDomain.Source, 0, len(retrievedResults))
	totalTokens := 0

	for i, result := range retrievedResults {
		if totalTokens >= h.config.MaxContextTokens {
			break
		}

		// Truncate if needed
		content := result.Content
		if totalTokens+result.TokenCount > h.config.MaxContextTokens {
			remainingTokens := h.config.MaxContextTokens - totalTokens
			content = truncateContent(content, remainingTokens)
		}

		contexts = append(contexts, RetrievalContext{
			SourceIndex: i + 1,
			ItemTitle:   result.Heading,
			Content:     content,
			Snippet:     buildSourceSnippet(content, 200),
		})

		sources = append(sources, chatDomain.Source{
			ChunkID:   result.ChunkID,
			ItemID:    result.ItemVersionID,
			ItemTitle: result.Heading,
			Snippet:   buildSourceSnippet(result.Content, 200),
			Score:     result.FinalScore,
		})

		totalTokens += result.TokenCount
	}

	// Build messages for LLM using memory service
	contextPrompt := buildContextPrompt(contexts)
	var messages []openai.ChatMessage
	if h.config.EnableMemory {
		messages = h.memoryService.BuildMessagesForLLM(
			systemPrompt,
			contextPrompt,
			conversationContext,
			cmd.Query,
		)
	} else {
		// Legacy behavior: get last 10 messages
		history, _ := h.msgReader.GetLastMessages(cmd.ConversationID, 10)
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
			systemPrompt,
			contextPrompt,
			cmd.Query,
			historyMessages,
		)
	}

	// Generate response
	var responseContent string
	var tokenCount int

	if len(retrievedResults) == 0 {
		// No context found, return helpful message
		responseContent = noContextResponse
		tokenCount = 0
	} else {
		// Call LLM
		resp, err := h.chatClient.CreateCompletion(ctx, messages, 2000)
		if err != nil {
			return nil, err
		}

		if len(resp.Choices) > 0 {
			responseContent = resp.Choices[0].Message.Content
		} else {
			responseContent = "I was unable to generate a response. Please try again."
		}
		tokenCount = resp.Usage.TotalTokens

		// Check citation coverage and append disclaimer if low
		if h.citationChecker != nil && responseContent != "" {
			coverage := h.citationChecker.CheckCoverage(responseContent, len(sources))
			if coverage.CoverageRatio < 0.5 && coverage.TotalSentences > 0 {
				responseContent += lowCoverageDisclaimer
			}
		}
	}

	latencyMs := int(time.Since(start).Milliseconds())

	// Create assistant message
	assistantMessage := chatDomain.NewAssistantMessage(
		cmd.ConversationID,
		cmd.WorkspaceID,
		responseContent,
		sources,
		h.chatClient.GetModel(),
		tokenCount,
		latencyMs,
	)
	if err := h.msgWriter.Create(assistantMessage); err != nil {
		return nil, err
	}

	// Increment message count for assistant message
	conversation.IncrementMessageCount()

	// Check if we need to update the summary (async, non-blocking)
	if h.config.EnableMemory && conversationContext.NeedsSummaryUpdate {
		go func() {
			bgCtx := context.Background()
			summary, lastMsgID, err := h.memoryService.GenerateSummary(bgCtx, conversation, conversation.Summary)
			if err == nil && summary != "" {
				conversation.UpdateSummary(summary, lastMsgID)
				_ = h.convWriter.Update(conversation)
			}
		}()
	}

	// Update conversation
	conversation.Touch()
	_ = h.convWriter.Update(conversation)

	return &Result{
		UserMessage:      userMessage,
		AssistantMessage: assistantMessage,
		Sources:          sources,
	}, nil
}

// HandleStream sends a message with streaming response
func (h *Handler) HandleStream(
	ctx context.Context,
	cmd Command,
	handler StreamCallback,
) (*Result, error) {
	if cmd.Query == "" {
		return nil, errors.New("query cannot be empty")
	}

	start := time.Now()

	// Create user message
	userMessage := chatDomain.NewUserMessage(
		cmd.ConversationID,
		cmd.WorkspaceID,
		cmd.Query,
	)
	if err := h.msgWriter.Create(userMessage); err != nil {
		return nil, err
	}

	// Get conversation and update title if first message
	conversation, err := h.convReader.GetByID(cmd.ConversationID)
	if err != nil {
		return nil, err
	}

	// Increment message count
	conversation.IncrementMessageCount()

	if conversation.Title == "New Conversation" {
		conversation.GenerateTitleFromQuery(cmd.Query)
	}

	// Get conversation context (summary + sliding window)
	var conversationContext *ConversationContext
	if h.config.EnableMemory {
		conversationContext, err = h.memoryService.GetConversationContext(ctx, conversation, userMessage.ID)
		if err != nil {
			conversationContext = &ConversationContext{}
		}
	} else {
		conversationContext = &ConversationContext{}
	}

	// Determine the query to use for retrieval
	retrievalQuery := cmd.Query
	if h.config.EnableQueryEnhance && (conversationContext.Summary != "" || len(conversationContext.RecentMessages) > 0) {
		enhancedQuery, err := h.memoryService.EnhanceQueryWithContext(ctx, cmd.Query, conversationContext)
		if err == nil && enhancedQuery != "" {
			retrievalQuery = enhancedQuery
		}
	}

	// Retrieve relevant chunks using hybrid retrieval
	retrievalResult, err := h.retriever.Retrieve(ctx, RetrievalInput{
		Query:       retrievalQuery,
		WorkspaceID: cmd.WorkspaceID,
		MaxResults:  h.config.MaxChunks,
		MaxTokens:   h.config.MaxContextTokens,
	})

	var retrievedResults []*RetrievalResult
	if err != nil {
		retrievedResults = []*RetrievalResult{}
	} else {
		retrievedResults = retrievalResult.Results
	}

	// Build context and sources from hybrid retrieval results
	contexts := make([]RetrievalContext, 0, len(retrievedResults))
	sources := make([]chatDomain.Source, 0, len(retrievedResults))
	totalTokens := 0

	for i, result := range retrievedResults {
		if totalTokens >= h.config.MaxContextTokens {
			break
		}

		content := result.Content
		if totalTokens+result.TokenCount > h.config.MaxContextTokens {
			remainingTokens := h.config.MaxContextTokens - totalTokens
			content = truncateContent(content, remainingTokens)
		}

		contexts = append(contexts, RetrievalContext{
			SourceIndex: i + 1,
			ItemTitle:   result.Heading,
			Content:     content,
		})

		sources = append(sources, chatDomain.Source{
			ChunkID:   result.ChunkID,
			ItemID:    result.ItemVersionID,
			ItemTitle: result.Heading,
			Snippet:   buildSourceSnippet(result.Content, 200),
			Score:     result.FinalScore,
		})

		totalTokens += result.TokenCount
	}

	var responseContent string
	var tokenCount int

	if len(retrievedResults) == 0 {
		responseContent = noContextResponse
		_ = handler(responseContent, true)
	} else {
		contextPrompt := buildContextPrompt(contexts)

		// Build messages using memory service
		var messages []openai.ChatMessage
		if h.config.EnableMemory {
			messages = h.memoryService.BuildMessagesForLLM(
				systemPrompt,
				contextPrompt,
				conversationContext,
				cmd.Query,
			)
		} else {
			history, _ := h.msgReader.GetLastMessages(cmd.ConversationID, 10)
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
				systemPrompt,
				contextPrompt,
				cmd.Query,
				historyMessages,
			)
		}

		var fullResponse string
		usage, err := h.chatClient.CreateCompletionStream(ctx, messages, 2000, func(token string, done bool) error {
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

		// Check citation coverage for stored message (disclaimer streamed separately if needed)
		if h.citationChecker != nil && responseContent != "" {
			coverage := h.citationChecker.CheckCoverage(responseContent, len(sources))
			if coverage.CoverageRatio < 0.5 && coverage.TotalSentences > 0 {
				responseContent += lowCoverageDisclaimer
				_ = handler(lowCoverageDisclaimer, false)
			}
		}
	}

	latencyMs := int(time.Since(start).Milliseconds())

	// Create assistant message
	assistantMessage := chatDomain.NewAssistantMessage(
		cmd.ConversationID,
		cmd.WorkspaceID,
		responseContent,
		sources,
		h.chatClient.GetModel(),
		tokenCount,
		latencyMs,
	)
	if err := h.msgWriter.Create(assistantMessage); err != nil {
		return nil, err
	}

	// Increment message count for assistant message
	conversation.IncrementMessageCount()

	// Check if we need to update the summary (async, non-blocking)
	if h.config.EnableMemory && conversationContext.NeedsSummaryUpdate {
		go func() {
			bgCtx := context.Background()
			summary, lastMsgID, err := h.memoryService.GenerateSummary(bgCtx, conversation, conversation.Summary)
			if err == nil && summary != "" {
				conversation.UpdateSummary(summary, lastMsgID)
				_ = h.convWriter.Update(conversation)
			}
		}()
	}

	conversation.Touch()
	_ = h.convWriter.Update(conversation)

	return &Result{
		UserMessage:      userMessage,
		AssistantMessage: assistantMessage,
		Sources:          sources,
	}, nil
}

package sendmessage

import (
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"strings"

	chatDomain "documind.jordi.org/internal/conversation/domain"
	sharedDomain "documind.jordi.org/internal/shared/domain"
	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

// --- Prompts and helpers (local to avoid circular imports with parent chat package) ---

// systemPrompt is the base system prompt for RAG with few-shot examples and chain-of-thought
const systemPrompt = `You are a helpful AI assistant for docuMind, a knowledge management system.

RULES:
1. Answer questions based ONLY on the provided context documents.
2. Cite every factual claim using [Source N] notation.
3. NEVER make up information that is not in the provided context.
4. If the context doesn't contain enough information, say so clearly.
5. DO NOT copy-paste entire documents. Synthesize information into focused answers.
6. Use markdown formatting for clarity when helpful.

EXAMPLES:

Example 1 — Factual question with citation:
Q: How do I configure JWT tokens?
A: JWT tokens are configured by setting the ` + "`JWT_SECRET`" + ` environment variable to a minimum 32-character string [Source 1]. The default expiry is 72 hours, configurable via ` + "`JWT_EXPIRY_HOURS`" + ` [Source 1]. You should also ensure the secret is stored securely and rotated periodically [Source 2].

Example 2 — Insufficient context:
Q: What is the pricing for the enterprise plan?
A: I don't have enough information in the provided documents to answer this question. The available sources cover technical setup and configuration but don't include pricing details. You may want to check the pricing page or contact the sales team.

COMPLEX QUERIES:
For complex questions, think step-by-step:
1. Identify the key concepts in the question
2. Find relevant information across all provided sources
3. Synthesize a complete answer with citations for each claim
4. Verify every claim is supported by a source before including it

If the query is complex, provide a thorough answer covering all aspects. Prioritize accuracy over brevity — every statement must be backed by a source.`

// noContextResponse is returned when no relevant documents are found
const noContextResponse = `I couldn't find any relevant documents in the knowledge base to answer your question.

You might want to:
1. Try rephrasing your question with different keywords
2. Check if the relevant documentation has been uploaded to this service
3. Ask a more specific question related to the available content`

// buildContextPrompt builds the context section for the prompt with clear delimiters
func buildContextPrompt(contexts []RetrievalContext) string {
	if len(contexts) == 0 {
		return "No relevant documents were found in the knowledge base."
	}

	var builder strings.Builder
	builder.WriteString("=== RETRIEVED DOCUMENTS ===\n\n")

	for _, ctx := range contexts {
		builder.WriteString(fmt.Sprintf("--- [Source %d]: %s ---\n", ctx.SourceIndex, ctx.ItemTitle))
		builder.WriteString(ctx.Content)
		builder.WriteString("\n--- END Source ---\n\n")
	}

	builder.WriteString("=== END DOCUMENTS ===\n\n")
	builder.WriteString("Use the sources above to answer the user's question. ")
	builder.WriteString("Cite each fact with [Source N]. If no source supports a claim, do not include it.")

	return builder.String()
}

// truncateContent truncates content to fit within token limits
// Approximate: 1 token ~ 4 characters
func truncateContent(content string, maxTokens int) string {
	maxChars := maxTokens * 4
	if len(content) <= maxChars {
		return content
	}
	return content[:maxChars-3] + "..."
}

// buildSourceSnippet extracts a relevant snippet from content
func buildSourceSnippet(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}
	// Try to break at a sentence boundary
	truncated := content[:maxLength]
	lastPeriod := strings.LastIndex(truncated, ". ")
	if lastPeriod > maxLength/2 {
		return truncated[:lastPeriod+1]
	}
	return truncated + "..."
}

// --- HTTP ---

// SendMessageRequest represents the request to send a message
type SendMessageRequest struct {
	ConversationID string `json:"conversation_id" binding:"required"`
	ServiceID      string `json:"service_id" binding:"required"`
	Query          string `json:"query" binding:"required"`
	Stream         bool   `json:"stream"`
}

// MessageResponse represents a message in API responses
type MessageResponse struct {
	ID             string           `json:"id"`
	ConversationID string           `json:"conversation_id"`
	Role           string           `json:"role"`
	Content        string           `json:"content"`
	Sources        []SourceResponse `json:"sources,omitempty"`
	TokenCount     int              `json:"token_count,omitempty"`
	Model          string           `json:"model,omitempty"`
	LatencyMs      int              `json:"latency_ms,omitempty"`
	CreatedAt      string           `json:"created_at"`
}

// SourceResponse represents a source citation in API responses
type SourceResponse struct {
	ChunkID   string  `json:"chunk_id"`
	ItemID    string  `json:"item_id"`
	ItemTitle string  `json:"item_title"`
	Snippet   string  `json:"snippet"`
	Score     float64 `json:"score"`
}

// Endpoint returns a gin handler for blocking message send
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SendMessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(nethttp.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cmd := Command{
			ConversationID: sharedDomain.ID(req.ConversationID),
			WorkspaceID:    sharedDomain.ID(middleware.GetWorkspaceID(c)),
			ServiceID:      sharedDomain.ID(req.ServiceID),
			Query:          req.Query,
			UserID:         sharedDomain.ID(middleware.GetUserID(c)),
		}

		result, err := h.Handle(c.Request.Context(), cmd)
		if err != nil {
			c.JSON(nethttp.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(nethttp.StatusOK, gin.H{
			"user_message":      toMessageResponse(result.UserMessage),
			"assistant_message": toMessageResponse(result.AssistantMessage),
			"sources":           toSourcesResponse(result.Sources),
		})
	}
}

// StreamEndpoint returns a gin handler for streaming message send via SSE
func StreamEndpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SendMessageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(nethttp.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cmd := Command{
			ConversationID: sharedDomain.ID(req.ConversationID),
			WorkspaceID:    sharedDomain.ID(middleware.GetWorkspaceID(c)),
			ServiceID:      sharedDomain.ID(req.ServiceID),
			Query:          req.Query,
			UserID:         sharedDomain.ID(middleware.GetUserID(c)),
		}

		// Set headers for SSE
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Transfer-Encoding", "chunked")

		// Use a channel to receive streaming tokens
		flusher, ok := c.Writer.(nethttp.Flusher)
		if !ok {
			c.JSON(nethttp.StatusInternalServerError, gin.H{"error": "streaming not supported"})
			return
		}

		// Send the user message event first
		sendSSEEvent(c.Writer, flusher, "user_message", gin.H{
			"id":      sharedDomain.NewID(),
			"content": cmd.Query,
			"role":    "user",
		})

		result, err := h.HandleStream(c.Request.Context(), cmd, func(token string, done bool) error {
			if done {
				return nil
			}
			// Send token event
			sendSSEEvent(c.Writer, flusher, "token", gin.H{
				"delta": token,
			})
			return nil
		})

		if err != nil {
			sendSSEEvent(c.Writer, flusher, "error", gin.H{
				"error": err.Error(),
			})
			return
		}

		// Send sources
		sendSSEEvent(c.Writer, flusher, "sources", gin.H{
			"sources": toSourcesResponse(result.Sources),
		})

		// Send complete message
		sendSSEEvent(c.Writer, flusher, "done", gin.H{
			"message": toMessageResponse(result.AssistantMessage),
		})
	}
}

// --- Helper functions ---

func toMessageResponse(msg *chatDomain.Message) MessageResponse {
	return MessageResponse{
		ID:             string(msg.ID),
		ConversationID: string(msg.ConversationID),
		Role:           string(msg.Role),
		Content:        msg.Content,
		Sources:        toSourcesResponse(msg.Sources),
		TokenCount:     msg.TokenCount,
		Model:          msg.Model,
		LatencyMs:      msg.LatencyMs,
		CreatedAt:      msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toSourcesResponse(sources []chatDomain.Source) []SourceResponse {
	response := make([]SourceResponse, len(sources))
	for i, src := range sources {
		response[i] = SourceResponse{
			ChunkID:   string(src.ChunkID),
			ItemID:    string(src.ItemID),
			ItemTitle: src.ItemTitle,
			Snippet:   src.Snippet,
			Score:     src.Score,
		}
	}
	return response
}

func sendSSEEvent(w nethttp.ResponseWriter, flusher nethttp.Flusher, event string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	flusher.Flush()
}

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"documind.jordi.org/internal/application/chat"
	chatDomain "documind.jordi.org/internal/domain/chat"
	"documind.jordi.org/internal/domain/common"
	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	service *chat.Service
}

func NewChatHandler(service *chat.Service) *ChatHandler {
	return &ChatHandler{service: service}
}

// ConversationResponse represents a conversation in API responses
type ConversationResponse struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspace_id"`
	ServiceID   string `json:"service_id"`
	UserID      string `json:"user_id"`
	Title       string `json:"title"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
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

// CreateConversationRequest represents the request to create a conversation
type CreateConversationRequest struct {
	WorkspaceID string `json:"workspace_id" binding:"required"`
	ServiceID   string `json:"service_id" binding:"required"`
}

// CreateConversation creates a new conversation
func (h *ChatHandler) CreateConversation(c *gin.Context) {
	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get actual user ID from auth context
	userID := common.NewID()

	input := chat.CreateConversationInput{
		WorkspaceID: common.ID(req.WorkspaceID),
		ServiceID:   common.ID(req.ServiceID),
		UserID:      userID,
	}

	conversation, err := h.service.CreateConversation(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toConversationResponse(conversation))
}

// GetConversation retrieves a conversation by ID
func (h *ChatHandler) GetConversation(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	conversation, err := h.service.GetConversation(c.Request.Context(), common.ID(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toConversationResponse(conversation))
}

// ListConversations lists conversations for a service
func (h *ChatHandler) ListConversations(c *gin.Context) {
	serviceID := c.Query("service_id")
	if serviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service_id is required"})
		return
	}

	limit := parseIntParam(c.Query("limit"), 20)
	offset := parseIntParam(c.Query("offset"), 0)

	conversations, err := h.service.ListConversations(
		c.Request.Context(),
		common.ID(serviceID),
		limit,
		offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]ConversationResponse, len(conversations))
	for i, conv := range conversations {
		response[i] = toConversationResponse(conv)
	}

	c.JSON(http.StatusOK, gin.H{
		"conversations": response,
		"limit":         limit,
		"offset":        offset,
	})
}

// GetMessages retrieves messages for a conversation
func (h *ChatHandler) GetMessages(c *gin.Context) {
	conversationID := c.Param("id")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	limit := parseIntParam(c.Query("limit"), 50)
	offset := parseIntParam(c.Query("offset"), 0)

	messages, err := h.service.GetMessages(
		c.Request.Context(),
		common.ID(conversationID),
		limit,
		offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]MessageResponse, len(messages))
	for i, msg := range messages {
		response[i] = toMessageResponse(msg)
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": response,
		"limit":    limit,
		"offset":   offset,
	})
}

// SendMessageRequest represents the request to send a message
type SendMessageRequest struct {
	ConversationID string `json:"conversation_id" binding:"required"`
	WorkspaceID    string `json:"workspace_id" binding:"required"`
	ServiceID      string `json:"service_id" binding:"required"`
	Query          string `json:"query" binding:"required"`
	Stream         bool   `json:"stream"`
}

// SendMessage sends a user message and gets an AI response
func (h *ChatHandler) SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get actual user ID from auth context
	userID := common.NewID()

	input := chat.SendMessageInput{
		ConversationID: common.ID(req.ConversationID),
		WorkspaceID:    common.ID(req.WorkspaceID),
		ServiceID:      common.ID(req.ServiceID),
		Query:          req.Query,
		UserID:         userID,
	}

	if req.Stream {
		h.sendMessageStream(c, input)
	} else {
		h.sendMessageBlocking(c, input)
	}
}

// sendMessageBlocking sends a message and waits for full response
func (h *ChatHandler) sendMessageBlocking(c *gin.Context, input chat.SendMessageInput) {
	result, err := h.service.SendMessage(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_message":      toMessageResponse(result.UserMessage),
		"assistant_message": toMessageResponse(result.AssistantMessage),
		"sources":           toSourcesResponse(result.Sources),
	})
}

// sendMessageStream sends a message with streaming response via SSE
func (h *ChatHandler) sendMessageStream(c *gin.Context, input chat.SendMessageInput) {
	// Set headers for SSE
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// Use a channel to receive streaming tokens
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	// Send the user message event first
	sendSSEEvent(c.Writer, flusher, "user_message", gin.H{
		"id":      common.NewID(),
		"content": input.Query,
		"role":    "user",
	})

	result, err := h.service.SendMessageStream(c.Request.Context(), input, func(token string, done bool) error {
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

// DeleteConversation deletes a conversation
func (h *ChatHandler) DeleteConversation(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	err := h.service.DeleteConversation(c.Request.Context(), common.ID(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "conversation deleted"})
}

// Helper functions

func toConversationResponse(conv *chatDomain.Conversation) ConversationResponse {
	return ConversationResponse{
		ID:          string(conv.ID),
		WorkspaceID: string(conv.WorkspaceID),
		ServiceID:   string(conv.ServiceID),
		UserID:      string(conv.UserID),
		Title:       conv.Title,
		CreatedAt:   conv.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   conv.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

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

func parseIntParam(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}

func sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, event string, data interface{}) {
	jsonData, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	flusher.Flush()
}

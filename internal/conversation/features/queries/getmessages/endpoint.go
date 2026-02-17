package getmessages

import (
	nethttp "net/http"
	"strconv"

	chatDomain "documind.jordi.org/internal/conversation/domain"
	sharedDomain "documind.jordi.org/internal/shared/domain"
	"github.com/gin-gonic/gin"
)

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

// Endpoint returns a gin handler for retrieving messages
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationID := c.Param("id")
		if conversationID == "" {
			c.JSON(nethttp.StatusBadRequest, gin.H{"error": "id is required"})
			return
		}

		limit := parseIntParam(c.Query("limit"), 50)
		offset := parseIntParam(c.Query("offset"), 0)

		messages, err := h.Handle(
			c.Request.Context(),
			sharedDomain.ID(conversationID),
			limit,
			offset,
		)
		if err != nil {
			c.JSON(nethttp.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response := make([]MessageResponse, len(messages))
		for i, msg := range messages {
			response[i] = toMessageResponse(msg)
		}

		c.JSON(nethttp.StatusOK, gin.H{
			"messages": response,
			"limit":    limit,
			"offset":   offset,
		})
	}
}

func toMessageResponse(msg *chatDomain.Message) MessageResponse {
	sources := make([]SourceResponse, len(msg.Sources))
	for i, src := range msg.Sources {
		sources[i] = SourceResponse{
			ChunkID:   string(src.ChunkID),
			ItemID:    string(src.ItemID),
			ItemTitle: src.ItemTitle,
			Snippet:   src.Snippet,
			Score:     src.Score,
		}
	}
	return MessageResponse{
		ID:             string(msg.ID),
		ConversationID: string(msg.ConversationID),
		Role:           string(msg.Role),
		Content:        msg.Content,
		Sources:        sources,
		TokenCount:     msg.TokenCount,
		Model:          msg.Model,
		LatencyMs:      msg.LatencyMs,
		CreatedAt:      msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
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

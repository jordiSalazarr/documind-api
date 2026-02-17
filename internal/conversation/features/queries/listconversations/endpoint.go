package listconversations

import (
	nethttp "net/http"
	"strconv"

	sharedDomain "documind.jordi.org/internal/shared/domain"
	"github.com/gin-gonic/gin"
)

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

// Endpoint returns a gin handler for listing conversations
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceID := c.Query("service_id")
		if serviceID == "" {
			c.JSON(nethttp.StatusBadRequest, gin.H{"error": "service_id is required"})
			return
		}

		limit := parseIntParam(c.Query("limit"), 20)
		offset := parseIntParam(c.Query("offset"), 0)

		conversations, err := h.Handle(
			c.Request.Context(),
			sharedDomain.ID(serviceID),
			limit,
			offset,
		)
		if err != nil {
			c.JSON(nethttp.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response := make([]ConversationResponse, len(conversations))
		for i, conv := range conversations {
			response[i] = ConversationResponse{
				ID:          string(conv.ID),
				WorkspaceID: string(conv.WorkspaceID),
				ServiceID:   string(conv.ServiceID),
				UserID:      string(conv.UserID),
				Title:       conv.Title,
				CreatedAt:   conv.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				UpdatedAt:   conv.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
		}

		c.JSON(nethttp.StatusOK, gin.H{
			"conversations": response,
			"limit":         limit,
			"offset":        offset,
		})
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

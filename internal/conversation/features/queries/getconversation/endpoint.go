package getconversation

import (
	nethttp "net/http"

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

// Endpoint returns a gin handler for retrieving a conversation
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(nethttp.StatusBadRequest, gin.H{"error": "id is required"})
			return
		}

		conversation, err := h.Handle(c.Request.Context(), sharedDomain.ID(id))
		if err != nil {
			c.JSON(nethttp.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(nethttp.StatusOK, ConversationResponse{
			ID:          string(conversation.ID),
			WorkspaceID: string(conversation.WorkspaceID),
			ServiceID:   string(conversation.ServiceID),
			UserID:      string(conversation.UserID),
			Title:       conversation.Title,
			CreatedAt:   conversation.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   conversation.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
}

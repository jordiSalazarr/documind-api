package createconversation

import (
	nethttp "net/http"

	sharedDomain "documind.jordi.org/internal/shared/domain"
	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

// CreateConversationRequest represents the request to create a conversation
type CreateConversationRequest struct {
	ServiceID string `json:"service_id" binding:"required"`
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

// Endpoint returns a gin handler for creating conversations
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateConversationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(nethttp.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cmd := Command{
			WorkspaceID: sharedDomain.ID(middleware.GetWorkspaceID(c)),
			ServiceID:   sharedDomain.ID(req.ServiceID),
			UserID:      sharedDomain.ID(middleware.GetUserID(c)),
		}

		conversation, err := h.Handle(c.Request.Context(), cmd)
		if err != nil {
			c.JSON(nethttp.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(nethttp.StatusCreated, ConversationResponse{
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

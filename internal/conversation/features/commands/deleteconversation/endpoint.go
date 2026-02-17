package deleteconversation

import (
	nethttp "net/http"

	sharedDomain "documind.jordi.org/internal/shared/domain"
	"github.com/gin-gonic/gin"
)

// Endpoint returns a gin handler for deleting conversations
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(nethttp.StatusBadRequest, gin.H{"error": "id is required"})
			return
		}

		err := h.Handle(c.Request.Context(), sharedDomain.ID(id))
		if err != nil {
			c.JSON(nethttp.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(nethttp.StatusOK, gin.H{"message": "conversation deleted"})
	}
}

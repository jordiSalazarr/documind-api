package delete

import (
	"net/http"

	"github.com/gin-gonic/gin"

	shareddomain "documind.jordi.org/internal/shared/domain"
)

// Endpoint returns a gin.HandlerFunc for the delete item endpoint.
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
			return
		}

		if err := h.Handle(c.Request.Context(), Command{
			ID: shareddomain.ID(id),
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusNoContent, nil)
	}
}

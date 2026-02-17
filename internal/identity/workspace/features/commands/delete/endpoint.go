package delete

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := h.Handle(c.Request.Context(), Command{ID: id}); err != nil {
			if errors.Is(err, ErrWorkspaceNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found", "message": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete workspace", "message": err.Error()})
			}
			return
		}
		c.Status(http.StatusNoContent)
	}
}

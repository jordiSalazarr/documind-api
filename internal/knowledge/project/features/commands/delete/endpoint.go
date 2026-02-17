package delete

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("projectId")
		if err := h.Handle(c.Request.Context(), Command{ID: id}); err != nil {
			if errors.Is(err, ErrProjectNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "project not found", "message": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete project", "message": err.Error()})
			}
			return
		}
		c.Status(http.StatusNoContent)
	}
}

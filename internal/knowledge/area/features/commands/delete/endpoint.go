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
			if errors.Is(err, ErrAreaNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "area not found", "message": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete area", "message": err.Error()})
			}
			return
		}
		c.Status(http.StatusNoContent)
	}
}

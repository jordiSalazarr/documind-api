package get

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		result, err := h.Handle(c.Request.Context(), Query{ID: id})
		if err != nil {
			if errors.Is(err, ErrAreaNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "area not found", "message": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get area", "message": err.Error()})
			}
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

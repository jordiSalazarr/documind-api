package get

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
			return
		}
		result, err := h.Handle(c.Request.Context(), Query{ID: id})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

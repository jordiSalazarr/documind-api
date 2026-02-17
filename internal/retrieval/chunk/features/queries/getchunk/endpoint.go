package getchunk

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Endpoint returns a gin.HandlerFunc for the GET /chunks/:id route.
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		chunkID := c.Param("id")
		if chunkID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "chunk id is required"})
			return
		}

		q := Query{
			ID:          chunkID,
			IncludeItem: c.Query("include_item") == "true",
		}

		result, err := h.Handle(q)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "chunk not found"})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

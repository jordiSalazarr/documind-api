package list

import (
	"net/http"
	"strconv"

	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		results, err := h.Handle(c.Request.Context(), Query{
			UserID: middleware.GetUserID(c),
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list workspaces", "message": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"workspaces": results,
			"limit":      limit,
			"offset":     offset,
		})
	}
}

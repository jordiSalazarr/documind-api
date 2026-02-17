package list

import (
	"net/http"

	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

// Endpoint returns a gin handler for listing API keys.
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		results, err := h.Handle(c.Request.Context(), Query{
			WorkspaceID: middleware.GetWorkspaceID(c),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"code": "INTERNAL_ERROR", "message": "failed to list API keys"},
			})
			return
		}

		if results == nil {
			results = []*Result{}
		}

		c.JSON(http.StatusOK, gin.H{"api_keys": results})
	}
}

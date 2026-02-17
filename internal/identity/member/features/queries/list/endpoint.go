package list

import (
	"net/http"

	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := middleware.GetWorkspaceID(c)

		results, err := h.Handle(c.Request.Context(), Query{WorkspaceID: workspaceID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"code": "INTERNAL_ERROR", "message": "failed to list members"},
			})
			return
		}

		if results == nil {
			results = []MemberResult{}
		}

		c.JSON(http.StatusOK, gin.H{"members": results})
	}
}

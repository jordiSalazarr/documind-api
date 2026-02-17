package revoke

import (
	"errors"
	"net/http"

	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

// Endpoint returns a gin handler for revoking an API key.
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"code": "BAD_REQUEST", "message": "API key ID is required"},
			})
			return
		}

		err := h.Handle(c.Request.Context(), Command{
			ID:          id,
			WorkspaceID: middleware.GetWorkspaceID(c),
		})
		if err != nil {
			switch {
			case errors.Is(err, ErrKeyNotFound):
				c.JSON(http.StatusNotFound, gin.H{
					"error": gin.H{"code": "NOT_FOUND", "message": err.Error()},
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{"code": "INTERNAL_ERROR", "message": "failed to revoke API key"},
				})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "API key revoked"})
	}
}

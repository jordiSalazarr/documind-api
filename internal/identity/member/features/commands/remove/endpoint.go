package remove

import (
	"errors"
	"net/http"

	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetUserID := c.Param("userId")

		err := h.Handle(c.Request.Context(), Command{
			WorkspaceID: middleware.GetWorkspaceID(c),
			UserID:      targetUserID,
		})
		if err != nil {
			switch {
			case errors.Is(err, ErrMemberNotFound):
				c.JSON(http.StatusNotFound, gin.H{
					"error": gin.H{"code": "NOT_FOUND", "message": err.Error()},
				})
			case errors.Is(err, ErrLastAdmin):
				c.JSON(http.StatusConflict, gin.H{
					"error": gin.H{"code": "LAST_ADMIN", "message": err.Error()},
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{"code": "INTERNAL_ERROR", "message": "failed to remove member"},
				})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "member removed"})
	}
}

package updaterole

import (
	"errors"
	"net/http"

	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

type request struct {
	Role string `json:"role" binding:"required"`
}

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()},
			})
			return
		}

		targetUserID := c.Param("userId")

		result, err := h.Handle(c.Request.Context(), Command{
			WorkspaceID: middleware.GetWorkspaceID(c),
			UserID:      targetUserID,
			NewRole:     req.Role,
		})
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidRole):
				c.JSON(http.StatusBadRequest, gin.H{
					"error": gin.H{"code": "INVALID_ROLE", "message": err.Error()},
				})
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
					"error": gin.H{"code": "INTERNAL_ERROR", "message": "failed to update role"},
				})
			}
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

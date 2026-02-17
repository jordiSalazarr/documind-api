package invite

import (
	"errors"
	"net/http"

	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

type request struct {
	Email string `json:"email" binding:"required"`
	Role  string `json:"role" binding:"required"`
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

		result, err := h.Handle(c.Request.Context(), Command{
			WorkspaceID: middleware.GetWorkspaceID(c),
			Email:       req.Email,
			Role:        req.Role,
			InvitedBy:   middleware.GetUserID(c),
		})
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidEmail):
				c.JSON(http.StatusBadRequest, gin.H{
					"error": gin.H{"code": "INVALID_EMAIL", "message": err.Error()},
				})
			case errors.Is(err, ErrInvalidRole):
				c.JSON(http.StatusBadRequest, gin.H{
					"error": gin.H{"code": "INVALID_ROLE", "message": err.Error()},
				})
			case errors.Is(err, ErrUserNotFound):
				c.JSON(http.StatusNotFound, gin.H{
					"error": gin.H{"code": "USER_NOT_FOUND", "message": err.Error()},
				})
			case errors.Is(err, ErrAlreadyMember):
				c.JSON(http.StatusConflict, gin.H{
					"error": gin.H{"code": "ALREADY_MEMBER", "message": err.Error()},
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{"code": "INTERNAL_ERROR", "message": "failed to invite member"},
				})
			}
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}

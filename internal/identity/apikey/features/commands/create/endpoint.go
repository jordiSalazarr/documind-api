package create

import (
	"errors"
	"net/http"

	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

type request struct {
	Name string `json:"name" binding:"required"`
}

// Endpoint returns a gin handler for creating an API key.
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"code": "BAD_REQUEST", "message": "name is required"},
			})
			return
		}

		result, err := h.Handle(c.Request.Context(), Command{
			WorkspaceID: middleware.GetWorkspaceID(c),
			Name:        req.Name,
			CreatedBy:   middleware.GetUserID(c),
		})
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidName):
				c.JSON(http.StatusBadRequest, gin.H{
					"error": gin.H{"code": "BAD_REQUEST", "message": err.Error()},
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{"code": "INTERNAL_ERROR", "message": "failed to create API key"},
				})
			}
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}

package create

import (
	"errors"
	"net/http"

	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

type request struct {
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
}

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "message": err.Error()})
			return
		}

		result, err := h.Handle(c.Request.Context(), Command{
			Name:      req.Name,
			Slug:      req.Slug,
			UserID:    middleware.GetUserID(c),
			UserEmail: middleware.GetUserEmail(c),
		})
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidWorkspaceName):
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workspace name", "message": err.Error()})
			case errors.Is(err, ErrInvalidSlug):
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid slug", "message": err.Error()})
			case errors.Is(err, ErrWorkspaceExists):
				c.JSON(http.StatusConflict, gin.H{"error": "workspace already exists", "message": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create workspace", "message": err.Error()})
			}
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}

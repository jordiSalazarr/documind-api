package create

import (
	"errors"
	"net/http"

	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

type request struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description"`
}

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "message": err.Error()})
			return
		}
		result, err := h.Handle(c.Request.Context(), Command{
			WorkspaceID: middleware.GetWorkspaceID(c), Name: req.Name, Slug: req.Slug, Description: req.Description,
			CreatedBy: middleware.GetUserID(c),
		})
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidProjectName):
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project name", "message": err.Error()})
			case errors.Is(err, ErrInvalidSlug):
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid slug", "message": err.Error()})
			case errors.Is(err, ErrProjectExists):
				c.JSON(http.StatusConflict, gin.H{"error": "project already exists", "message": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create project", "message": err.Error()})
			}
			return
		}
		c.JSON(http.StatusCreated, result)
	}
}

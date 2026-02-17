package create

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type request struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description"`
}

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("projectId")
		if projectID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
			return
		}
		var req request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "message": err.Error()})
			return
		}
		result, err := h.Handle(c.Request.Context(), Command{
			ProjectID: projectID, Name: req.Name, Slug: req.Slug, Description: req.Description,
		})
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidAreaName):
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid area name", "message": err.Error()})
			case errors.Is(err, ErrInvalidSlug):
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid slug", "message": err.Error()})
			case errors.Is(err, ErrAreaExists):
				c.JSON(http.StatusConflict, gin.H{"error": "area already exists", "message": err.Error()})
			case errors.Is(err, ErrProjectNotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": "project not found", "message": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create area", "message": err.Error()})
			}
			return
		}
		c.JSON(http.StatusCreated, result)
	}
}

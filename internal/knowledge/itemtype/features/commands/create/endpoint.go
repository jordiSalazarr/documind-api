package create

import (
	"net/http"

	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

type request struct {
	Name        string                 `json:"name" binding:"required"`
	Slug        string                 `json:"slug"`
	Description *string                `json:"description"`
	Icon        *string                `json:"icon"`
	Fields      map[string]interface{} `json:"fields"`
}

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := h.Handle(c.Request.Context(), Command{
			WorkspaceID: middleware.GetWorkspaceID(c), Name: req.Name, Slug: req.Slug,
			Description: req.Description, Icon: req.Icon, Fields: req.Fields,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, result)
	}
}

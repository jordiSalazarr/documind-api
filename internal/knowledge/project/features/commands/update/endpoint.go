package update

import (
	"errors"
	"net/http"

	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

type request struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("projectId")
		var req request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "message": err.Error()})
			return
		}
		result, err := h.Handle(c.Request.Context(), Command{ID: id, Name: req.Name, Description: req.Description, UpdatedBy: middleware.GetUserID(c)})
		if err != nil {
			if errors.Is(err, ErrProjectNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "project not found", "message": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update project", "message": err.Error()})
			}
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

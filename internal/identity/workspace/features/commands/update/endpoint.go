package update

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type request struct {
	Name     string                 `json:"name,omitempty"`
	Settings map[string]interface{} `json:"settings,omitempty"`
}

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "message": err.Error()})
			return
		}

		result, err := h.Handle(c.Request.Context(), Command{ID: id, Name: req.Name, Settings: req.Settings})
		if err != nil {
			if errors.Is(err, ErrWorkspaceNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found", "message": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update workspace", "message": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

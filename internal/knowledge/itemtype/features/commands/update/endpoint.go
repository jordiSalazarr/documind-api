package update

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type request struct {
	Name        string                 `json:"name" binding:"required"`
	Description *string                `json:"description"`
	Icon        *string                `json:"icon"`
	Fields      map[string]interface{} `json:"fields"`
}

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
			return
		}
		var req request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := h.Handle(c.Request.Context(), Command{
			ID: id, Name: req.Name, Description: req.Description, Icon: req.Icon, Fields: req.Fields,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

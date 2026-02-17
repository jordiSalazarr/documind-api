package getbyslug

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		result, err := h.Handle(c.Request.Context(), Query{Slug: slug})
		if err != nil {
			if errors.Is(err, ErrWorkspaceNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found", "message": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get workspace", "message": err.Error()})
			}
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

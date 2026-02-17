package updatetype

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type request struct {
	Name          *string `json:"name"`
	Slug          *string `json:"slug"`
	IsDirectional *bool   `json:"is_directional"`
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
		result, err := h.Handle(c.Request.Context(), Command{ID: id, Name: req.Name, Slug: req.Slug, IsDirectional: req.IsDirectional})
		if err != nil {
			switch {
			case errors.Is(err, ErrRelationTypeNotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			case errors.Is(err, ErrRelationTypeExists):
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			case errors.Is(err, ErrInvalidSlug):
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

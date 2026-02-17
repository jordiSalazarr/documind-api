package createtype

import (
	"errors"
	"net/http"

	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

type request struct {
	Name          string `json:"name" binding:"required"`
	Slug          string `json:"slug" binding:"required"`
	IsDirectional bool   `json:"is_directional"`
}

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := h.Handle(c.Request.Context(), Command{
			WorkspaceID: middleware.GetWorkspaceID(c), Name: req.Name, Slug: req.Slug, IsDirectional: req.IsDirectional,
		})
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidRelationTypeName):
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			case errors.Is(err, ErrInvalidSlug):
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			case errors.Is(err, ErrRelationTypeExists):
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(http.StatusCreated, result)
	}
}

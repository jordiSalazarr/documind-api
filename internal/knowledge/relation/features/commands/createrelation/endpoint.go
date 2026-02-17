package createrelation

import (
	"errors"
	"net/http"

	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

type request struct {
	FromItemID     string `json:"from_item_id" binding:"required"`
	ToItemID       string `json:"to_item_id" binding:"required"`
	RelationTypeID string `json:"relation_type_id" binding:"required"`
}

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := h.Handle(c.Request.Context(), Command{
			WorkspaceID:    middleware.GetWorkspaceID(c),
			FromItemID:     req.FromItemID,
			ToItemID:       req.ToItemID,
			RelationTypeID: req.RelationTypeID,
			CreatedBy:      middleware.GetUserID(c),
		})
		if err != nil {
			switch {
			case errors.Is(err, ErrRelationTypeNotFound):
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			case errors.Is(err, ErrRelationAlreadyExists):
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			case errors.Is(err, ErrSelfRelation):
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		c.JSON(http.StatusCreated, result)
	}
}

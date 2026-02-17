package updatereport

import (
	"errors"
	"net/http"

	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

type request struct {
	Status string `json:"status" binding:"required"`
}

// Endpoint returns a gin handler for updating a report status.
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"code": "BAD_REQUEST", "message": "report ID is required"},
			})
			return
		}

		var req request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"code": "BAD_REQUEST", "message": "status is required"},
			})
			return
		}

		err := h.Handle(c.Request.Context(), Command{
			ID:          id,
			WorkspaceID: middleware.GetWorkspaceID(c),
			Status:      req.Status,
		})
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidStatus):
				c.JSON(http.StatusBadRequest, gin.H{
					"error": gin.H{"code": "BAD_REQUEST", "message": err.Error()},
				})
			case errors.Is(err, ErrReportNotFound):
				c.JSON(http.StatusNotFound, gin.H{
					"error": gin.H{"code": "NOT_FOUND", "message": err.Error()},
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{"code": "INTERNAL_ERROR", "message": "failed to update report"},
				})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "report status updated"})
	}
}

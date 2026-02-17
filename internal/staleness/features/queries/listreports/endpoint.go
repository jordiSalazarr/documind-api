package listreports

import (
	"net/http"
	"strconv"

	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

// Endpoint returns a gin handler for listing staleness reports.
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

		results, err := h.Handle(c.Request.Context(), Query{
			WorkspaceID: middleware.GetWorkspaceID(c),
			ProjectID:   c.Query("project_id"),
			Status:      c.Query("status"),
			Page:        page,
			Limit:       limit,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"code": "INTERNAL_ERROR", "message": "failed to list reports"},
			})
			return
		}

		if results == nil {
			results = []*Result{}
		}

		c.JSON(http.StatusOK, gin.H{
			"reports": results,
			"page":    page,
			"limit":   limit,
		})
	}
}

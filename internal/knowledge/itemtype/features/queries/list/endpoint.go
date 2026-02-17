package list

import (
	"net/http"
	"strconv"

	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := middleware.GetWorkspaceID(c)

		limit := parseIntParam(c.Query("limit"), defaultLimit)
		offset := parseIntParam(c.Query("offset"), 0)

		results, err := h.Handle(c.Request.Context(), Query{
			WorkspaceID: workspaceID,
			Limit:       limit,
			Offset:      offset,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"item_types": results,
			"limit":      limit,
			"offset":     offset,
		})
	}
}

func parseIntParam(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}

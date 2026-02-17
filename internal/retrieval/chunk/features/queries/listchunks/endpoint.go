package listchunks

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Endpoint returns a gin.HandlerFunc for the GET /chunks/by-item/:item_version_id route.
func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		itemVersionID := c.Param("item_version_id")
		if itemVersionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "item_version_id is required"})
			return
		}

		limit := parseIntParam(c.Query("limit"), defaultLimit)
		offset := parseIntParam(c.Query("offset"), 0)

		q := Query{ItemVersionID: itemVersionID, Limit: limit, Offset: offset}

		results, err := h.Handle(q)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve chunks"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"chunks": results,
			"total":  len(results),
			"limit":  limit,
			"offset": offset,
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

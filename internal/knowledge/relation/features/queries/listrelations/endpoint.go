package listrelations

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		itemID := c.Query("item_id")
		if itemID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "item_id is required"})
			return
		}

		limit := parseIntParam(c.Query("limit"), defaultLimit)
		offset := parseIntParam(c.Query("offset"), 0)

		results, err := h.Handle(c.Request.Context(), Query{
			ItemID: itemID,
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"relations": results,
			"limit":     limit,
			"offset":    offset,
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

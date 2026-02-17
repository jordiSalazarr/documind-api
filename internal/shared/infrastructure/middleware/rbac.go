package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var roleHierarchy = map[string]int{
	"reader": 1,
	"editor": 2,
	"admin":  3,
}

func RequireRole(minRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("member_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{"code": "FORBIDDEN", "message": "no role assigned"},
			})
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{"code": "FORBIDDEN", "message": "invalid role"},
			})
			return
		}

		userLevel, ok := roleHierarchy[roleStr]
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{"code": "FORBIDDEN", "message": "unknown role"},
			})
			return
		}

		requiredLevel, ok := roleHierarchy[minRole]
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"code": "INTERNAL_ERROR", "message": "invalid required role configuration"},
			})
			return
		}

		if userLevel < requiredLevel {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{"code": "FORBIDDEN", "message": "insufficient permissions, requires " + minRole + " role"},
			})
			return
		}

		c.Next()
	}
}

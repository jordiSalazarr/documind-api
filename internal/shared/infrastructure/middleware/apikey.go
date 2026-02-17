package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIKeyChecker validates an API key and returns the associated workspace ID.
type APIKeyChecker interface {
	ValidateKey(rawKey string) (workspaceID string, err error)
}

// APIKeyMiddleware authenticates requests using the X-API-Key header.
// On success it sets "workspace_id" in the gin context, matching WorkspaceMiddleware behavior.
func APIKeyMiddleware(checker APIKeyChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "UNAUTHORIZED", "message": "missing X-API-Key header"},
			})
			return
		}

		workspaceID, err := checker.ValidateKey(apiKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "UNAUTHORIZED", "message": "invalid or expired API key"},
			})
			return
		}

		c.Set("workspace_id", workspaceID)
		c.Next()
	}
}

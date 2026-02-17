package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func WorkspaceMiddleware(checker MembershipChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		workspaceID := c.GetHeader("X-Workspace-ID")
		if workspaceID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"code": "MISSING_WORKSPACE", "message": "X-Workspace-ID header is required"},
			})
			return
		}

		userID, _ := c.Get("auth_user_id")
		userIDStr, ok := userID.(string)
		if !ok || userIDStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "UNAUTHORIZED", "message": "user not authenticated"},
			})
			return
		}

		role, err := checker.CheckMembership(c.Request.Context(), workspaceID, userIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": gin.H{"code": "FORBIDDEN", "message": "not a member of this workspace"},
			})
			return
		}

		c.Set("workspace_id", workspaceID)
		c.Set("member_role", role)
		c.Next()
	}
}

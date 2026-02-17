package middleware

import (
	"github.com/gin-gonic/gin"
)

func GetUserID(c *gin.Context) string {
	val, _ := c.Get("auth_user_id")
	s, _ := val.(string)
	return s
}

func GetUserEmail(c *gin.Context) string {
	val, _ := c.Get("auth_user_email")
	s, _ := val.(string)
	return s
}

func GetUserName(c *gin.Context) string {
	val, _ := c.Get("auth_user_name")
	s, _ := val.(string)
	return s
}

func GetWorkspaceID(c *gin.Context) string {
	val, _ := c.Get("workspace_id")
	s, _ := val.(string)
	return s
}

func GetMemberRole(c *gin.Context) string {
	val, _ := c.Get("member_role")
	s, _ := val.(string)
	return s
}

package http

import (
	"github.com/gin-gonic/gin"

	"documind.jordi.org/internal/interfaces/http/handlers"
)

func SetupWorkspaceRoutes(router *gin.RouterGroup, workspaceHandler *handlers.WorkspaceHandler) {
	workspaces := router.Group("/workspaces")
	{
		workspaces.POST("", workspaceHandler.CreateWorkspace)
		workspaces.GET("", workspaceHandler.ListWorkspaces)
		workspaces.GET("/:id", workspaceHandler.GetWorkspace)
		workspaces.PUT("/:id", workspaceHandler.UpdateWorkspace)
		workspaces.DELETE("/:id", workspaceHandler.DeleteWorkspace)
		workspaces.GET("/slug/:slug", workspaceHandler.GetWorkspaceBySlug)
	}
}

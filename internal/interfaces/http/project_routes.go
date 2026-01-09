package http

import (
	"documind.jordi.org/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

func SetupProjectRoutes(router *gin.RouterGroup, projectHandler *handlers.ProjectHandler, areaHandler *handlers.AreaHandler) {
	// Project routes - using /projects with workspace_id as a query param or in the body
	// To avoid route conflicts with /workspaces/:id
	projects := router.Group("/projects")
	{
		projects.POST("", projectHandler.Create)
		projects.GET("", projectHandler.List)
		projects.GET("/:id", projectHandler.Get)
		projects.PUT("/:id", projectHandler.Update)
		projects.DELETE("/:id", projectHandler.Delete)
	}

	// Service routes
	services := router.Group("/services")
	{
		services.POST("", areaHandler.Create)
		services.GET("", areaHandler.List)
		services.GET("/:id", areaHandler.Get)
		services.PUT("/:id", areaHandler.Update)
		services.DELETE("/:id", areaHandler.Delete)
	}
}

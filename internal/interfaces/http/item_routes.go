package http

import (
	"documind.jordi.org/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

func SetupItemRoutes(router *gin.RouterGroup, itemHandler *handlers.ItemHandler) {
	items := router.Group("/items")
	{
		items.POST("", itemHandler.Create)
		items.GET("", itemHandler.ListByProject)
		items.GET("/:id", itemHandler.Get)
		items.PUT("/:id", itemHandler.Update)
		items.POST("/:id/publish", itemHandler.Publish)
		items.DELETE("/:id", itemHandler.Delete)

		// Version endpoints
		items.POST("/:id/versions", itemHandler.CreateVersion)
		items.GET("/:id/versions", itemHandler.ListVersions)
		items.GET("/:id/versions/:version", itemHandler.GetVersion)
	}
}

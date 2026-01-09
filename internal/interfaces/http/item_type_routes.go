package http

import (
	"documind.jordi.org/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

func SetupItemTypeRoutes(router *gin.RouterGroup, itemTypeHandler *handlers.ItemTypeHandler) {
	itemTypes := router.Group("/item-types")
	{
		itemTypes.POST("", itemTypeHandler.Create)
		itemTypes.GET("", itemTypeHandler.List)
		itemTypes.GET("/:id", itemTypeHandler.Get)
		itemTypes.PUT("/:id", itemTypeHandler.Update)
		itemTypes.DELETE("/:id", itemTypeHandler.Delete)
	}
}

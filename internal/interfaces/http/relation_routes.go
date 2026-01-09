package http

import (
	"documind.jordi.org/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRelationRoutes(router *gin.RouterGroup, handler *handlers.RelationHandler) {
	// Relation Type routes
	relationTypes := router.Group("/relation-types")
	{
		relationTypes.POST("", handler.CreateRelationType)
		relationTypes.GET("/:id", handler.GetRelationType)
		relationTypes.GET("", handler.ListRelationTypes)
		relationTypes.PUT("/:id", handler.UpdateRelationType)
		relationTypes.DELETE("/:id", handler.DeleteRelationType)
	}

	// Relation routes
	relations := router.Group("/relations")
	{
		relations.POST("", handler.CreateRelation)
		relations.GET("/:id", handler.GetRelation)
		relations.GET("", handler.ListRelationsByItem)
		relations.DELETE("/:id", handler.DeleteRelation)
	}
}

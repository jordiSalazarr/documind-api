package http

import (
	"documind.jordi.org/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterSearchRoutes(router *gin.RouterGroup, handler *handlers.SearchHandler) {
	search := router.Group("/search")
	{
		search.GET("/items", handler.SearchItemVersions)
		search.POST("/semantic", handler.SemanticSearch)
	}

	embeddings := router.Group("/embeddings")
	{
		embeddings.PUT("", handler.UpdateEmbedding)
	}
}

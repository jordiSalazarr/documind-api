package http

import (
	"documind.jordi.org/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterChunkRoutes(router *gin.RouterGroup, handler *handlers.ChunkHandler) {
	chunks := router.Group("/chunks")
	{
		// Get a chunk by ID (with optional item info via ?include_item=true)
		chunks.GET("/:id", handler.GetChunk)

		// Get all chunks for an item version
		chunks.GET("/by-item/:item_version_id", handler.GetChunksByItem)
	}
}

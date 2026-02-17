package retrieval

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"documind.jordi.org/internal/retrieval/chunk/features/queries/getchunk"
	"documind.jordi.org/internal/retrieval/chunk/features/queries/listchunks"
	"documind.jordi.org/internal/retrieval/search/features/queries/searchitems"
	"documind.jordi.org/internal/shared/infrastructure/eventbus"
)

// Deps holds external dependencies needed by the retrieval context.
type Deps struct {
	DB                *sql.DB
	Bus               eventbus.EventBus
	ItemVersionReader getchunk.ItemVersionReader
	SearchItemsHandler *searchitems.Handler
}

// RegisterRoutes registers all retrieval context routes under /retrieval group.
func RegisterRoutes(rg *gin.RouterGroup, deps Deps) {
	rg2 := rg.Group("/retrieval")

	// Chunk routes
	getChunkHandler := getchunk.NewHandler(deps.DB, deps.ItemVersionReader)
	listChunksHandler := listchunks.NewHandler(deps.DB)

	chunks := rg2.Group("/chunks")
	{
		chunks.GET("/:id", getchunk.Endpoint(getChunkHandler))
		chunks.GET("/by-item/:item_version_id", listchunks.Endpoint(listChunksHandler))
	}

	// Search routes
	search := rg2.Group("/search")
	{
		search.GET("/items", searchitems.Endpoint(deps.SearchItemsHandler))
		search.POST("/semantic", searchitems.SemanticEndpoint(deps.SearchItemsHandler))
	}
}

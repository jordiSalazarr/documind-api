package staleness

import (
	"database/sql"

	"documind.jordi.org/internal/retrieval/search/features/queries/retrieve"
	"documind.jordi.org/internal/staleness/features/commands/check"
	"documind.jordi.org/internal/staleness/features/commands/updatereport"
	"documind.jordi.org/internal/staleness/features/queries/getreport"
	"documind.jordi.org/internal/staleness/features/queries/listreports"
	"github.com/gin-gonic/gin"
)

// Deps holds external dependencies for the staleness context.
type Deps struct {
	DB              *sql.DB
	RetrieveHandler *retrieve.Handler
	ChatClient      check.ChatClient
}

// RegisterAPIKeyRoutes registers routes authenticated via API key (for CI/CD).
// Expected to be mounted on an API-key-scoped router group.
func RegisterAPIKeyRoutes(rg *gin.RouterGroup, deps Deps) {
	retriever := NewRetrieverAdapter(deps.RetrieveHandler, deps.DB)
	checkHandler := check.NewHandler(deps.DB, retriever, deps.ChatClient)

	rg.POST("/staleness/check", check.Endpoint(checkHandler))
}

// RegisterJWTRoutes registers routes authenticated via JWT (for dashboard).
// Expected to be mounted on a workspace-scoped router group.
func RegisterJWTRoutes(rg *gin.RouterGroup, deps Deps) {
	reports := rg.Group("/staleness/reports")
	{
		reports.GET("", listreports.Endpoint(listreports.NewHandler(deps.DB)))
		reports.GET("/:id", getreport.Endpoint(getreport.NewHandler(deps.DB)))
		reports.PATCH("/:id", updatereport.Endpoint(updatereport.NewHandler(deps.DB)))
	}
}

package apikey

import (
	"database/sql"

	"documind.jordi.org/internal/identity/apikey/features/commands/create"
	"documind.jordi.org/internal/identity/apikey/features/commands/revoke"
	"documind.jordi.org/internal/identity/apikey/features/queries/list"
	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers API key management routes under the given router group.
// Expected path: /identity/api-keys
func RegisterRoutes(rg *gin.RouterGroup, db *sql.DB) {
	keys := rg.Group("/api-keys")
	{
		keys.GET("", list.Endpoint(list.NewHandler(db)))
		keys.POST("", middleware.RequireRole("admin"), create.Endpoint(create.NewHandler(db)))
		keys.DELETE("/:id", middleware.RequireRole("admin"), revoke.Endpoint(revoke.NewHandler(db)))
	}
}

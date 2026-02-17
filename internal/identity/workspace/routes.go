package workspace

import (
	"database/sql"

	"documind.jordi.org/internal/identity/workspace/features/commands/create"
	del "documind.jordi.org/internal/identity/workspace/features/commands/delete"
	"documind.jordi.org/internal/identity/workspace/features/commands/update"
	"documind.jordi.org/internal/identity/workspace/features/queries/get"
	"documind.jordi.org/internal/identity/workspace/features/queries/getbyslug"
	"documind.jordi.org/internal/identity/workspace/features/queries/list"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, db *sql.DB) {
	workspaces := rg.Group("/workspaces")
	{
		workspaces.POST("", create.Endpoint(create.NewHandler(db)))
		workspaces.GET("", list.Endpoint(list.NewHandler(db)))
		workspaces.GET("/:id", get.Endpoint(get.NewHandler(db)))
		workspaces.PUT("/:id", update.Endpoint(update.NewHandler(db)))
		workspaces.DELETE("/:id", del.Endpoint(del.NewHandler(db)))
		workspaces.GET("/slug/:slug", getbyslug.Endpoint(getbyslug.NewHandler(db)))
	}
}

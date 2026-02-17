package member

import (
	"database/sql"

	"documind.jordi.org/internal/identity/member/features/commands/invite"
	"documind.jordi.org/internal/identity/member/features/commands/remove"
	"documind.jordi.org/internal/identity/member/features/commands/updaterole"
	"documind.jordi.org/internal/identity/member/features/queries/list"
	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all member management routes.
// All routes require admin role except list (any member can view).
func RegisterRoutes(rg *gin.RouterGroup, db *sql.DB) {
	members := rg.Group("/members")
	{
		members.GET("", list.Endpoint(list.NewHandler(db)))
		members.POST("/invite", middleware.RequireRole("admin"), invite.Endpoint(invite.NewHandler(db)))
		members.DELETE("/:userId", middleware.RequireRole("admin"), remove.Endpoint(remove.NewHandler(db)))
		members.PUT("/:userId/role", middleware.RequireRole("admin"), updaterole.Endpoint(updaterole.NewHandler(db)))
	}
}

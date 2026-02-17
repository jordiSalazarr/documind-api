package identity

import (
	"database/sql"

	"documind.jordi.org/internal/identity/apikey"
	"documind.jordi.org/internal/identity/member"
	"documind.jordi.org/internal/identity/user"
	"documind.jordi.org/internal/identity/workspace"
	"documind.jordi.org/internal/shared/infrastructure/jwt"
	"github.com/gin-gonic/gin"
)

// RegisterPublicRoutes registers unauthenticated routes (register, login).
func RegisterPublicRoutes(rg *gin.RouterGroup, db *sql.DB, jwtService *jwt.Service) {
	user.RegisterPublicRoutes(rg, db, jwtService)
}

// RegisterProtectedRoutes registers routes that require authentication but not workspace context.
func RegisterProtectedRoutes(rg *gin.RouterGroup, db *sql.DB) {
	user.RegisterProtectedRoutes(rg, db)
	workspace.RegisterRoutes(rg.Group("/identity"), db)
}

// RegisterWorkspaceRoutes registers routes that require auth + workspace membership.
func RegisterWorkspaceRoutes(rg *gin.RouterGroup, db *sql.DB) {
	member.RegisterRoutes(rg.Group("/identity"), db)
	apikey.RegisterRoutes(rg.Group("/identity"), db)
}

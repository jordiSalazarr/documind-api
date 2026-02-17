package user

import (
	"database/sql"

	"documind.jordi.org/internal/identity/user/features/commands/login"
	"documind.jordi.org/internal/identity/user/features/commands/register"
	"documind.jordi.org/internal/identity/user/features/queries/me"
	"documind.jordi.org/internal/shared/infrastructure/jwt"
	"github.com/gin-gonic/gin"
)

// RegisterPublicRoutes registers unauthenticated auth routes (register, login).
func RegisterPublicRoutes(rg *gin.RouterGroup, db *sql.DB, jwtService *jwt.Service) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", register.Endpoint(register.NewHandler(db, jwtService)))
		auth.POST("/login", login.Endpoint(login.NewHandler(db, jwtService)))
	}
}

// RegisterProtectedRoutes registers authenticated user routes (me).
func RegisterProtectedRoutes(rg *gin.RouterGroup, db *sql.DB) {
	users := rg.Group("/identity/users")
	{
		users.GET("/me", me.Endpoint(me.NewHandler(db)))
	}
}

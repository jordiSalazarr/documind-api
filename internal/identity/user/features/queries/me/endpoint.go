package me

import (
	"errors"
	"net/http"

	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "UNAUTHORIZED", "message": "user not authenticated"},
			})
			return
		}

		result, err := h.Handle(c.Request.Context(), Query{UserID: userID})
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"error": gin.H{"code": "NOT_FOUND", "message": "user not found"},
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{"code": "INTERNAL_ERROR", "message": "failed to get user profile"},
			})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

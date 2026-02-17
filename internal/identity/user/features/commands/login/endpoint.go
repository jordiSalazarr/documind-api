package login

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type request struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Endpoint(h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()},
			})
			return
		}

		result, err := h.Handle(c.Request.Context(), Command{
			Email:    req.Email,
			Password: req.Password,
		})
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidEmail), errors.Is(err, ErrInvalidPassword):
				c.JSON(http.StatusBadRequest, gin.H{
					"error": gin.H{"code": "INVALID_INPUT", "message": err.Error()},
				})
			case errors.Is(err, ErrInvalidCredentials):
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": gin.H{"code": "INVALID_CREDENTIALS", "message": err.Error()},
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{"code": "INTERNAL_ERROR", "message": "failed to login"},
				})
			}
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

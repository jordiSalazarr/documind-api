package register

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type request struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
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
			Name:     req.Name,
		})
		if err != nil {
			switch {
			case errors.Is(err, ErrInvalidEmail):
				c.JSON(http.StatusBadRequest, gin.H{
					"error": gin.H{"code": "INVALID_EMAIL", "message": err.Error()},
				})
			case errors.Is(err, ErrInvalidPassword):
				c.JSON(http.StatusBadRequest, gin.H{
					"error": gin.H{"code": "INVALID_PASSWORD", "message": err.Error()},
				})
			case errors.Is(err, ErrInvalidName):
				c.JSON(http.StatusBadRequest, gin.H{
					"error": gin.H{"code": "INVALID_NAME", "message": err.Error()},
				})
			case errors.Is(err, ErrEmailTaken):
				c.JSON(http.StatusConflict, gin.H{
					"error": gin.H{"code": "EMAIL_TAKEN", "message": err.Error()},
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{"code": "INTERNAL_ERROR", "message": "failed to register user"},
				})
			}
			return
		}

		c.JSON(http.StatusCreated, result)
	}
}

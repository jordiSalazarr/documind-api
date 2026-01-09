package http

import (
	"documind.jordi.org/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterChatRoutes(router *gin.RouterGroup, handler *handlers.ChatHandler) {
	chat := router.Group("/chat")
	{
		// Conversation endpoints
		chat.POST("/conversations", handler.CreateConversation)
		chat.GET("/conversations", handler.ListConversations)
		chat.GET("/conversations/:id", handler.GetConversation)
		chat.DELETE("/conversations/:id", handler.DeleteConversation)

		// Message endpoints
		chat.GET("/conversations/:id/messages", handler.GetMessages)
		chat.POST("/messages", handler.SendMessage)
	}
}

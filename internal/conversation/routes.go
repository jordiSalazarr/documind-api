package conversation

import (
	"database/sql"

	"documind.jordi.org/internal/conversation/features/commands/createconversation"
	"documind.jordi.org/internal/conversation/features/commands/deleteconversation"
	"documind.jordi.org/internal/conversation/features/commands/sendmessage"
	"documind.jordi.org/internal/conversation/features/queries/getconversation"
	"documind.jordi.org/internal/conversation/features/queries/getmessages"
	"documind.jordi.org/internal/conversation/features/queries/listconversations"
	"documind.jordi.org/internal/shared/infrastructure/middleware"
	"github.com/gin-gonic/gin"
)

// Deps holds external dependencies needed by the conversation context
type Deps struct {
	DB            *sql.DB
	Retriever     sendmessage.Retriever
	ChatClient    sendmessage.ChatClient
	MemoryService *MemoryService
}

// RegisterRoutes registers conversation context routes under /conversation group
func RegisterRoutes(rg *gin.RouterGroup, deps Deps) {
	convRepo := NewConversationRepo(deps.DB)
	msgRepo := NewMessageRepo(deps.DB)

	sendMsgHandler := sendmessage.NewHandler(
		convRepo, convRepo,
		msgRepo, msgRepo,
		deps.Retriever,
		deps.ChatClient,
		deps.MemoryService,
		sendmessage.DefaultConfig(),
	)
	sendMsgHandler.SetCitationChecker(newCitationCheckerAdapter())

	editor := middleware.RequireRole("editor")

	conversations := rg.Group("/conversation/conversations")
	{
		conversations.POST("", editor, createconversation.Endpoint(createconversation.NewHandler(convRepo)))
		conversations.GET("", listconversations.Endpoint(listconversations.NewHandler(convRepo)))
		conversations.GET("/:id", getconversation.Endpoint(getconversation.NewHandler(convRepo)))
		conversations.DELETE("/:id", editor, deleteconversation.Endpoint(deleteconversation.NewHandler(convRepo)))
		conversations.POST("/:id/messages", editor, sendmessage.Endpoint(sendMsgHandler))
		conversations.POST("/:id/messages/stream", editor, sendmessage.StreamEndpoint(sendMsgHandler))
		conversations.GET("/:id/messages", getmessages.Endpoint(getmessages.NewHandler(msgRepo)))
	}
}

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"awebo/app/dto"
	"awebo/app/exceptions"
)

type conversationHandler struct {
	*Handler
}

func newConversationHandler(h *Handler) *conversationHandler {
	return &conversationHandler{h}
}

func (h *conversationHandler) RegisterRoutes(router *gin.RouterGroup) {
	conversations := router.Group("/conversations")
	{
		conversations.GET("", h.requireAuth(), h.list)
		conversations.POST("", h.requireAuth(), h.start)
		conversations.GET("/:id/messages", h.requireAuth(), h.listMessages)
		conversations.POST("/:id/messages", h.requireAuth(), h.sendMessage)
		conversations.POST("/:id/read", h.requireAuth(), h.markRead)
	}
}

func (h *conversationHandler) list(c *gin.Context) {
	userID, _ := currentUserID(c)

	var query dto.PageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	list, err := h.services.Conversation.ListMine(userID, query.Page, query.Limit)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *conversationHandler) start(c *gin.Context) {
	userID, _ := currentUserID(c)

	var input dto.StartConversationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	conversation, err := h.services.Conversation.StartOrGet(userID, input.Username)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, conversation)
}

func (h *conversationHandler) listMessages(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, parseErr := uuid.Parse(c.Param("id"))
	if parseErr != nil {
		respondError(c, exceptions.NotFound(""))
		return
	}

	var query dto.PageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	list, err := h.services.Conversation.ListMessages(userID, id, query.Page, query.Limit)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *conversationHandler) sendMessage(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, parseErr := uuid.Parse(c.Param("id"))
	if parseErr != nil {
		respondError(c, exceptions.NotFound(""))
		return
	}

	var input dto.SendMessageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	message, err := h.services.Conversation.Send(userID, id, input.Body)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, message)
}

func (h *conversationHandler) markRead(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, parseErr := uuid.Parse(c.Param("id"))
	if parseErr != nil {
		respondError(c, exceptions.NotFound(""))
		return
	}

	if err := h.services.Conversation.MarkRead(userID, id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

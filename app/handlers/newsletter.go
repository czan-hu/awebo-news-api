package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"awebo/app/dto"
	"awebo/app/exceptions"
)

type newsletterHandler struct {
	*Handler
}

func newNewsletterHandler(h *Handler) *newsletterHandler {
	return &newsletterHandler{h}
}

func (h *newsletterHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/newsletter/subscribe", h.subscribe)
}

func (h *newsletterHandler) subscribe(c *gin.Context) {
	var input dto.SubscribeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	if err := h.services.Newsletter.Subscribe(input.Email); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusAccepted)
}

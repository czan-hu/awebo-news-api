package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type contactHandler struct {
	*Handler
}

func newContactHandler(h *Handler) *contactHandler {
	return &contactHandler{h}
}

func (h *contactHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/contacts", h.list)
}

func (h *contactHandler) list(c *gin.Context) {
	contacts, err := h.services.Contact.List()
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, contacts)
}

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type categoryHandler struct {
	*Handler
}

func newCategoryHandler(h *Handler) *categoryHandler {
	return &categoryHandler{h}
}

func (h *categoryHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/categories", h.list)
}

func (h *categoryHandler) list(c *gin.Context) {
	categories, err := h.services.Category.List()
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, categories)
}

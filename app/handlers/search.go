package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"awebo/app/dto"
	"awebo/app/exceptions"
)

type searchHandler struct {
	*Handler
}

func newSearchHandler(h *Handler) *searchHandler {
	return &searchHandler{h}
}

func (h *searchHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/search", h.search)
}

func (h *searchHandler) search(c *gin.Context) {
	var query dto.SearchQuery
	if err := c.ShouldBindQuery(&query); err != nil || len(query.Q) < 2 {
		respondError(c, exceptions.BadRequest("Проверьте правильность заполнения полей."))
		return
	}

	result, err := h.services.Search.Search(query.Q, query.Limit)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"awebo/app/dto"
	"awebo/app/exceptions"
)

type creatorRequestHandler struct {
	*Handler
}

func newCreatorRequestHandler(h *Handler) *creatorRequestHandler {
	return &creatorRequestHandler{h}
}

func (h *creatorRequestHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/users/me/creator-request", h.requireAuth(), h.mine)
	router.POST("/users/me/creator-request", h.requireAuth(), h.submit)

	admin := router.Group("/admin/creator-requests", h.requireAuth(), h.requireAdmin())
	{
		admin.GET("", h.list)
		admin.POST("/:id/approve", h.approve)
		admin.POST("/:id/reject", h.reject)
	}
}

// mine returns the caller's latest application, or 404 if they never applied.
func (h *creatorRequestHandler) mine(c *gin.Context) {
	userID, _ := currentUserID(c)
	req, err := h.services.CreatorRequest.LatestForUser(userID)
	if err != nil {
		respondError(c, err)
		return
	}
	if req == nil {
		respondError(c, exceptions.NotFound(""))
		return
	}
	c.JSON(http.StatusOK, req)
}

func (h *creatorRequestHandler) submit(c *gin.Context) {
	userID, _ := currentUserID(c)
	req, err := h.services.CreatorRequest.Submit(userID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, req)
}

func (h *creatorRequestHandler) list(c *gin.Context) {
	var query dto.PageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	list, err := h.services.CreatorRequest.ListPending(query.Page, query.Limit)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *creatorRequestHandler) approve(c *gin.Context) {
	h.decide(c, h.services.CreatorRequest.Approve)
}

func (h *creatorRequestHandler) reject(c *gin.Context) {
	h.decide(c, h.services.CreatorRequest.Reject)
}

func (h *creatorRequestHandler) decide(c *gin.Context, apply func(reviewerID, requestID uuid.UUID) error) {
	reviewerID, _ := currentUserID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	if err := apply(reviewerID, id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

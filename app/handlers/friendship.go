package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"awebo/app/dto"
	"awebo/app/exceptions"
)

type friendshipHandler struct {
	*Handler
}

func newFriendshipHandler(h *Handler) *friendshipHandler {
	return &friendshipHandler{h}
}

func (h *friendshipHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/users/:username/friend-requests", h.requireAuth(), h.send)
	router.GET("/friend-requests", h.requireAuth(), h.listRequests)
	router.POST("/friend-requests/:id/accept", h.requireAuth(), h.accept)
	router.POST("/friend-requests/:id/decline", h.requireAuth(), h.decline)
	router.DELETE("/friendships/:id", h.requireAuth(), h.unfriend)
	router.GET("/friends", h.requireAuth(), h.listFriends)
}

func (h *friendshipHandler) send(c *gin.Context) {
	userID, _ := currentUserID(c)

	req, err := h.services.Friendship.SendRequest(userID, c.Param("username"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, req)
}

// listRequests defaults to the caller's incoming pending requests; pass
// ?direction=outgoing for the ones they sent.
func (h *friendshipHandler) listRequests(c *gin.Context) {
	userID, _ := currentUserID(c)

	var query dto.PageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	var (
		list any
		err  *exceptions.AppError
	)
	if c.Query("direction") == "outgoing" {
		list, err = h.services.Friendship.ListOutgoing(userID, query.Page, query.Limit)
	} else {
		list, err = h.services.Friendship.ListIncoming(userID, query.Page, query.Limit)
	}
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *friendshipHandler) accept(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, parseErr := uuid.Parse(c.Param("id"))
	if parseErr != nil {
		respondError(c, exceptions.NotFound(""))
		return
	}

	req, err := h.services.Friendship.Accept(userID, id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, req)
}

func (h *friendshipHandler) decline(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, parseErr := uuid.Parse(c.Param("id"))
	if parseErr != nil {
		respondError(c, exceptions.NotFound(""))
		return
	}

	if err := h.services.Friendship.Decline(userID, id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *friendshipHandler) unfriend(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, parseErr := uuid.Parse(c.Param("id"))
	if parseErr != nil {
		respondError(c, exceptions.NotFound(""))
		return
	}

	if err := h.services.Friendship.Unfriend(userID, id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *friendshipHandler) listFriends(c *gin.Context) {
	userID, _ := currentUserID(c)

	var query dto.PageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	list, err := h.services.Friendship.ListFriends(userID, query.Page, query.Limit)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"awebo/app/dto"
	"awebo/app/exceptions"
)

type forumHandler struct {
	*Handler
}

func newForumHandler(h *Handler) *forumHandler {
	return &forumHandler{h}
}

func (h *forumHandler) RegisterRoutes(router *gin.RouterGroup) {
	forum := router.Group("/forum")
	{
		forum.GET("/categories", h.categories)
		forum.GET("/topics", h.listTopics)
		forum.POST("/topics", h.requireAuth(), h.createTopic)
		forum.GET("/topics/:id", h.getTopic)
		forum.GET("/topics/:id/comments", h.optionalAuth(), h.listComments)
		forum.POST("/topics/:id/comments", h.requireAuth(), h.createComment)
		forum.PUT("/comments/:id/vote", h.requireAuth(), h.vote)
	}
}

func (h *forumHandler) categories(c *gin.Context) {
	c.JSON(http.StatusOK, h.services.Forum.Categories())
}

func (h *forumHandler) listTopics(c *gin.Context) {
	category := c.Query("category")
	sort := c.DefaultQuery("sort", "new")

	var query dto.PageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	list, err := h.services.Forum.ListTopics(category, sort, query.Page, query.Limit)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *forumHandler) createTopic(c *gin.Context) {
	userID, _ := currentUserID(c)

	var input dto.CreateTopicInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	topic, err := h.services.Forum.CreateTopic(userID, input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, topic)
}

func (h *forumHandler) getTopic(c *gin.Context) {
	topic, err := h.services.Forum.GetTopic(c.Param("id"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, topic)
}

func (h *forumHandler) listComments(c *gin.Context) {
	sort := c.DefaultQuery("sort", "best")
	comments, err := h.services.Forum.ListComments(c.Param("id"), sort, viewerFromContext(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, comments)
}

func (h *forumHandler) createComment(c *gin.Context) {
	userID, _ := currentUserID(c)

	var input dto.CreateCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	comment, err := h.services.Forum.CreateComment(c.Param("id"), userID, input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, comment)
}

func (h *forumHandler) vote(c *gin.Context) {
	userID, _ := currentUserID(c)

	var input dto.VoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	result, err := h.services.Forum.Vote(c.Param("id"), userID, input.Direction)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

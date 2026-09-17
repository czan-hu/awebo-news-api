package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"awebo/app/dto"
	"awebo/app/exceptions"
)

type postHandler struct {
	*Handler
}

func newPostHandler(h *Handler) *postHandler {
	return &postHandler{h}
}

func (h *postHandler) RegisterRoutes(router *gin.RouterGroup) {
	posts := router.Group("/posts")
	{
		posts.POST("", h.requireAuth(), h.create)
		posts.POST("/images", h.requireAuth(), h.uploadImage)
		posts.GET("/feed", h.requireAuth(), h.feed)
		posts.GET("/:id", h.optionalAuth(), h.getByID)
		posts.POST("/:id/like", h.requireAuth(), h.like)
		posts.DELETE("/:id/like", h.requireAuth(), h.unlike)
		posts.POST("/:id/comments", h.requireAuth(), h.createComment)
		posts.GET("/:id/comments", h.optionalAuth(), h.listComments)
	}
}

func (h *postHandler) create(c *gin.Context) {
	userID, _ := currentUserID(c)

	var input dto.CreatePostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	post, err := h.services.Post.Create(userID, input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, post)
}

// uploadImage stores an optional image attachment for a post, mirroring
// articleHandler.uploadImage.
func (h *postHandler) uploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		respondError(c, exceptions.BadRequest("Прикрепите файл изображения."))
		return
	}
	defer file.Close()

	url, err := h.services.Upload.SaveImage("posts", file, header)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *postHandler) feed(c *gin.Context) {
	userID, _ := currentUserID(c)

	var query dto.PageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	list, err := h.services.Post.Feed(userID, query.Page, query.Limit)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *postHandler) getByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, exceptions.NotFound(""))
		return
	}

	post, svcErr := h.services.Post.GetByID(viewerFromContext(c), id)
	if svcErr != nil {
		respondError(c, svcErr)
		return
	}
	c.JSON(http.StatusOK, post)
}

func (h *postHandler) like(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, exceptions.NotFound(""))
		return
	}

	post, svcErr := h.services.Post.Like(userID, id)
	if svcErr != nil {
		respondError(c, svcErr)
		return
	}
	c.JSON(http.StatusOK, post)
}

func (h *postHandler) unlike(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, exceptions.NotFound(""))
		return
	}

	post, svcErr := h.services.Post.Unlike(userID, id)
	if svcErr != nil {
		respondError(c, svcErr)
		return
	}
	c.JSON(http.StatusOK, post)
}

func (h *postHandler) createComment(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, exceptions.NotFound(""))
		return
	}

	var input dto.CreatePostCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	comment, svcErr := h.services.Post.CreateComment(userID, id, input)
	if svcErr != nil {
		respondError(c, svcErr)
		return
	}
	c.JSON(http.StatusCreated, comment)
}

func (h *postHandler) listComments(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, exceptions.NotFound(""))
		return
	}

	var query dto.PageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	comments, svcErr := h.services.Post.ListComments(id, query.Page, query.Limit)
	if svcErr != nil {
		respondError(c, svcErr)
		return
	}
	c.JSON(http.StatusOK, comments)
}

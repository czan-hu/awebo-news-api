package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"awebo/app/dto"
	"awebo/app/exceptions"
)

type articleHandler struct {
	*Handler
}

func newArticleHandler(h *Handler) *articleHandler {
	return &articleHandler{h}
}

func (h *articleHandler) RegisterRoutes(router *gin.RouterGroup) {
	articles := router.Group("/articles")
	{
		articles.GET("", h.optionalAuth(), h.list)
		articles.POST("", h.requireAuth(), h.requirePublisher(), h.create)
		articles.POST("/images", h.requireAuth(), h.requirePublisher(), h.uploadImage)
		articles.GET("/:slug", h.optionalAuth(), h.getBySlug)
		articles.PUT("/:slug", h.requireAuth(), h.update)
		articles.DELETE("/:slug", h.requireAuth(), h.remove)
		articles.GET("/:slug/related", h.optionalAuth(), h.related)
		articles.POST("/:slug/like", h.requireAuth(), h.like)
		articles.DELETE("/:slug/like", h.requireAuth(), h.unlike)
		articles.POST("/:slug/save", h.requireAuth(), h.save)
		articles.DELETE("/:slug/save", h.requireAuth(), h.unsave)
		articles.POST("/:slug/view", h.requireAuth(), h.view)
	}
}

func (h *articleHandler) list(c *gin.Context) {
	var query dto.ArticleListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	list, err := h.services.Article.List(query, viewerFromContext(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *articleHandler) create(c *gin.Context) {
	userID, _ := currentUserID(c)

	var input dto.CreateArticleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	article, err := h.services.Article.Create(userID, input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, article)
}

func (h *articleHandler) update(c *gin.Context) {
	userID, _ := currentUserID(c)

	var input dto.UpdateArticleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	article, err := h.services.Article.Update(userID, c.Param("slug"), input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, article)
}

func (h *articleHandler) remove(c *gin.Context) {
	userID, _ := currentUserID(c)
	if err := h.services.Article.Delete(userID, c.Param("slug")); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// uploadImage stores a cover or inline content image for the article editor.
// It's a standalone endpoint (not tied to a specific article) since the
// image is picked before the article exists yet.
func (h *articleHandler) uploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		respondError(c, exceptions.BadRequest("Прикрепите файл изображения."))
		return
	}
	defer file.Close()

	url, err := h.services.Upload.SaveImage("articles", file, header)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *articleHandler) getBySlug(c *gin.Context) {
	article, err := h.services.Article.GetBySlug(c.Param("slug"), viewerFromContext(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, article)
}

func (h *articleHandler) related(c *gin.Context) {
	limit := 3
	if raw := c.Query("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			limit = n
		}
	}

	related, err := h.services.Article.Related(c.Param("slug"), limit, viewerFromContext(c))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, related)
}

func (h *articleHandler) like(c *gin.Context) {
	userID, _ := currentUserID(c)
	result, err := h.services.Article.Like(c.Param("slug"), userID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *articleHandler) unlike(c *gin.Context) {
	userID, _ := currentUserID(c)
	result, err := h.services.Article.Unlike(c.Param("slug"), userID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *articleHandler) save(c *gin.Context) {
	userID, _ := currentUserID(c)
	result, err := h.services.Article.SaveArticle(c.Param("slug"), userID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *articleHandler) unsave(c *gin.Context) {
	userID, _ := currentUserID(c)
	result, err := h.services.Article.UnsaveArticle(c.Param("slug"), userID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *articleHandler) view(c *gin.Context) {
	userID, _ := currentUserID(c)
	if err := h.services.Article.RecordView(c.Param("slug"), userID); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

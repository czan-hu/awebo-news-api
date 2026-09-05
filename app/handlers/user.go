package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"awebo/app/dto"
	"awebo/app/exceptions"
)

type userHandler struct {
	*Handler
}

func newUserHandler(h *Handler) *userHandler {
	return &userHandler{h}
}

func (h *userHandler) RegisterRoutes(router *gin.RouterGroup) {
	users := router.Group("/users")
	{
		users.GET("/me", h.requireAuth(), h.getMe)
		users.PATCH("/me", h.requireAuth(), h.updateMe)
		users.GET("/me/liked", h.requireAuth(), h.liked)
		users.GET("/me/saved", h.requireAuth(), h.saved)
		users.GET("/me/history", h.requireAuth(), h.history)
		users.GET("/:username", h.optionalAuth(), h.publicProfile)
		users.GET("/:username/articles", h.optionalAuth(), h.articlesByUser)
	}
}

func (h *userHandler) getMe(c *gin.Context) {
	userID, _ := currentUserID(c)
	user, err := h.services.User.GetAuthUser(userID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *userHandler) updateMe(c *gin.Context) {
	userID, _ := currentUserID(c)

	var input dto.UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	profile, err := h.services.User.UpdateProfile(userID, input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (h *userHandler) publicProfile(c *gin.Context) {
	profile, err := h.services.User.GetProfileByUsername(c.Param("username"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (h *userHandler) articlesByUser(c *gin.Context) {
	var query dto.PageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	list, err := h.services.Article.ListByAuthor(c.Param("username"), viewerFromContext(c), query.Page, query.Limit)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *userHandler) liked(c *gin.Context) {
	userID, _ := currentUserID(c)

	var query dto.PageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	list, err := h.services.Article.ListLiked(userID, query.Page, query.Limit)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *userHandler) saved(c *gin.Context) {
	userID, _ := currentUserID(c)

	var query dto.PageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	list, err := h.services.Article.ListSaved(userID, query.Page, query.Limit)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *userHandler) history(c *gin.Context) {
	userID, _ := currentUserID(c)

	var query dto.PageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	list, err := h.services.Article.ListHistory(userID, query.Page, query.Limit)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

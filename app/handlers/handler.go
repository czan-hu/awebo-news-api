package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"awebo/app/exceptions"
	"awebo/app/infrastructure/config"
	"awebo/app/services"
)

// Services bundles every feature service the handlers need. Any handler can
// reach any service through Handler.services (e.g. the user handler calls
// into Article for "my liked/saved/history" lists).
type Services struct {
	Auth           *services.AuthService
	User           *services.UserService
	Article        *services.ArticleService
	Category       *services.CategoryService
	Newsletter     *services.NewsletterService
	Forum          *services.ForumService
	Search         *services.SearchService
	CreatorRequest *services.CreatorRequestService
}

type Handler struct {
	services Services
	cfg      *config.Config
}

func NewHandler(s Services, cfg *config.Config) *Handler {
	return &Handler{services: s, cfg: cfg}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), corsMiddleware())

	api := router.Group(h.cfg.APIPrefix)

	newAuthHandler(h).RegisterRoutes(api)
	newUserHandler(h).RegisterRoutes(api)
	newArticleHandler(h).RegisterRoutes(api)
	newCategoryHandler(h).RegisterRoutes(api)
	newNewsletterHandler(h).RegisterRoutes(api)
	newForumHandler(h).RegisterRoutes(api)
	newSearchHandler(h).RegisterRoutes(api)
	newCreatorRequestHandler(h).RegisterRoutes(api)

	return router
}

// respondError renders any error as the `{ statusCode, message, errors? }`
// shape the frontend expects; anything that isn't an *exceptions.AppError is
// treated as an unexpected 500 and logged.
func respondError(c *gin.Context, err error) {
	var appErr *exceptions.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.StatusCode, appErr)
		return
	}
	log.Println("unexpected error:", err)
	c.JSON(http.StatusInternalServerError, exceptions.Internal(""))
}

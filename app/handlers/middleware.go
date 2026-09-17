package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"awebo/app/entities"
	"awebo/app/exceptions"
	"awebo/app/pkg/jwtutil"
)

const userIDContextKey = "userID"

func (h *Handler) parseToken(c *gin.Context) (uuid.UUID, error) {
	header := c.GetHeader("Authorization")
	if header == "" {
		return uuid.Nil, errors.New("no authorization header")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return uuid.Nil, errors.New("malformed authorization header")
	}

	claims, err := jwtutil.Parse(strings.TrimSpace(parts[1]), h.cfg.JWTSecret)
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(claims.UserID)
}

// requireAuth aborts with 401 when there is no valid bearer token.
func (h *Handler) requireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := h.parseToken(c)
		if err != nil {
			respondError(c, exceptions.Unauthorized(""))
			c.Abort()
			return
		}
		c.Set(userIDContextKey, userID)
		c.Next()
	}
}

// optionalAuth attaches the caller's user id when a valid token is present,
// but never rejects the request when it's missing/invalid.
func (h *Handler) optionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if userID, err := h.parseToken(c); err == nil {
			c.Set(userIDContextKey, userID)
		}
		c.Next()
	}
}

// requirePublisher aborts with 403 unless the caller's role allows
// authoring articles (see entities.CanPublishArticles). Must run after
// requireAuth so currentUserID is already set.
func (h *Handler) requirePublisher() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := currentUserID(c)
		if !ok {
			respondError(c, exceptions.Unauthorized(""))
			c.Abort()
			return
		}

		user, err := h.services.User.GetAuthUser(userID)
		if err != nil {
			respondError(c, err)
			c.Abort()
			return
		}

		if !entities.CanPublishArticles(user.Role) {
			respondError(c, exceptions.Forbidden("Публиковать статьи может только автор."))
			c.Abort()
			return
		}

		c.Next()
	}
}

// requireAdmin aborts with 403 unless the caller's role is "admin". Must run
// after requireAuth so currentUserID is already set.
func (h *Handler) requireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := currentUserID(c)
		if !ok {
			respondError(c, exceptions.Unauthorized(""))
			c.Abort()
			return
		}

		user, err := h.services.User.GetAuthUser(userID)
		if err != nil {
			respondError(c, err)
			c.Abort()
			return
		}

		if !entities.IsAdmin(user.Role) {
			respondError(c, exceptions.Forbidden("Доступно только администраторам."))
			c.Abort()
			return
		}

		c.Next()
	}
}

// wsAuth authenticates the WebSocket upgrade handshake. Browsers can't set
// a custom Authorization header on that request, so the token travels as a
// `?token=` query parameter instead — the only route that accepts a token
// this way. Rejects with 401 before any upgrade attempt on failure.
func (h *Handler) wsAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			respondError(c, exceptions.Unauthorized(""))
			c.Abort()
			return
		}

		claims, err := jwtutil.Parse(token, h.cfg.JWTSecret)
		if err != nil {
			respondError(c, exceptions.Unauthorized(""))
			c.Abort()
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			respondError(c, exceptions.Unauthorized(""))
			c.Abort()
			return
		}

		c.Set(userIDContextKey, userID)
		c.Next()
	}
}

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(userIDContextKey)
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

func viewerFromContext(c *gin.Context) *uuid.UUID {
	if id, ok := currentUserID(c); ok {
		return &id
	}
	return nil
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

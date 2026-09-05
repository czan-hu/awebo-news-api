package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"awebo/app/dto"
	"awebo/app/exceptions"
	"awebo/app/pkg/validate"
)

type authHandler struct {
	*Handler
}

func newAuthHandler(h *Handler) *authHandler {
	return &authHandler{h}
}

func (h *authHandler) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/request-code", h.requestCode)
		auth.POST("/verify-code", h.verifyCode)
		auth.POST("/complete-profile", h.requireAuth(), h.completeProfile)
		auth.POST("/logout", h.requireAuth(), h.logout)
	}
}

func (h *authHandler) requestCode(c *gin.Context) {
	var input dto.RequestCodeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	result, err := h.services.Auth.RequestCode(input.Email)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"ttl": result.TTL, "resendAfter": result.ResendAfter})
}

func (h *authHandler) verifyCode(c *gin.Context) {
	var input dto.VerifyCodeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}
	if !validate.Code(input.Code) {
		respondError(c, exceptions.Unauthorized("Неверный код. Проверьте письмо и попробуйте ещё раз."))
		return
	}

	result, err := h.services.Auth.VerifyCode(input.Email, input.Code)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":     result.Token,
		"isNewUser": result.IsNewUser,
		"user":      result.User,
	})
}

func (h *authHandler) completeProfile(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		respondError(c, exceptions.Unauthorized(""))
		return
	}

	var input dto.CompleteProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, exceptions.BadRequest(""))
		return
	}

	user, err := h.services.Auth.CompleteProfile(userID, input.Name, input.Username)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// logout is a no-op beyond 204: tokens are stateless JWTs, so "logging out"
// just means the client discards its copy.
func (h *authHandler) logout(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

package dto

type RequestCodeInput struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyCodeInput struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

type CompleteProfileInput struct {
	Name     string `json:"name" binding:"required"`
	Username string `json:"username" binding:"required"`
}

package dto

type SubscribeInput struct {
	Email string `json:"email" binding:"required,email"`
}

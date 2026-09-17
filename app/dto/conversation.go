package dto

// StartConversationInput is the body for `POST /conversations` — starts (or
// reuses) a 1:1 conversation with the named user.
type StartConversationInput struct {
	Username string `json:"username" binding:"required"`
}

// SendMessageInput is the body for `POST /conversations/{id}/messages`.
type SendMessageInput struct {
	Body string `json:"body" binding:"required"`
}

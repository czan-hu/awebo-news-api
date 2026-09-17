package entities

import "time"

// Conversation is a 1:1 direct-message thread, from the viewer's
// perspective — Other is always "whichever participant isn't the viewer",
// resolved server-side same as Friend/FriendRequest resolve "the other
// user". Group chat isn't exposed in v1, though the underlying schema
// (conversation_participants) is ready for more than two participants.
type Conversation struct {
	ID            string     `json:"id"`
	Other         PublicUser `json:"other"`
	LastMessage   *Message   `json:"lastMessage,omitempty"`
	UnreadCount   int64      `json:"unreadCount"`
	CreatedAt     time.Time  `json:"createdAt"`
}

type ConversationList struct {
	Data []Conversation `json:"data"`
	Meta PageMeta       `json:"meta"`
}

type Message struct {
	ID             string     `json:"id"`
	ConversationID string     `json:"conversationId"`
	Sender         PublicUser `json:"sender"`
	Body           string     `json:"body"`
	CreatedAt      time.Time  `json:"createdAt"`
}

type MessageList struct {
	Data []Message `json:"data"`
	Meta PageMeta  `json:"meta"`
}

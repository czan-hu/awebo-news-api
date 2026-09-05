package entities

import "time"

// Valid CreatorRequest.Status values.
const (
	CreatorRequestPending  = "pending"
	CreatorRequestApproved = "approved"
	CreatorRequestRejected = "rejected"
)

// CreatorRequestSelf is what a user sees about their own latest application
// (GET/POST /users/me/creator-request).
type CreatorRequestSelf struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

// CreatorRequest is the admin-facing view of an application, with enough of
// the applicant's info to review it.
type CreatorRequest struct {
	ID        string     `json:"id"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"createdAt"`
	User      PublicUser `json:"user"`
	Email     string     `json:"email"`
}

type CreatorRequestList struct {
	Data []CreatorRequest `json:"data"`
	Meta PageMeta         `json:"meta"`
}

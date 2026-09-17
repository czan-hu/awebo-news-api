package entities

import "time"

// Valid Friendship.Status values (the underlying request/relationship row).
const (
	FriendshipPending  = "pending"
	FriendshipAccepted = "accepted"
	FriendshipDeclined = "declined"
)

// FriendshipView* values populate UserProfile.FriendshipStatus — the
// relationship between the profile owner and whoever is viewing it, from the
// viewer's perspective. Distinct from the row-status consts above:
// "pending_incoming" vs "pending_outgoing" only makes sense relative to a
// specific viewer, not to the underlying row.
const (
	FriendshipViewNone            = "none"
	FriendshipViewFriends         = "friends"
	FriendshipViewPendingOutgoing = "pending_outgoing"
	FriendshipViewPendingIncoming = "pending_incoming"
)

// Friend is one entry in the caller's friends list.
type Friend struct {
	ID           string     `json:"id"` // friendship row id, used to unfriend via DELETE /friendships/{id}
	User         PublicUser `json:"user"`
	FriendsSince time.Time  `json:"friendsSince"`
}

// FriendRequest is one entry in an incoming/outgoing request list, and the
// response shape for sending/accepting a request.
type FriendRequest struct {
	ID        string     `json:"id"`
	User      PublicUser `json:"user"` // the other party
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"createdAt"`
}

type FriendList struct {
	Data []Friend `json:"data"`
	Meta PageMeta `json:"meta"`
}

type FriendRequestList struct {
	Data []FriendRequest `json:"data"`
	Meta PageMeta        `json:"meta"`
}

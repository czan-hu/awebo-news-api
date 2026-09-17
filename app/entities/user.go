package entities

import "time"

// AuthUser is what gets placed into the client session after login,
// see `AuthUser` in docs/openapi.yaml. Role is the system role
// (user/creator/editor/admin) — not to be confused with UserProfile.Role,
// which is the public display title (e.g. "Автор · Технологии").
type AuthUser struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// CanPublishArticles reports whether a system role is allowed to author
// articles via POST /articles.
func CanPublishArticles(role string) bool {
	switch role {
	case "creator", "editor", "admin":
		return true
	default:
		return false
	}
}

// IsAdmin gates the /admin/* moderation endpoints.
func IsAdmin(role string) bool {
	return role == "admin"
}

// PublicUser is the compact representation used across the forum
// (topic/comment author).
type PublicUser struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

type UserLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
	Icon  string `json:"icon"`
}

type UserStats struct {
	Articles int64 `json:"articles"`
	Liked    int64 `json:"liked"`
	Saved    int64 `json:"saved"`
	History  int64 `json:"history"`
}

type UserProfile struct {
	Name     string     `json:"name"`
	Username string     `json:"username"`
	Role     string     `json:"role"`
	Avatar   string     `json:"avatar"`
	Bio      string      `json:"bio"`
	Location string     `json:"location"`
	JoinedAt time.Time  `json:"joinedAt"`
	Links    []UserLink `json:"links"`
	Stats    UserStats  `json:"stats"`
	// FriendshipStatus is the viewer's relationship to this profile — one of
	// the FriendshipView* consts. Nil when viewing your own profile or when
	// logged out (no viewer to compute a relationship for).
	FriendshipStatus *string `json:"friendshipStatus,omitempty"`
}

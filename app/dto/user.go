package dto

type UserLinkInput struct {
	Label string `json:"label" binding:"required"`
	URL   string `json:"url" binding:"required"`
	Icon  string `json:"icon" binding:"required"`
}

// UpdateProfileInput is a partial update: nil pointers mean "leave unchanged".
// Links, when present (even as an empty array), fully replaces the list.
type UpdateProfileInput struct {
	Name     *string         `json:"name"`
	Username *string         `json:"username"`
	Bio      *string         `json:"bio"`
	Location *string         `json:"location"`
	Avatar   *string         `json:"avatar"`
	Links    []UserLinkInput `json:"links"`
}

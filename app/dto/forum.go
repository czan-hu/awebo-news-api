package dto

type CreateTopicInput struct {
	Title    string `json:"title" binding:"required"`
	Category string `json:"category" binding:"required"`
	Body     string `json:"body" binding:"required"`
}

type CreateCommentInput struct {
	Body     string  `json:"body" binding:"required"`
	ParentID *string `json:"parentId"`
}

// Direction uses `oneof` instead of `required` because 0 (remove vote) is a
// valid value that `required` would otherwise reject as "empty".
type VoteInput struct {
	Direction int `json:"direction" binding:"oneof=-1 0 1"`
}

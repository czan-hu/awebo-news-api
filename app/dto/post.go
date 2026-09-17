package dto

// CreatePostInput is the body for `POST /posts` — a lightweight status
// update, unlike the block-based CreateArticleInput. Image is a path
// returned by a prior `POST /posts/images` upload, same flow as articles.
type CreatePostInput struct {
	Body  string `json:"body" binding:"required"`
	Image string `json:"image"`
}

// CreatePostCommentInput is the body for `POST /posts/{id}/comments`. Named
// distinctly from forum.go's CreateCommentInput (same package) — post
// comments are flat, with no parentId/threading.
type CreatePostCommentInput struct {
	Body string `json:"body" binding:"required"`
}

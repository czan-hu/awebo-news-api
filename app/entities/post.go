package entities

import "time"

// PostSummary is the JSON shape of a feed/wall post. Unlike articles, posts
// carry no separate content-block body, so summary and detail share one
// shape — see the Post alias below.
type PostSummary struct {
	ID           string     `json:"id"`
	Author       PublicUser `json:"author"`
	Body         string     `json:"body"`
	Image        string     `json:"image,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	LikeCount    int64      `json:"likeCount"`
	CommentCount int64      `json:"commentCount"`
	// Liked is nil for anonymous viewers, matching ArticleSummary.Liked.
	Liked *bool `json:"liked,omitempty"`
}

// Post is an alias for PostSummary: a single post has nothing extra to show
// beyond what the feed already renders.
type Post = PostSummary

type PostList struct {
	Data []PostSummary `json:"data"`
	Meta PageMeta      `json:"meta"`
}

type PostComment struct {
	ID        string     `json:"id"`
	PostID    string     `json:"postId"`
	Author    PublicUser `json:"author"`
	Body      string     `json:"body"`
	CreatedAt time.Time  `json:"createdAt"`
}

type PostCommentList struct {
	Data []PostComment `json:"data"`
	Meta PageMeta      `json:"meta"`
}

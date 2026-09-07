package dto

import "awebo/app/entities"

// CreateArticleInput is the body for `POST /articles` — only available to
// users whose role passes entities.CanPublishArticles.
type CreateArticleInput struct {
	Title    string                  `json:"title" binding:"required"`
	Excerpt  string                  `json:"excerpt" binding:"required"`
	Category string                  `json:"category" binding:"required"`
	Cover    string                  `json:"cover"`
	Tags     []string                `json:"tags"`
	Content  []entities.ContentBlock `json:"content" binding:"required"`
}

// UpdateArticleInput is the body for `PUT /articles/{slug}`. Same shape as
// CreateArticleInput; the author (or an admin) may edit every field except
// the slug, which stays fixed so existing links keep working.
type UpdateArticleInput = CreateArticleInput

// ArticleListQuery binds `GET /articles` query parameters.
type ArticleListQuery struct {
	Category string `form:"category"`
	Tag      string `form:"tag"`
	Author   string `form:"author"`
	Featured *bool  `form:"featured"`
	Sort     string `form:"sort"`
	Page     int    `form:"page,default=1"`
	Limit    int    `form:"limit,default=12"`
}

type PageQuery struct {
	Page  int `form:"page,default=1"`
	Limit int `form:"limit,default=12"`
}

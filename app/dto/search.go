package dto

type SearchQuery struct {
	Q     string `form:"q" binding:"required"`
	Limit int    `form:"limit,default=10"`
}

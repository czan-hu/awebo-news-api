package entities

import "math"

// PageMeta mirrors the `PageMeta` schema used by every paginated list response.
type PageMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

func NewPageMeta(page, limit int, total int64) PageMeta {
	totalPages := 0
	if limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}
	return PageMeta{Page: page, Limit: limit, Total: total, TotalPages: totalPages}
}

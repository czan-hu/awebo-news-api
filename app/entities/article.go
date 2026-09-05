package entities

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Author is the byline shown on an article card/detail page.
type Author struct {
	Name   string `json:"name"`
	Role   string `json:"role"`
	Avatar string `json:"avatar"`
}

// ContentBlock is a flattened representation of the `ContentBlock` oneOf
// union from docs/openapi.yaml (p, h2, quote, ul, image, callout). Only the
// fields relevant to `Type` are populated when marshalled.
type ContentBlock struct {
	Type    string   `json:"type"`
	Text    string   `json:"text,omitempty"`
	Cite    string   `json:"cite,omitempty"`
	Items   []string `json:"items,omitempty"`
	Src     string   `json:"src,omitempty"`
	Caption string   `json:"caption,omitempty"`
	Emoji   string   `json:"emoji,omitempty"`
}

type ArticleSummary struct {
	ID             int       `json:"id"`
	Slug           string    `json:"slug"`
	Title          string    `json:"title"`
	Excerpt        string    `json:"excerpt"`
	Category       string    `json:"category"`
	Author         Author    `json:"author"`
	PublishedAt    time.Time `json:"publishedAt"`
	ReadingMinutes int       `json:"readingMinutes"`
	Cover          string    `json:"cover"`
	Tags           []string  `json:"tags"`
	Featured       bool      `json:"featured"`
	Liked          *bool     `json:"liked,omitempty"`
	Saved          *bool     `json:"saved,omitempty"`
}

type Article struct {
	ArticleSummary
	Content []ContentBlock `json:"content"`
}

type ArticleList struct {
	Data []ArticleSummary `json:"data"`
	Meta PageMeta         `json:"meta"`
}

type ArticleInteraction struct {
	Slug      string `json:"slug"`
	Liked     bool   `json:"liked"`
	Saved     bool   `json:"saved"`
	LikeCount int64  `json:"likeCount"`
}

// ValidateContentBlocks checks that every block carries the fields its type
// requires, mirroring the oneOf constraints from docs/openapi.yaml.
func ValidateContentBlocks(blocks []ContentBlock) error {
	if len(blocks) == 0 {
		return fmt.Errorf("добавьте хотя бы один блок контента")
	}
	for i, b := range blocks {
		switch b.Type {
		case "p", "h2":
			if strings.TrimSpace(b.Text) == "" {
				return fmt.Errorf("блок %d: текст не может быть пустым", i+1)
			}
		case "quote":
			if strings.TrimSpace(b.Text) == "" {
				return fmt.Errorf("блок %d: цитата не может быть пустой", i+1)
			}
		case "ul":
			if len(b.Items) == 0 {
				return fmt.Errorf("блок %d: список не может быть пустым", i+1)
			}
		case "image":
			if strings.TrimSpace(b.Src) == "" {
				return fmt.Errorf("блок %d: укажите изображение", i+1)
			}
		case "callout":
			if strings.TrimSpace(b.Emoji) == "" || strings.TrimSpace(b.Text) == "" {
				return fmt.Errorf("блок %d: заполните эмодзи и текст выноски", i+1)
			}
		default:
			return fmt.Errorf("блок %d: неизвестный тип %q", i+1, b.Type)
		}
	}
	return nil
}

// EstimateReadingMinutes counts words across all text-bearing fields and
// converts them to a reading time at ~200 words per minute (min. 1 minute).
func EstimateReadingMinutes(blocks []ContentBlock) int {
	words := 0
	for _, b := range blocks {
		words += len(strings.Fields(b.Text))
		words += len(strings.Fields(b.Caption))
		for _, item := range b.Items {
			words += len(strings.Fields(item))
		}
	}
	minutes := int(math.Ceil(float64(words) / 200))
	if minutes < 1 {
		minutes = 1
	}
	return minutes
}

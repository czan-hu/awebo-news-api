package entities

import "time"

// ForumCategories is the fixed enum of forum sections (`ForumCategory` schema).
var ForumCategories = []string{"Общее", "Технологии", "Общество", "Культура", "Наука", "Мнения"}

func IsValidForumCategory(category string) bool {
	for _, c := range ForumCategories {
		if c == category {
			return true
		}
	}
	return false
}

type ForumTopicSummary struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Category     string     `json:"category"`
	Author       PublicUser `json:"author"`
	CreatedAt    time.Time  `json:"createdAt"`
	Pinned       bool       `json:"pinned"`
	CommentCount int64      `json:"commentCount"`
	Excerpt      string     `json:"excerpt"`
}

type ForumTopic struct {
	ForumTopicSummary
	Body string `json:"body"`
}

type ForumTopicList struct {
	Data []ForumTopicSummary `json:"data"`
	Meta PageMeta            `json:"meta"`
}

type ForumComment struct {
	ID        string     `json:"id"`
	TopicID   string     `json:"topicId"`
	ParentID  *string    `json:"parentId"`
	Author    PublicUser `json:"author"`
	Body      string     `json:"body"`
	CreatedAt time.Time  `json:"createdAt"`
	Score     int        `json:"score"`
	MyVote    int        `json:"myVote"`
}

type VoteResult struct {
	ID     string `json:"id"`
	Score  int    `json:"score"`
	MyVote int    `json:"myVote"`
}

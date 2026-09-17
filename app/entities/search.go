package entities

type SearchResult struct {
	Articles []ArticleSummary    `json:"articles"`
	Topics   []ForumTopicSummary `json:"topics"`
	Users    []PublicUser        `json:"users"`
}

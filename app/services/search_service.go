package services

import (
	"awebo/app/entities"
	"awebo/app/exceptions"
	"awebo/app/repositories"
)

type SearchService struct {
	articleRepo *repositories.ArticleRepository
	forumRepo   *repositories.ForumRepository
}

func NewSearchService(articleRepo *repositories.ArticleRepository, forumRepo *repositories.ForumRepository) *SearchService {
	return &SearchService{articleRepo: articleRepo, forumRepo: forumRepo}
}

func (s *SearchService) Search(query string, limit int) (*entities.SearchResult, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}

	articles, err := s.articleRepo.Search(query, limit)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	topics, err := s.forumRepo.SearchTopics(query, limit)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	articleSummaries := make([]entities.ArticleSummary, 0, len(articles))
	for _, a := range articles {
		articleSummaries = append(articleSummaries, articleSummaryOf(a))
	}

	topicSummaries := make([]entities.ForumTopicSummary, 0, len(topics))
	for _, t := range topics {
		topicSummaries = append(topicSummaries, toTopicSummary(t))
	}

	return &entities.SearchResult{Articles: articleSummaries, Topics: topicSummaries}, nil
}

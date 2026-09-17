package services

import (
	"awebo/app/entities"
	"awebo/app/exceptions"
	"awebo/app/repositories"
)

type SearchService struct {
	articleRepo *repositories.ArticleRepository
	forumRepo   *repositories.ForumRepository
	userRepo    *repositories.UserRepository
}

func NewSearchService(articleRepo *repositories.ArticleRepository, forumRepo *repositories.ForumRepository, userRepo *repositories.UserRepository) *SearchService {
	return &SearchService{articleRepo: articleRepo, forumRepo: forumRepo, userRepo: userRepo}
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

	users, err := s.userRepo.Search(query, limit)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	publicUsers := make([]entities.PublicUser, 0, len(users))
	for _, u := range users {
		publicUsers = append(publicUsers, entities.PublicUser{Name: u.Name, Username: usernameOf(u), Avatar: u.AvatarPath})
	}

	return &entities.SearchResult{Articles: articleSummaries, Topics: topicSummaries, Users: publicUsers}, nil
}

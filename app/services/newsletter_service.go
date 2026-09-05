package services

import (
	"awebo/app/exceptions"
	"awebo/app/repositories"
)

type NewsletterService struct {
	repo *repositories.NewsletterRepository
}

func NewNewsletterService(repo *repositories.NewsletterRepository) *NewsletterService {
	return &NewsletterService{repo: repo}
}

func (s *NewsletterService) Subscribe(email string) error {
	subscribed, err := s.repo.IsSubscribed(email)
	if err != nil {
		return exceptions.Internal("")
	}
	if subscribed {
		return exceptions.Conflict("Email уже подписан.")
	}
	if err := s.repo.Subscribe(email); err != nil {
		return exceptions.Internal("")
	}
	return nil
}

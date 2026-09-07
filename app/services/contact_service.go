package services

import (
	"awebo/app/entities"
	"awebo/app/exceptions"
	"awebo/app/repositories"
)

type ContactService struct {
	repo *repositories.ContactRepository
}

func NewContactService(repo *repositories.ContactRepository) *ContactService {
	return &ContactService{repo: repo}
}

func (s *ContactService) List() ([]entities.Contact, error) {
	contacts, err := s.repo.List()
	if err != nil {
		return nil, exceptions.Internal("")
	}

	result := make([]entities.Contact, 0, len(contacts))
	for _, c := range contacts {
		result = append(result, entities.Contact{Label: c.Label, Href: c.Href, Icon: c.Icon})
	}
	return result, nil
}

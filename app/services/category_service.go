package services

import (
	"awebo/app/entities"
	"awebo/app/exceptions"
	"awebo/app/repositories"
)

type CategoryService struct {
	repo *repositories.CategoryRepository
}

func NewCategoryService(repo *repositories.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) List() ([]entities.Category, error) {
	categories, err := s.repo.List()
	if err != nil {
		return nil, exceptions.Internal("")
	}

	result := make([]entities.Category, 0, len(categories))
	for _, c := range categories {
		result = append(result, entities.Category{Slug: c.Slug, Title: c.Title, Icon: c.Icon})
	}
	return result, nil
}

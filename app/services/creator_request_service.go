package services

import (
	"github.com/google/uuid"

	"awebo/app/entities"
	"awebo/app/exceptions"
	"awebo/app/infrastructure/database/models"
	"awebo/app/repositories"
)

type CreatorRequestService struct {
	repo     *repositories.CreatorRequestRepository
	userRepo *repositories.UserRepository
}

func NewCreatorRequestService(repo *repositories.CreatorRequestRepository, userRepo *repositories.UserRepository) *CreatorRequestService {
	return &CreatorRequestService{repo: repo, userRepo: userRepo}
}

// Submit files a "become a creator" application for review by an admin.
func (s *CreatorRequestService) Submit(userID uuid.UUID) (*entities.CreatorRequestSelf, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if user == nil {
		return nil, exceptions.Unauthorized("")
	}
	if !user.ProfileCompleted {
		return nil, exceptions.Conflict("Сначала заполните профиль.")
	}
	if entities.CanPublishArticles(user.Role) {
		return nil, exceptions.Conflict("У вас уже есть право публикации.")
	}

	pending, err := s.repo.HasPending(userID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if pending {
		return nil, exceptions.Conflict("Заявка уже отправлена и ожидает рассмотрения.")
	}

	created, err := s.repo.Create(userID)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	return &entities.CreatorRequestSelf{ID: created.ID.String(), Status: created.Status, CreatedAt: created.CreatedAt}, nil
}

// LatestForUser powers "what's the state of my application" on the write page.
func (s *CreatorRequestService) LatestForUser(userID uuid.UUID) (*entities.CreatorRequestSelf, error) {
	req, err := s.repo.LatestForUser(userID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if req == nil {
		return nil, nil
	}
	return &entities.CreatorRequestSelf{ID: req.ID.String(), Status: req.Status, CreatedAt: req.CreatedAt}, nil
}

func (s *CreatorRequestService) ListPending(page, limit int) (*entities.CreatorRequestList, error) {
	page, limit = normalizePage(page, limit)

	requests, total, err := s.repo.ListPending(page, limit)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	data := make([]entities.CreatorRequest, 0, len(requests))
	for _, r := range requests {
		data = append(data, toCreatorRequest(r))
	}

	return &entities.CreatorRequestList{Data: data, Meta: entities.NewPageMeta(page, limit, total)}, nil
}

func (s *CreatorRequestService) Approve(reviewerID, requestID uuid.UUID) error {
	return s.decide(reviewerID, requestID, entities.CreatorRequestApproved)
}

func (s *CreatorRequestService) Reject(reviewerID, requestID uuid.UUID) error {
	return s.decide(reviewerID, requestID, entities.CreatorRequestRejected)
}

func (s *CreatorRequestService) decide(reviewerID, requestID uuid.UUID, status string) error {
	req, err := s.repo.FindByID(requestID)
	if err != nil {
		return exceptions.Internal("")
	}
	if req == nil {
		return exceptions.NotFound("Заявка не найдена.")
	}
	if req.Status != entities.CreatorRequestPending {
		return exceptions.Conflict("Заявка уже рассмотрена.")
	}

	if err := s.repo.Review(requestID, reviewerID, status); err != nil {
		return exceptions.Internal("")
	}

	if status == entities.CreatorRequestApproved {
		if _, err := s.userRepo.UpdateFields(req.UserID, map[string]any{"role": "creator"}); err != nil {
			return exceptions.Internal("")
		}
	}

	return nil
}

func toCreatorRequest(r models.CreatorRequest) entities.CreatorRequest {
	return entities.CreatorRequest{
		ID:        r.ID.String(),
		Status:    r.Status,
		CreatedAt: r.CreatedAt,
		User:      entities.PublicUser{Name: r.User.Name, Username: usernameOf(r.User), Avatar: r.User.AvatarPath},
		Email:     r.User.Email,
	}
}

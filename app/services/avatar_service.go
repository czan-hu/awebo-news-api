package services

import (
	"mime/multipart"

	"github.com/google/uuid"

	"awebo/app/entities"
	"awebo/app/exceptions"
	"awebo/app/repositories"
)

type AvatarService struct {
	repo    *repositories.UserRepository
	uploads *UploadService
}

func NewAvatarService(repo *repositories.UserRepository, uploads *UploadService) *AvatarService {
	return &AvatarService{repo: repo, uploads: uploads}
}

// Upload validates, stores, and immediately applies a new avatar image for
// userID, returning the updated session user.
func (s *AvatarService) Upload(userID uuid.UUID, file multipart.File, header *multipart.FileHeader) (*entities.AuthUser, error) {
	url, err := s.uploads.SaveImage("avatars", file, header)
	if err != nil {
		return nil, err
	}

	updated, err := s.repo.UpdateFields(userID, map[string]any{"avatar_path": url})
	if err != nil {
		return nil, exceptions.Internal("")
	}

	return toAuthUser(updated), nil
}

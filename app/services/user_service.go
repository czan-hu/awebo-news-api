package services

import (
	"github.com/google/uuid"

	"awebo/app/dto"
	"awebo/app/entities"
	"awebo/app/exceptions"
	"awebo/app/infrastructure/database/models"
	"awebo/app/pkg/validate"
	"awebo/app/repositories"
)

type UserService struct {
	repo           *repositories.UserRepository
	friendshipRepo *repositories.FriendshipRepository
}

func NewUserService(repo *repositories.UserRepository, friendshipRepo *repositories.FriendshipRepository) *UserService {
	return &UserService{repo: repo, friendshipRepo: friendshipRepo}
}

func (s *UserService) GetAuthUser(userID uuid.UUID) (*entities.AuthUser, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if user == nil {
		return nil, exceptions.Unauthorized("")
	}
	return toAuthUser(user), nil
}

func (s *UserService) GetProfileByUsername(username string, viewerID *uuid.UUID) (*entities.UserProfile, error) {
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if user == nil {
		return nil, exceptions.NotFound("Пользователь не найден.")
	}
	return s.buildProfile(user, viewerID)
}

func (s *UserService) UpdateProfile(userID uuid.UUID, input dto.UpdateProfileInput) (*entities.UserProfile, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if user == nil {
		return nil, exceptions.Unauthorized("")
	}

	fields := map[string]any{}

	if input.Username != nil {
		username := *input.Username
		if !validate.Username(username) {
			return nil, exceptions.ValidationError("Проверьте правильность заполнения полей.", map[string]string{
				"username": "Никнейм: 3-20 символов, латиница, цифры и подчёркивание.",
			})
		}
		taken, err := s.repo.IsUsernameTakenByOther(username, userID)
		if err != nil {
			return nil, exceptions.Internal("")
		}
		if taken {
			return nil, exceptions.Conflict("Никнейм занят.")
		}
		fields["username"] = username
	}
	if input.Name != nil {
		fields["name"] = *input.Name
	}
	if input.Bio != nil {
		fields["bio"] = *input.Bio
	}
	if input.Location != nil {
		fields["location"] = *input.Location
	}
	if input.Avatar != nil {
		fields["avatar_path"] = *input.Avatar
	}

	updated, err := s.repo.UpdateFields(userID, fields)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	if input.Links != nil {
		links := make([]models.UserLink, 0, len(input.Links))
		for i, l := range input.Links {
			links = append(links, models.UserLink{
				ID:       uuid.New(),
				UserID:   userID,
				Label:    l.Label,
				URL:      l.URL,
				Icon:     l.Icon,
				Position: i,
			})
		}
		if err := s.repo.ReplaceLinks(userID, links); err != nil {
			return nil, exceptions.Internal("")
		}
	}

	return s.buildProfile(updated, nil)
}

func (s *UserService) buildProfile(user *models.User, viewerID *uuid.UUID) (*entities.UserProfile, error) {
	links, err := s.repo.ListLinks(user.ID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	stats, err := s.repo.Stats(user.ID)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	entityLinks := make([]entities.UserLink, 0, len(links))
	for _, l := range links {
		entityLinks = append(entityLinks, entities.UserLink{Label: l.Label, URL: l.URL, Icon: l.Icon})
	}

	var friendshipStatus *string
	if viewerID != nil && *viewerID != user.ID {
		status, err := s.resolveFriendshipView(*viewerID, user.ID)
		if err != nil {
			return nil, exceptions.Internal("")
		}
		friendshipStatus = &status
	}

	return &entities.UserProfile{
		Name:     user.Name,
		Username: usernameOf(*user),
		Role:     user.Title,
		Avatar:   user.AvatarPath,
		Bio:      user.Bio,
		Location: user.Location,
		JoinedAt: user.CreatedAt,
		Links:    entityLinks,
		Stats: entities.UserStats{
			Articles: stats.Articles,
			Liked:    stats.Liked,
			Saved:    stats.Saved,
			History:  stats.History,
		},
		FriendshipStatus: friendshipStatus,
	}, nil
}

// resolveFriendshipView reports the viewer's relationship to profileID, as
// one of the entities.FriendshipView* consts.
func (s *UserService) resolveFriendshipView(viewerID, profileID uuid.UUID) (string, error) {
	f, err := s.friendshipRepo.FindPair(viewerID, profileID)
	if err != nil {
		return "", err
	}
	if f == nil {
		return entities.FriendshipViewNone, nil
	}
	switch f.Status {
	case entities.FriendshipAccepted:
		return entities.FriendshipViewFriends, nil
	case entities.FriendshipPending:
		if f.RequesterID == viewerID {
			return entities.FriendshipViewPendingOutgoing, nil
		}
		return entities.FriendshipViewPendingIncoming, nil
	default:
		return entities.FriendshipViewNone, nil
	}
}

func usernameOf(u models.User) string {
	if u.Username != nil {
		return *u.Username
	}
	return ""
}

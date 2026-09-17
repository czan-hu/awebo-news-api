package services

import (
	"github.com/google/uuid"

	"awebo/app/entities"
	"awebo/app/exceptions"
	"awebo/app/infrastructure/database/models"
	"awebo/app/repositories"
)

type FriendshipService struct {
	repo     *repositories.FriendshipRepository
	userRepo *repositories.UserRepository
}

func NewFriendshipService(repo *repositories.FriendshipRepository, userRepo *repositories.UserRepository) *FriendshipService {
	return &FriendshipService{repo: repo, userRepo: userRepo}
}

// SendRequest files a friend request from requesterID to the user with the
// given username. A pending/accepted row already existing in either
// direction is a conflict — the unique index would reject a duplicate insert
// anyway, but checking first gives a clear message instead of a raw DB error.
func (s *FriendshipService) SendRequest(requesterID uuid.UUID, addresseeUsername string) (*entities.FriendRequest, *exceptions.AppError) {
	addressee, err := s.userRepo.FindByUsername(addresseeUsername)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if addressee == nil {
		return nil, exceptions.NotFound("Пользователь не найден.")
	}
	if addressee.ID == requesterID {
		return nil, exceptions.BadRequest("Нельзя добавить себя в друзья.")
	}

	existing, err := s.repo.FindPair(requesterID, addressee.ID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if existing != nil {
		if existing.Status == entities.FriendshipAccepted {
			return nil, exceptions.Conflict("Вы уже друзья.")
		}
		return nil, exceptions.Conflict("Заявка уже отправлена и ожидает решения.")
	}

	created, err := s.repo.Create(requesterID, addressee.ID)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	full, err := s.repo.FindByID(created.ID)
	if err != nil || full == nil {
		return nil, exceptions.Internal("")
	}

	return toFriendRequest(*full, requesterID), nil
}

func (s *FriendshipService) Accept(userID, requestID uuid.UUID) (*entities.FriendRequest, *exceptions.AppError) {
	f, appErr := s.mustPending(requestID)
	if appErr != nil {
		return nil, appErr
	}
	if f.AddresseeID != userID {
		return nil, exceptions.Forbidden("Только адресат может принять заявку.")
	}

	if err := s.repo.UpdateStatus(requestID, entities.FriendshipAccepted); err != nil {
		return nil, exceptions.Internal("")
	}

	full, err := s.repo.FindByID(requestID)
	if err != nil || full == nil {
		return nil, exceptions.Internal("")
	}
	return toFriendRequest(*full, userID), nil
}

func (s *FriendshipService) Decline(userID, requestID uuid.UUID) *exceptions.AppError {
	f, appErr := s.mustPending(requestID)
	if appErr != nil {
		return appErr
	}
	if f.AddresseeID != userID {
		return exceptions.Forbidden("Только адресат может отклонить заявку.")
	}

	if err := s.repo.UpdateStatus(requestID, entities.FriendshipDeclined); err != nil {
		return exceptions.Internal("")
	}
	return nil
}

func (s *FriendshipService) mustPending(requestID uuid.UUID) (*models.Friendship, *exceptions.AppError) {
	f, err := s.repo.FindByID(requestID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if f == nil {
		return nil, exceptions.NotFound("Заявка не найдена.")
	}
	if f.Status != entities.FriendshipPending {
		return nil, exceptions.Conflict("Заявка уже рассмотрена.")
	}
	return f, nil
}

// Unfriend deletes a friendship row. It doubles as "cancel my own pending
// request": the requester may delete a pending row they created (declining
// is reserved for the addressee, via Decline), while an accepted row may be
// deleted by either party to unfriend.
func (s *FriendshipService) Unfriend(userID, friendshipID uuid.UUID) *exceptions.AppError {
	f, err := s.repo.FindByID(friendshipID)
	if err != nil {
		return exceptions.Internal("")
	}
	if f == nil {
		return exceptions.NotFound("Дружба не найдена.")
	}
	if f.RequesterID != userID && f.AddresseeID != userID {
		return exceptions.Forbidden("")
	}

	switch f.Status {
	case entities.FriendshipAccepted:
		// either party may unfriend, already checked above
	case entities.FriendshipPending:
		if f.RequesterID != userID {
			return exceptions.Forbidden("Отменить заявку может только отправитель.")
		}
	default:
		return exceptions.Conflict("Вы не друзья.")
	}

	if err := s.repo.Delete(friendshipID); err != nil {
		return exceptions.Internal("")
	}
	return nil
}

func (s *FriendshipService) ListFriends(userID uuid.UUID, page, limit int) (*entities.FriendList, *exceptions.AppError) {
	page, limit = normalizePage(page, limit)

	friendships, total, err := s.repo.ListFriends(userID, page, limit)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	data := make([]entities.Friend, 0, len(friendships))
	for _, f := range friendships {
		data = append(data, toFriend(f, userID))
	}

	return &entities.FriendList{Data: data, Meta: entities.NewPageMeta(page, limit, total)}, nil
}

func (s *FriendshipService) ListIncoming(userID uuid.UUID, page, limit int) (*entities.FriendRequestList, *exceptions.AppError) {
	page, limit = normalizePage(page, limit)

	friendships, total, err := s.repo.ListIncoming(userID, page, limit)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	data := make([]entities.FriendRequest, 0, len(friendships))
	for _, f := range friendships {
		data = append(data, *toFriendRequest(f, userID))
	}

	return &entities.FriendRequestList{Data: data, Meta: entities.NewPageMeta(page, limit, total)}, nil
}

func (s *FriendshipService) ListOutgoing(userID uuid.UUID, page, limit int) (*entities.FriendRequestList, *exceptions.AppError) {
	page, limit = normalizePage(page, limit)

	friendships, total, err := s.repo.ListOutgoing(userID, page, limit)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	data := make([]entities.FriendRequest, 0, len(friendships))
	for _, f := range friendships {
		data = append(data, *toFriendRequest(f, userID))
	}

	return &entities.FriendRequestList{Data: data, Meta: entities.NewPageMeta(page, limit, total)}, nil
}

// FriendIDs is a thin wrapper over the repository, used by PostService to
// scope the home feed to the viewer plus their accepted friends.
func (s *FriendshipService) FriendIDs(userID uuid.UUID) ([]uuid.UUID, *exceptions.AppError) {
	ids, err := s.repo.FriendIDs(userID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	return ids, nil
}

// toFriend resolves "the other user" (whichever side isn't viewerID) into a
// PublicUser-shaped entry for the caller's friends list.
func toFriend(f models.Friendship, viewerID uuid.UUID) entities.Friend {
	other := f.Requester
	if f.RequesterID == viewerID {
		other = f.Addressee
	}
	return entities.Friend{
		ID:           f.ID.String(),
		User:         entities.PublicUser{Name: other.Name, Username: usernameOf(other), Avatar: other.AvatarPath},
		FriendsSince: f.UpdatedAt,
	}
}

// toFriendRequest mirrors toFriend for incoming/outgoing request lists and
// single-request responses (send/accept).
func toFriendRequest(f models.Friendship, viewerID uuid.UUID) *entities.FriendRequest {
	other := f.Requester
	if f.RequesterID == viewerID {
		other = f.Addressee
	}
	return &entities.FriendRequest{
		ID:        f.ID.String(),
		User:      entities.PublicUser{Name: other.Name, Username: usernameOf(other), Avatar: other.AvatarPath},
		Status:    f.Status,
		CreatedAt: f.CreatedAt,
	}
}

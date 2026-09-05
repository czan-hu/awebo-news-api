package services

import (
	"crypto/subtle"
	"time"

	"github.com/google/uuid"

	"awebo/app/entities"
	"awebo/app/exceptions"
	"awebo/app/infrastructure/config"
	"awebo/app/infrastructure/database/models"
	"awebo/app/pkg/idgen"
	"awebo/app/pkg/jwtutil"
	"awebo/app/pkg/validate"
	"awebo/app/repositories"
)

type AuthService struct {
	repo   *repositories.AuthRepository
	cfg    *config.Config
	mailer Mailer
}

func NewAuthService(repo *repositories.AuthRepository, cfg *config.Config, mailer Mailer) *AuthService {
	return &AuthService{repo: repo, cfg: cfg, mailer: mailer}
}

type RequestCodeResult struct {
	TTL         int
	ResendAfter int
}

func (s *AuthService) RequestCode(email string) (*RequestCodeResult, error) {
	existing, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if existing != nil && existing.CodeRequestedAt != nil {
		nextAllowed := existing.CodeRequestedAt.Add(s.cfg.CodeResendTTL)
		if time.Now().Before(nextAllowed) {
			return nil, exceptions.TooManyRequests("")
		}
	}

	code := idgen.RandomCode()
	now := time.Now()
	expiresAt := now.Add(s.cfg.CodeTTL)

	if _, err := s.repo.UpsertVerificationCode(email, code, expiresAt, now); err != nil {
		return nil, exceptions.Internal("")
	}

	if err := s.mailer.SendVerificationCode(email, code); err != nil {
		return nil, exceptions.Internal("Не удалось отправить письмо.")
	}

	return &RequestCodeResult{
		TTL:         int(s.cfg.CodeTTL.Seconds()),
		ResendAfter: int(s.cfg.CodeResendTTL.Seconds()),
	}, nil
}

type VerifyCodeResult struct {
	Token     string
	IsNewUser bool
	User      *entities.AuthUser
}

func (s *AuthService) VerifyCode(email, code string) (*VerifyCodeResult, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	invalidCodeErr := exceptions.Unauthorized("Неверный код. Проверьте письмо и попробуйте ещё раз.")

	if user == nil || user.VerificationCode == nil || user.VerificationCodeExpiresAt == nil {
		return nil, invalidCodeErr
	}
	if time.Now().After(*user.VerificationCodeExpiresAt) {
		return nil, invalidCodeErr
	}
	if subtle.ConstantTimeCompare([]byte(*user.VerificationCode), []byte(code)) != 1 {
		return nil, invalidCodeErr
	}

	if err := s.repo.ClearVerificationCode(user.ID); err != nil {
		return nil, exceptions.Internal("")
	}

	token, err := jwtutil.Generate(user.ID.String(), s.cfg.JWTTTL, s.cfg.JWTSecret)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	if !user.ProfileCompleted {
		return &VerifyCodeResult{Token: token, IsNewUser: true, User: nil}, nil
	}

	return &VerifyCodeResult{
		Token:     token,
		IsNewUser: false,
		User:      toAuthUser(user),
	}, nil
}

func (s *AuthService) CompleteProfile(userID uuid.UUID, name, username string) (*entities.AuthUser, error) {
	if !validate.Username(username) {
		return nil, exceptions.ValidationError("Проверьте правильность заполнения полей.", map[string]string{
			"username": "Никнейм: 3-20 символов, латиница, цифры и подчёркивание.",
		})
	}

	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if user == nil {
		return nil, exceptions.Unauthorized("")
	}
	if user.ProfileCompleted {
		return nil, exceptions.Conflict("Профиль уже заполнен.")
	}

	taken, err := s.repo.IsUsernameTaken(username)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if taken {
		return nil, exceptions.Conflict("Этот никнейм уже занят.")
	}

	updated, err := s.repo.CompleteProfile(userID, name, username)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	return toAuthUser(updated), nil
}

func toAuthUser(u *models.User) *entities.AuthUser {
	return &entities.AuthUser{
		Name:     u.Name,
		Username: usernameOf(*u),
		Avatar:   u.AvatarPath,
		Email:    u.Email,
		Role:     u.Role,
	}
}

package service

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/sreagent/sreagent/internal/config"
	"github.com/sreagent/sreagent/internal/middleware"
	"github.com/sreagent/sreagent/internal/model"
	apperr "github.com/sreagent/sreagent/internal/pkg/errors"
	"github.com/sreagent/sreagent/internal/repository"
)

type AuthService struct {
	userRepo *repository.UserRepository
	jwtCfg   *config.JWTConfig
	logger   *zap.Logger
}

func NewAuthService(userRepo *repository.UserRepository, jwtCfg *config.JWTConfig, logger *zap.Logger) *AuthService {
	return &AuthService{userRepo: userRepo, jwtCfg: jwtCfg, logger: logger}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, int, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", 0, apperr.ErrInvalidCreds
	}

	if !user.IsActive {
		return "", 0, apperr.WithMessage(apperr.ErrForbidden, "account is disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", 0, apperr.ErrInvalidCreds
	}

	token, err := middleware.GenerateToken(user.ID, user.Username, string(user.Role), s.jwtCfg.Secret, s.jwtCfg.Expire)
	if err != nil {
		s.logger.Error("failed to generate token", zap.Error(err))
		return "", 0, apperr.Wrap(apperr.ErrInternal, err)
	}

	return token, s.jwtCfg.Expire, nil
}

func (s *AuthService) GetProfile(ctx context.Context, userID uint) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperr.ErrUserNotFound
	}
	return user, nil
}

// HashPassword hashes a password using bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/sreagent/sreagent/internal/model"
	apperr "github.com/sreagent/sreagent/internal/pkg/errors"
	"github.com/sreagent/sreagent/internal/repository"
)

type UserNotifyConfigService struct {
	repo   *repository.UserNotifyConfigRepository
	logger *zap.Logger
}

func NewUserNotifyConfigService(repo *repository.UserNotifyConfigRepository, logger *zap.Logger) *UserNotifyConfigService {
	return &UserNotifyConfigService{repo: repo, logger: logger}
}

// ListByUserID returns all notify configs for a user.
func (s *UserNotifyConfigService) ListByUserID(ctx context.Context, userID uint) ([]model.UserNotifyConfig, error) {
	return s.repo.ListByUserID(ctx, userID)
}

// Upsert creates or updates a single notify config (one per media_type).
func (s *UserNotifyConfigService) Upsert(ctx context.Context, cfg *model.UserNotifyConfig) error {
	if err := s.repo.Upsert(ctx, cfg); err != nil {
		s.logger.Error("failed to upsert user notify config", zap.Error(err))
		return apperr.Wrap(apperr.ErrDatabase, err)
	}
	return nil
}

// DeleteByMediaType removes a specific media type config.
func (s *UserNotifyConfigService) DeleteByMediaType(ctx context.Context, userID uint, mediaType string) error {
	if err := s.repo.DeleteByMediaType(ctx, userID, mediaType); err != nil {
		s.logger.Error("failed to delete user notify config", zap.Error(err))
		return apperr.Wrap(apperr.ErrDatabase, err)
	}
	return nil
}

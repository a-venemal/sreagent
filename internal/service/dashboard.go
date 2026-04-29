package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/sreagent/sreagent/internal/model"
	apperr "github.com/sreagent/sreagent/internal/pkg/errors"
	"github.com/sreagent/sreagent/internal/repository"
)

type DashboardService struct {
	repo   *repository.DashboardRepository
	logger *zap.Logger
}

func NewDashboardService(repo *repository.DashboardRepository, logger *zap.Logger) *DashboardService {
	return &DashboardService{repo: repo, logger: logger}
}

func (s *DashboardService) Create(ctx context.Context, d *model.Dashboard) error {
	if err := s.repo.Create(ctx, d); err != nil {
		s.logger.Error("failed to create dashboard", zap.Error(err))
		return apperr.Wrap(apperr.ErrDatabase, err)
	}
	return nil
}

func (s *DashboardService) GetByID(ctx context.Context, id uint) (*model.Dashboard, error) {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return d, nil
}

func (s *DashboardService) List(ctx context.Context, search string, page, pageSize int) ([]model.Dashboard, int64, error) {
	return s.repo.List(ctx, search, page, pageSize)
}

func (s *DashboardService) Update(ctx context.Context, d *model.Dashboard) error {
	existing, err := s.repo.GetByID(ctx, d.ID)
	if err != nil {
		return apperr.ErrNotFound
	}

	existing.Name = d.Name
	existing.Description = d.Description
	existing.Tags = d.Tags
	existing.Config = d.Config
	existing.IsPublic = d.IsPublic
	existing.UpdatedBy = d.UpdatedBy

	if err := s.repo.Update(ctx, existing); err != nil {
		s.logger.Error("failed to update dashboard", zap.Error(err))
		return apperr.Wrap(apperr.ErrDatabase, err)
	}
	return nil
}

func (s *DashboardService) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return apperr.ErrNotFound
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete dashboard", zap.Error(err))
		return apperr.Wrap(apperr.ErrDatabase, err)
	}
	return nil
}

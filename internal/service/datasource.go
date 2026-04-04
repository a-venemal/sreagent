package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/sreagent/sreagent/internal/model"
	"github.com/sreagent/sreagent/internal/pkg/datasource"
	apperr "github.com/sreagent/sreagent/internal/pkg/errors"
	"github.com/sreagent/sreagent/internal/repository"
)

type DataSourceService struct {
	repo   *repository.DataSourceRepository
	logger *zap.Logger
}

func NewDataSourceService(repo *repository.DataSourceRepository, logger *zap.Logger) *DataSourceService {
	return &DataSourceService{repo: repo, logger: logger}
}

func (s *DataSourceService) Create(ctx context.Context, ds *model.DataSource) error {
	// Check if name already exists
	existing, _ := s.repo.GetByName(ctx, ds.Name)
	if existing != nil {
		return apperr.WithMessage(apperr.ErrDuplicateName, fmt.Sprintf("datasource '%s' already exists", ds.Name))
	}

	if err := s.repo.Create(ctx, ds); err != nil {
		s.logger.Error("failed to create datasource", zap.Error(err))
		return apperr.Wrap(apperr.ErrDatabase, err)
	}

	return nil
}

func (s *DataSourceService) GetByID(ctx context.Context, id uint) (*model.DataSource, error) {
	ds, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrDSNotFound
	}
	return ds, nil
}

func (s *DataSourceService) List(ctx context.Context, dsType string, page, pageSize int) ([]model.DataSource, int64, error) {
	return s.repo.List(ctx, dsType, page, pageSize)
}

func (s *DataSourceService) Update(ctx context.Context, ds *model.DataSource) error {
	existing, err := s.repo.GetByID(ctx, ds.ID)
	if err != nil {
		return apperr.ErrDSNotFound
	}

	// Update fields
	existing.Name = ds.Name
	existing.Type = ds.Type
	existing.Endpoint = ds.Endpoint
	existing.Description = ds.Description
	existing.Labels = ds.Labels
	existing.AuthType = ds.AuthType
	if ds.AuthConfig != "" {
		existing.AuthConfig = ds.AuthConfig
	}
	existing.HealthCheckInterval = ds.HealthCheckInterval

	if err := s.repo.Update(ctx, existing); err != nil {
		s.logger.Error("failed to update datasource", zap.Error(err))
		return apperr.Wrap(apperr.ErrDatabase, err)
	}

	return nil
}

func (s *DataSourceService) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return apperr.ErrDSNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete datasource", zap.Error(err))
		return apperr.Wrap(apperr.ErrDatabase, err)
	}

	return nil
}

// HealthCheck performs a connectivity check against the datasource.
func (s *DataSourceService) HealthCheck(ctx context.Context, id uint) (model.DataSourceStatus, error) {
	ds, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return model.DSStatusUnknown, apperr.ErrDSNotFound
	}

	checker, err := datasource.NewChecker(string(ds.Type))
	if err != nil {
		s.logger.Warn("unsupported datasource type for health check",
			zap.String("type", string(ds.Type)),
		)
		return model.DSStatusUnknown, nil
	}

	status := model.DSStatusHealthy
	if err := checker.CheckHealth(ctx, ds.Endpoint, ds.AuthType, ds.AuthConfig); err != nil {
		status = model.DSStatusUnhealthy
		s.logger.Warn("datasource health check failed",
			zap.String("datasource", ds.Name),
			zap.Error(err),
		)
	}

	ds.Status = status
	if err := s.repo.Update(ctx, ds); err != nil {
		s.logger.Error("failed to persist datasource health status",
			zap.String("datasource", ds.Name),
			zap.Error(err),
		)
	}

	s.logger.Info("health check completed",
		zap.String("datasource", ds.Name),
		zap.String("status", string(status)),
	)

	return status, nil
}

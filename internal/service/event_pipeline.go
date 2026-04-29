package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/sreagent/sreagent/internal/engine/pipeline"
	"github.com/sreagent/sreagent/internal/model"
	apperr "github.com/sreagent/sreagent/internal/pkg/errors"
	"github.com/sreagent/sreagent/internal/repository"
)

type EventPipelineService struct {
	repo     *repository.EventPipelineRepository
	execRepo *repository.PipelineExecutionRepository
	engine   *pipeline.Engine
	logger   *zap.Logger
}

func NewEventPipelineService(
	repo *repository.EventPipelineRepository,
	execRepo *repository.PipelineExecutionRepository,
	engine *pipeline.Engine,
	logger *zap.Logger,
) *EventPipelineService {
	return &EventPipelineService{
		repo:     repo,
		execRepo: execRepo,
		engine:   engine,
		logger:   logger,
	}
}

func (s *EventPipelineService) Create(ctx context.Context, p *model.EventPipeline) error {
	if err := s.repo.Create(ctx, p); err != nil {
		s.logger.Error("failed to create event pipeline", zap.Error(err))
		return apperr.Wrap(apperr.ErrDatabase, err)
	}
	return nil
}

func (s *EventPipelineService) GetByID(ctx context.Context, id uint) (*model.EventPipeline, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	return p, nil
}

func (s *EventPipelineService) List(ctx context.Context, search string, page, pageSize int) ([]model.EventPipeline, int64, error) {
	return s.repo.List(ctx, search, page, pageSize)
}

func (s *EventPipelineService) Update(ctx context.Context, p *model.EventPipeline) error {
	existing, err := s.repo.GetByID(ctx, p.ID)
	if err != nil {
		return apperr.ErrNotFound
	}

	existing.Name = p.Name
	existing.Description = p.Description
	existing.Disabled = p.Disabled
	existing.FilterEnable = p.FilterEnable
	existing.LabelFilters = p.LabelFilters
	existing.Nodes = p.Nodes
	existing.Connections = p.Connections
	existing.UpdatedBy = p.UpdatedBy

	if err := s.repo.Update(ctx, existing); err != nil {
		s.logger.Error("failed to update event pipeline", zap.Error(err))
		return apperr.Wrap(apperr.ErrDatabase, err)
	}
	return nil
}

func (s *EventPipelineService) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return apperr.ErrNotFound
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete event pipeline", zap.Error(err))
		return apperr.Wrap(apperr.ErrDatabase, err)
	}
	return nil
}

// TryRun executes a pipeline with a sample event for testing.
func (s *EventPipelineService) TryRun(ctx context.Context, pipelineID uint, sampleEvent *model.AlertEvent) (*pipeline.WorkflowResult, error) {
	p, err := s.repo.GetByID(ctx, pipelineID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}

	result, err := s.engine.Execute(ctx, p, sampleEvent)
	if err != nil {
		s.logger.Error("pipeline tryrun failed", zap.Uint("pipeline_id", pipelineID), zap.Error(err))
		return nil, apperr.Wrap(apperr.ErrExternalAPI, err)
	}

	return result, nil
}

// ListExecutions returns paginated execution records for a pipeline.
func (s *EventPipelineService) ListExecutions(ctx context.Context, pipelineID uint, page, pageSize int) ([]model.PipelineExecution, int64, error) {
	return s.execRepo.ListByPipelineID(ctx, pipelineID, page, pageSize)
}

// ExecuteMatching runs all enabled pipelines against an event.
// Called from the alert engine's onAlertFn callback.
func (s *EventPipelineService) ExecuteMatching(ctx context.Context, event *model.AlertEvent) (*pipeline.WorkflowResult, error) {
	pipelines, err := s.repo.ListAllEnabled(ctx)
	if err != nil {
		s.logger.Error("failed to list pipelines", zap.Error(err))
		return nil, err
	}
	if len(pipelines) == 0 {
		return nil, nil
	}

	// Convert to pointers
	ptrs := make([]*model.EventPipeline, len(pipelines))
	for i := range pipelines {
		ptrs[i] = &pipelines[i]
	}

	return s.engine.ExecuteMatching(ctx, ptrs, event)
}

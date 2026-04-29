package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/sreagent/sreagent/internal/model"
)

type EventPipelineRepository struct {
	db *gorm.DB
}

func NewEventPipelineRepository(db *gorm.DB) *EventPipelineRepository {
	return &EventPipelineRepository{db: db}
}

func (r *EventPipelineRepository) Create(ctx context.Context, p *model.EventPipeline) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *EventPipelineRepository) GetByID(ctx context.Context, id uint) (*model.EventPipeline, error) {
	var p model.EventPipeline
	err := r.db.WithContext(ctx).First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *EventPipelineRepository) List(ctx context.Context, search string, page, pageSize int) ([]model.EventPipeline, int64, error) {
	var list []model.EventPipeline
	var total int64

	query := r.db.WithContext(ctx).Model(&model.EventPipeline{})
	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("updated_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *EventPipelineRepository) ListAllEnabled(ctx context.Context) ([]model.EventPipeline, error) {
	var list []model.EventPipeline
	err := r.db.WithContext(ctx).Where("disabled = ?", false).Find(&list).Error
	return list, err
}

func (r *EventPipelineRepository) Update(ctx context.Context, p *model.EventPipeline) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *EventPipelineRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.EventPipeline{}, id).Error
}

// PipelineExecutionRepository handles pipeline execution record persistence.
type PipelineExecutionRepository struct {
	db *gorm.DB
}

func NewPipelineExecutionRepository(db *gorm.DB) *PipelineExecutionRepository {
	return &PipelineExecutionRepository{db: db}
}

func (r *PipelineExecutionRepository) Create(ctx context.Context, exec *model.PipelineExecution) error {
	return r.db.WithContext(ctx).Create(exec).Error
}

func (r *PipelineExecutionRepository) ListByPipelineID(ctx context.Context, pipelineID uint, page, pageSize int) ([]model.PipelineExecution, int64, error) {
	var list []model.PipelineExecution
	var total int64

	query := r.db.WithContext(ctx).Model(&model.PipelineExecution{}).Where("pipeline_id = ?", pipelineID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("started_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/sreagent/sreagent/internal/model"
)

// AlertChannelRepository handles alert_channels persistence.
type AlertChannelRepository struct {
	db *gorm.DB
}

// NewAlertChannelRepository creates a new AlertChannelRepository.
func NewAlertChannelRepository(db *gorm.DB) *AlertChannelRepository {
	return &AlertChannelRepository{db: db}
}

// Create creates a new alert channel.
func (r *AlertChannelRepository) Create(ctx context.Context, ch *model.AlertChannel) error {
	return r.db.WithContext(ctx).Create(ch).Error
}

// GetByID returns an alert channel by its ID.
func (r *AlertChannelRepository) GetByID(ctx context.Context, id uint) (*model.AlertChannel, error) {
	var ch model.AlertChannel
	err := r.db.WithContext(ctx).First(&ch, id).Error
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

// List returns a paginated list of alert channels.
func (r *AlertChannelRepository) List(ctx context.Context, page, pageSize int) ([]model.AlertChannel, int64, error) {
	var list []model.AlertChannel
	var total int64

	query := r.db.WithContext(ctx).Model(&model.AlertChannel{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Update updates an existing alert channel.
func (r *AlertChannelRepository) Update(ctx context.Context, ch *model.AlertChannel) error {
	return r.db.WithContext(ctx).Save(ch).Error
}

// Delete soft-deletes an alert channel by ID.
func (r *AlertChannelRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.AlertChannel{}, id).Error
}

// ListEnabled returns all enabled alert channels.
func (r *AlertChannelRepository) ListEnabled(ctx context.Context) ([]model.AlertChannel, error) {
	var list []model.AlertChannel
	err := r.db.WithContext(ctx).
		Where("is_enabled = ?", true).
		Order("id ASC").
		Find(&list).Error
	return list, err
}

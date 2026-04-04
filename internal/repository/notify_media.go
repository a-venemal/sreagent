package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/sreagent/sreagent/internal/model"
)

// NotifyMediaRepository handles notify_medias persistence.
type NotifyMediaRepository struct {
	db *gorm.DB
}

// NewNotifyMediaRepository creates a new NotifyMediaRepository.
func NewNotifyMediaRepository(db *gorm.DB) *NotifyMediaRepository {
	return &NotifyMediaRepository{db: db}
}

// Create creates a new notify media.
func (r *NotifyMediaRepository) Create(ctx context.Context, media *model.NotifyMedia) error {
	return r.db.WithContext(ctx).Create(media).Error
}

// GetByID returns a notify media by its ID.
func (r *NotifyMediaRepository) GetByID(ctx context.Context, id uint) (*model.NotifyMedia, error) {
	var media model.NotifyMedia
	err := r.db.WithContext(ctx).First(&media, id).Error
	if err != nil {
		return nil, err
	}
	return &media, nil
}

// List returns a paginated list of notify medias.
func (r *NotifyMediaRepository) List(ctx context.Context, page, pageSize int) ([]model.NotifyMedia, int64, error) {
	var list []model.NotifyMedia
	var total int64

	query := r.db.WithContext(ctx).Model(&model.NotifyMedia{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Update updates an existing notify media.
func (r *NotifyMediaRepository) Update(ctx context.Context, media *model.NotifyMedia) error {
	return r.db.WithContext(ctx).Save(media).Error
}

// Delete soft-deletes a notify media by ID.
func (r *NotifyMediaRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.NotifyMedia{}, id).Error
}

// ListEnabled returns all enabled notify medias.
func (r *NotifyMediaRepository) ListEnabled(ctx context.Context) ([]model.NotifyMedia, error) {
	var list []model.NotifyMedia
	err := r.db.WithContext(ctx).
		Where("is_enabled = ?", true).
		Order("id ASC").
		Find(&list).Error
	return list, err
}

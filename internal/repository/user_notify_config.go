package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/sreagent/sreagent/internal/model"
)

type UserNotifyConfigRepository struct {
	db *gorm.DB
}

func NewUserNotifyConfigRepository(db *gorm.DB) *UserNotifyConfigRepository {
	return &UserNotifyConfigRepository{db: db}
}

// ListByUserID returns all notify configs for a user.
func (r *UserNotifyConfigRepository) ListByUserID(ctx context.Context, userID uint) ([]model.UserNotifyConfig, error) {
	var cfgs []model.UserNotifyConfig
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&cfgs).Error
	return cfgs, err
}

// Upsert inserts or updates a config by (user_id, media_type).
func (r *UserNotifyConfigRepository) Upsert(ctx context.Context, cfg *model.UserNotifyConfig) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "media_type"}},
			DoUpdates: clause.AssignmentColumns([]string{"config", "is_enabled", "updated_at"}),
		}).
		Create(cfg).Error
}

// DeleteByMediaType removes a specific media type config for a user.
func (r *UserNotifyConfigRepository) DeleteByMediaType(ctx context.Context, userID uint, mediaType string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND media_type = ?", userID, mediaType).
		Delete(&model.UserNotifyConfig{}).Error
}

// DeleteAll removes all configs for a user.
func (r *UserNotifyConfigRepository) DeleteAll(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.UserNotifyConfig{}).Error
}

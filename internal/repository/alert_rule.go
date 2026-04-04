package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/sreagent/sreagent/internal/model"
)

type AlertRuleRepository struct {
	db *gorm.DB
}

func NewAlertRuleRepository(db *gorm.DB) *AlertRuleRepository {
	return &AlertRuleRepository{db: db}
}

func (r *AlertRuleRepository) Create(ctx context.Context, rule *model.AlertRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *AlertRuleRepository) GetByID(ctx context.Context, id uint) (*model.AlertRule, error) {
	var rule model.AlertRule
	err := r.db.WithContext(ctx).Preload("DataSource").First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *AlertRuleRepository) List(ctx context.Context, severity, status, groupName string, page, pageSize int) ([]model.AlertRule, int64, error) {
	var list []model.AlertRule
	var total int64

	query := r.db.WithContext(ctx).Model(&model.AlertRule{})
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if groupName != "" {
		query = query.Where("group_name = ?", groupName)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("DataSource").Offset(offset).Limit(pageSize).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *AlertRuleRepository) Update(ctx context.Context, rule *model.AlertRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

func (r *AlertRuleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.AlertRule{}, id).Error
}

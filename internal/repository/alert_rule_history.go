package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/sreagent/sreagent/internal/model"
)

// AlertRuleHistoryRepository provides CRUD operations for alert rule change history.
type AlertRuleHistoryRepository struct {
	db *gorm.DB
}

// NewAlertRuleHistoryRepository creates a new AlertRuleHistoryRepository.
func NewAlertRuleHistoryRepository(db *gorm.DB) *AlertRuleHistoryRepository {
	return &AlertRuleHistoryRepository{db: db}
}

// Create inserts a new alert rule history record.
func (r *AlertRuleHistoryRepository) Create(ctx context.Context, h *model.AlertRuleHistory) error {
	return r.db.WithContext(ctx).Create(h).Error
}

// ListByRuleID returns all history records for a given rule, newest first.
func (r *AlertRuleHistoryRepository) ListByRuleID(ctx context.Context, ruleID uint, page, pageSize int) ([]model.AlertRuleHistory, int64, error) {
	var list []model.AlertRuleHistory
	var total int64

	query := r.db.WithContext(ctx).Model(&model.AlertRuleHistory{}).Where("rule_id = ?", ruleID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// GetByID returns a single history record by ID.
func (r *AlertRuleHistoryRepository) GetByID(ctx context.Context, id uint) (*model.AlertRuleHistory, error) {
	var h model.AlertRuleHistory
	if err := r.db.WithContext(ctx).First(&h, id).Error; err != nil {
		return nil, err
	}
	return &h, nil
}

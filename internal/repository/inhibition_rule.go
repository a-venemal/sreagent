package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/sreagent/sreagent/internal/model"
)

// InhibitionRuleRepository handles persistence for InhibitionRule records.
type InhibitionRuleRepository struct {
	db *gorm.DB
}

// NewInhibitionRuleRepository creates a new InhibitionRuleRepository.
func NewInhibitionRuleRepository(db *gorm.DB) *InhibitionRuleRepository {
	return &InhibitionRuleRepository{db: db}
}

// Create inserts a new inhibition rule.
func (r *InhibitionRuleRepository) Create(ctx context.Context, rule *model.InhibitionRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

// GetByID returns an inhibition rule by primary key.
func (r *InhibitionRuleRepository) GetByID(ctx context.Context, id uint) (*model.InhibitionRule, error) {
	var rule model.InhibitionRule
	err := r.db.WithContext(ctx).First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// List returns a paginated list of inhibition rules.
func (r *InhibitionRuleRepository) List(ctx context.Context, page, pageSize int) ([]model.InhibitionRule, int64, error) {
	var list []model.InhibitionRule
	var total int64

	q := r.db.WithContext(ctx).Model(&model.InhibitionRule{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// Update saves changes to an existing inhibition rule.
func (r *InhibitionRuleRepository) Update(ctx context.Context, rule *model.InhibitionRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

// Delete soft-deletes an inhibition rule.
func (r *InhibitionRuleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.InhibitionRule{}, id).Error
}

// FindAllEnabled returns all currently enabled inhibition rules.
func (r *InhibitionRuleRepository) FindAllEnabled(ctx context.Context) ([]model.InhibitionRule, error) {
	var rules []model.InhibitionRule
	err := r.db.WithContext(ctx).Where("is_enabled = ?", true).Find(&rules).Error
	return rules, err
}

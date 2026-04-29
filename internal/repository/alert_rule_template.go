package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/sreagent/sreagent/internal/model"
)

type AlertRuleTemplateRepository struct {
	db *gorm.DB
}

func NewAlertRuleTemplateRepository(db *gorm.DB) *AlertRuleTemplateRepository {
	return &AlertRuleTemplateRepository{db: db}
}

func (r *AlertRuleTemplateRepository) Create(ctx context.Context, tpl *model.AlertRuleTemplate) error {
	return r.db.WithContext(ctx).Create(tpl).Error
}

func (r *AlertRuleTemplateRepository) GetByID(ctx context.Context, id uint) (*model.AlertRuleTemplate, error) {
	var tpl model.AlertRuleTemplate
	err := r.db.WithContext(ctx).First(&tpl, id).Error
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

func (r *AlertRuleTemplateRepository) Update(ctx context.Context, tpl *model.AlertRuleTemplate) error {
	return r.db.WithContext(ctx).Save(tpl).Error
}

func (r *AlertRuleTemplateRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.AlertRuleTemplate{}, id).Error
}

func (r *AlertRuleTemplateRepository) List(ctx context.Context, category, search string, page, pageSize int) ([]model.AlertRuleTemplate, int64, error) {
	var list []model.AlertRuleTemplate
	var total int64

	query := r.db.WithContext(ctx).Model(&model.AlertRuleTemplate{})
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *AlertRuleTemplateRepository) ListCategories(ctx context.Context) ([]string, error) {
	var categories []string
	err := r.db.WithContext(ctx).Model(&model.AlertRuleTemplate{}).
		Where("category != ''").
		Distinct("category").
		Pluck("category", &categories).Error
	return categories, err
}

func (r *AlertRuleTemplateRepository) IncrementUsage(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.AlertRuleTemplate{}).
		Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}

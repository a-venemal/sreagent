package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/sreagent/sreagent/internal/model"
)

type DashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) Create(ctx context.Context, d *model.Dashboard) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *DashboardRepository) GetByID(ctx context.Context, id uint) (*model.Dashboard, error) {
	var d model.Dashboard
	err := r.db.WithContext(ctx).First(&d, id).Error
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DashboardRepository) List(ctx context.Context, search string, page, pageSize int) ([]model.Dashboard, int64, error) {
	var list []model.Dashboard
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Dashboard{})
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

func (r *DashboardRepository) Update(ctx context.Context, d *model.Dashboard) error {
	return r.db.WithContext(ctx).Save(d).Error
}

func (r *DashboardRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Dashboard{}, id).Error
}

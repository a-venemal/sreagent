package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/sreagent/sreagent/internal/model"
)

type DataSourceRepository struct {
	db *gorm.DB
}

func NewDataSourceRepository(db *gorm.DB) *DataSourceRepository {
	return &DataSourceRepository{db: db}
}

func (r *DataSourceRepository) Create(ctx context.Context, ds *model.DataSource) error {
	return r.db.WithContext(ctx).Create(ds).Error
}

func (r *DataSourceRepository) GetByID(ctx context.Context, id uint) (*model.DataSource, error) {
	var ds model.DataSource
	err := r.db.WithContext(ctx).First(&ds, id).Error
	if err != nil {
		return nil, err
	}
	return &ds, nil
}

func (r *DataSourceRepository) GetByName(ctx context.Context, name string) (*model.DataSource, error) {
	var ds model.DataSource
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&ds).Error
	if err != nil {
		return nil, err
	}
	return &ds, nil
}

func (r *DataSourceRepository) List(ctx context.Context, dsType string, page, pageSize int) ([]model.DataSource, int64, error) {
	var list []model.DataSource
	var total int64

	query := r.db.WithContext(ctx).Model(&model.DataSource{})
	if dsType != "" {
		query = query.Where("type = ?", dsType)
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

func (r *DataSourceRepository) Update(ctx context.Context, ds *model.DataSource) error {
	return r.db.WithContext(ctx).Save(ds).Error
}

func (r *DataSourceRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.DataSource{}, id).Error
}

// ListEnabled returns all enabled datasources.
func (r *DataSourceRepository) ListEnabled(ctx context.Context) ([]model.DataSource, error) {
	var list []model.DataSource
	err := r.db.WithContext(ctx).Where("is_enabled = ?", true).Find(&list).Error
	return list, err
}

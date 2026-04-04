package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/sreagent/sreagent/internal/model"
)

// MessageTemplateRepository handles message_templates persistence.
type MessageTemplateRepository struct {
	db *gorm.DB
}

// NewMessageTemplateRepository creates a new MessageTemplateRepository.
func NewMessageTemplateRepository(db *gorm.DB) *MessageTemplateRepository {
	return &MessageTemplateRepository{db: db}
}

// Create creates a new message template.
func (r *MessageTemplateRepository) Create(ctx context.Context, tmpl *model.MessageTemplate) error {
	return r.db.WithContext(ctx).Create(tmpl).Error
}

// GetByID returns a message template by its ID.
func (r *MessageTemplateRepository) GetByID(ctx context.Context, id uint) (*model.MessageTemplate, error) {
	var tmpl model.MessageTemplate
	err := r.db.WithContext(ctx).First(&tmpl, id).Error
	if err != nil {
		return nil, err
	}
	return &tmpl, nil
}

// GetByName returns a message template by its unique name.
func (r *MessageTemplateRepository) GetByName(ctx context.Context, name string) (*model.MessageTemplate, error) {
	var tmpl model.MessageTemplate
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&tmpl).Error
	if err != nil {
		return nil, err
	}
	return &tmpl, nil
}

// List returns a paginated list of message templates.
func (r *MessageTemplateRepository) List(ctx context.Context, page, pageSize int) ([]model.MessageTemplate, int64, error) {
	var list []model.MessageTemplate
	var total int64

	query := r.db.WithContext(ctx).Model(&model.MessageTemplate{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Update updates an existing message template.
func (r *MessageTemplateRepository) Update(ctx context.Context, tmpl *model.MessageTemplate) error {
	return r.db.WithContext(ctx).Save(tmpl).Error
}

// Delete soft-deletes a message template by ID.
func (r *MessageTemplateRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.MessageTemplate{}, id).Error
}

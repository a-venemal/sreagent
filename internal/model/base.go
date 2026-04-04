package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel is the common base for all models.
type BaseModel struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Labels represents a map of key-value label pairs.
// Stored as JSON in the database.
type Labels map[string]string

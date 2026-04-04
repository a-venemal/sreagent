package model

import "time"

// SystemSetting stores platform-level settings as key-value pairs.
// Settings are grouped by a "group" field (e.g. "ai", "lark") so the UI
// can fetch/save an entire group at once.
// Sensitive values (api_key, app_secret, etc.) are encrypted with AES-256-GCM
// via SystemSettingService before storage (see service/system_setting.go).
type SystemSetting struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Group     string    `gorm:"type:varchar(64);not null;uniqueIndex:idx_group_key" json:"group"`
	Key       string    `gorm:"type:varchar(128);not null;uniqueIndex:idx_group_key" json:"key"`
	Value     string    `gorm:"type:text;not null" json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (SystemSetting) TableName() string { return "system_settings" }

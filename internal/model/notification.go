package model

// NotifyChannelType defines the type of notification channel.
type NotifyChannelType string

const (
	ChannelTypeLarkWebhook NotifyChannelType = "lark_webhook"
	ChannelTypeLarkBot     NotifyChannelType = "lark_bot"
	ChannelTypeEmail       NotifyChannelType = "email"
	ChannelTypeSMS         NotifyChannelType = "sms"
	ChannelTypeCustom      NotifyChannelType = "custom_webhook"
)

// NotifyChannel represents a notification channel (e.g., a Lark group webhook).
type NotifyChannel struct {
	BaseModel
	Name        string            `json:"name" gorm:"size:128;not null"`
	Type        NotifyChannelType `json:"type" gorm:"size:32;not null;index"`
	Description string            `json:"description" gorm:"size:512"`
	Labels      JSONLabels        `json:"labels" gorm:"type:json"` // for matching routing rules
	// Channel-specific config (stored as JSON)
	// Lark webhook: {"webhook_url": "https://..."}
	// Email: {"smtp_host": "...", "recipients": ["a@b.com"]}
	Config    string `json:"-" gorm:"type:text;not null"`
	IsEnabled bool   `json:"is_enabled" gorm:"default:true"`
}

func (NotifyChannel) TableName() string {
	return "notify_channels"
}

// NotifyPolicy defines routing rules: which alerts go to which channels.
type NotifyPolicy struct {
	BaseModel
	Name        string `json:"name" gorm:"size:128;not null"`
	Description string `json:"description" gorm:"size:512"`
	// Label matchers for this policy (must match ALL labels)
	MatchLabels JSONLabels `json:"match_labels" gorm:"type:json;not null"`
	// Severity filter (empty = all severities)
	Severities string `json:"severities" gorm:"size:128"` // comma-separated: "critical,warning"
	// Target channel
	ChannelID uint          `json:"channel_id" gorm:"index;not null"`
	Channel   NotifyChannel `json:"channel,omitempty" gorm:"foreignKey:ChannelID"`
	// Throttle settings
	ThrottleMinutes int `json:"throttle_minutes" gorm:"default:5"` // min interval between notifications
	// Template
	TemplateName string `json:"template_name" gorm:"size:64;default:default"`
	IsEnabled    bool   `json:"is_enabled" gorm:"default:true"`
	// Priority (higher = evaluated first)
	Priority int `json:"priority" gorm:"default:0"`
}

func (NotifyPolicy) TableName() string {
	return "notify_policies"
}

// NotifyRecord tracks sent notifications for audit and throttling.
type NotifyRecord struct {
	BaseModel
	EventID   uint   `json:"event_id" gorm:"index;not null"`
	ChannelID uint   `json:"channel_id" gorm:"index;not null"`
	PolicyID  uint   `json:"policy_id" gorm:"index"`
	Status    string `json:"status" gorm:"size:32;not null"` // sent, failed, throttled
	Response  string `json:"response" gorm:"type:text"`      // API response for debugging
}

func (NotifyRecord) TableName() string {
	return "notify_records"
}

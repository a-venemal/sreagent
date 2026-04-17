package model

// AlertSeverity defines the severity level of an alert.
type AlertSeverity string

const (
	SeverityCritical AlertSeverity = "critical"
	SeverityWarning  AlertSeverity = "warning"
	SeverityInfo     AlertSeverity = "info"
)

// AlertRuleStatus defines the status of an alert rule.
type AlertRuleStatus string

const (
	RuleStatusEnabled  AlertRuleStatus = "enabled"
	RuleStatusDisabled AlertRuleStatus = "disabled"
	RuleStatusMuted    AlertRuleStatus = "muted"
)

// AlertRule represents an alerting rule definition.
type AlertRule struct {
	BaseModel
	Name         string     `json:"name" gorm:"size:256;not null;index"`
	DisplayName  string     `json:"display_name" gorm:"size:256"`
	Description  string     `json:"description" gorm:"type:text"`
	DataSourceID uint       `json:"datasource_id" gorm:"index;not null"`
	DataSource   DataSource `json:"datasource,omitempty" gorm:"foreignKey:DataSourceID"`
	// Rule expression (PromQL, LogsQL, Zabbix trigger expression, etc.)
	Expression string `json:"expression" gorm:"type:text;not null"`
	// For duration (e.g., "5m" - alert must be firing for this duration)
	ForDuration string          `json:"for_duration" gorm:"size:32;default:0s"`
	Severity    AlertSeverity   `json:"severity" gorm:"size:32;not null;index"`
	Labels      JSONLabels      `json:"labels" gorm:"type:json"`
	Annotations JSONLabels      `json:"annotations" gorm:"type:json"` // summary, description templates
	Status      AlertRuleStatus `json:"status" gorm:"size:32;default:enabled;index"`
	// Grouping
	GroupName string `json:"group_name" gorm:"size:128;index"`
	Category  string `json:"category" gorm:"size:64;index;default:''"`
	// Version tracking
	Version   int  `json:"version" gorm:"default:1"`
	CreatedBy uint `json:"created_by" gorm:"index"`
	UpdatedBy uint `json:"updated_by"`
	// Evaluation interval in seconds (default 60)
	EvalInterval int `json:"eval_interval" gorm:"default:60"`
	// Recovery hold duration (留观时长) - e.g., "5m"
	RecoveryHold string `json:"recovery_hold" gorm:"size:32;default:0s"`
	// NoData detection
	NoDataEnabled  bool   `json:"nodata_enabled" gorm:"default:false"`
	NoDataDuration string `json:"nodata_duration" gorm:"size:32;default:5m"` // after this duration of no data, fire nodata alert
	// Level suppression (for rules with multiple severity conditions)
	SuppressEnabled bool `json:"suppress_enabled" gorm:"default:false"`
	// Business group
	BizGroupID *uint `json:"biz_group_id" gorm:"index"`
}

func (AlertRule) TableName() string {
	return "alert_rules"
}

// AlertRuleHistory records changes to alert rules for audit trail.
type AlertRuleHistory struct {
	BaseModel
	RuleID     uint   `json:"rule_id" gorm:"index;not null"`
	Version    int    `json:"version" gorm:"not null"`
	ChangeType string `json:"change_type" gorm:"size:32;not null"` // created, updated, deleted
	Snapshot   string `json:"snapshot" gorm:"type:text;not null"`  // JSON snapshot of the rule
	ChangedBy  uint   `json:"changed_by" gorm:"index"`
}

func (AlertRuleHistory) TableName() string {
	return "alert_rule_histories"
}

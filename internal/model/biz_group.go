package model

// BizGroup represents a business group for organizing resources and teams.
// Group names support "/" for tree hierarchy (e.g., "DBA/MySQL", "DBA/Redis").
type BizGroup struct {
	BaseModel
	Name        string     `json:"name" gorm:"size:128;not null"` // supports "/" for tree: "DBA/MySQL"
	Description string     `json:"description" gorm:"size:512"`
	ParentID    *uint      `json:"parent_id" gorm:"index"`
	Labels      JSONLabels `json:"labels" gorm:"type:json"`
	// MatchLabels defines which alert labels this group "owns".
	// Example: {"biz_project": "mdc"} means all alerts with biz_project=mdc belong to this group.
	// Supports operator prefixes: "!=", "=~", "!~" (same as MuteRule/NotifyRule).
	MatchLabels JSONLabels `json:"match_labels" gorm:"type:json"`
	// Members
	Members []User `json:"members,omitempty" gorm:"many2many:biz_group_members;"`
}

func (BizGroup) TableName() string {
	return "biz_groups"
}

// BizGroupMember is the join table for business group-user relationship with role info.
type BizGroupMember struct {
	BizGroupID uint   `json:"biz_group_id" gorm:"primaryKey"`
	UserID     uint   `json:"user_id" gorm:"primaryKey"`
	Role       string `json:"role" gorm:"size:32;default:member"` // admin, member
}

func (BizGroupMember) TableName() string {
	return "biz_group_members"
}

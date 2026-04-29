package model

// Dashboard represents a monitoring dashboard with panels and variables.
type Dashboard struct {
	BaseModel
	Name        string     `json:"name" gorm:"size:256;not null"`
	Description string     `json:"description" gorm:"size:1024"`
	Tags        JSONLabels `json:"tags" gorm:"type:json"`
	// Config stores the full dashboard configuration as JSON string:
	// { panels: [...], layout: {...}, variables: [...] }
	Config    string `json:"config" gorm:"type:longtext"`
	CreatedBy uint   `json:"created_by" gorm:"index"`
	UpdatedBy uint   `json:"updated_by"`
	IsPublic  bool   `json:"is_public" gorm:"default:false"`
}

func (Dashboard) TableName() string {
	return "dashboards"
}

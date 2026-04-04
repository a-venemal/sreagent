package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sreagent/sreagent/internal/model"
	"github.com/sreagent/sreagent/internal/service"
)

// MuteRuleHandler handles mute rule API requests.
type MuteRuleHandler struct {
	svc *service.MuteRuleService
}

// NewMuteRuleHandler creates a new MuteRuleHandler.
func NewMuteRuleHandler(svc *service.MuteRuleService) *MuteRuleHandler {
	return &MuteRuleHandler{svc: svc}
}

// CreateMuteRuleRequest is the request body for creating a mute rule.
type CreateMuteRuleRequest struct {
	Name          string           `json:"name" binding:"required"`
	Description   string           `json:"description"`
	MatchLabels   model.JSONLabels `json:"match_labels"`
	Severities    string           `json:"severities"`
	StartTime     *time.Time       `json:"start_time"`
	EndTime       *time.Time       `json:"end_time"`
	PeriodicStart string           `json:"periodic_start"`
	PeriodicEnd   string           `json:"periodic_end"`
	DaysOfWeek    string           `json:"days_of_week"`
	Timezone      string           `json:"timezone"`
	IsEnabled     bool             `json:"is_enabled"`
	RuleIDs       string           `json:"rule_ids"`
}

// Create creates a new mute rule.
func (h *MuteRuleHandler) Create(c *gin.Context) {
	var req CreateMuteRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	tz := req.Timezone
	if tz == "" {
		tz = "Asia/Shanghai"
	}

	rule := &model.MuteRule{
		Name:          req.Name,
		Description:   req.Description,
		MatchLabels:   req.MatchLabels,
		Severities:    req.Severities,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		PeriodicStart: req.PeriodicStart,
		PeriodicEnd:   req.PeriodicEnd,
		DaysOfWeek:    req.DaysOfWeek,
		Timezone:      tz,
		CreatedBy:     GetCurrentUserID(c),
		IsEnabled:     req.IsEnabled,
		RuleIDs:       req.RuleIDs,
	}

	if err := h.svc.Create(c.Request.Context(), rule); err != nil {
		Error(c, err)
		return
	}

	Success(c, rule)
}

// Get returns a mute rule by ID.
func (h *MuteRuleHandler) Get(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	rule, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, rule)
}

// List returns a paginated list of mute rules.
func (h *MuteRuleHandler) List(c *gin.Context) {
	pq := GetPageQuery(c)

	list, total, err := h.svc.List(c.Request.Context(), pq.Page, pq.PageSize)
	if err != nil {
		Error(c, err)
		return
	}

	SuccessPage(c, list, total, pq.Page, pq.PageSize)
}

// Update updates an existing mute rule.
func (h *MuteRuleHandler) Update(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	var req CreateMuteRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	tz := req.Timezone
	if tz == "" {
		tz = "Asia/Shanghai"
	}

	rule := &model.MuteRule{
		Name:          req.Name,
		Description:   req.Description,
		MatchLabels:   req.MatchLabels,
		Severities:    req.Severities,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		PeriodicStart: req.PeriodicStart,
		PeriodicEnd:   req.PeriodicEnd,
		DaysOfWeek:    req.DaysOfWeek,
		Timezone:      tz,
		IsEnabled:     req.IsEnabled,
		RuleIDs:       req.RuleIDs,
	}
	rule.ID = id

	if err := h.svc.Update(c.Request.Context(), rule); err != nil {
		Error(c, err)
		return
	}

	Success(c, rule)
}

// Delete deletes a mute rule by ID.
func (h *MuteRuleHandler) Delete(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

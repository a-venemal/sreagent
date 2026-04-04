package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sreagent/sreagent/internal/model"
	"github.com/sreagent/sreagent/internal/service"
)

type ScheduleHandler struct {
	svc *service.ScheduleService
}

func NewScheduleHandler(svc *service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{svc: svc}
}

// ---------------------------------------------------------------------------
// Request types
// ---------------------------------------------------------------------------

// CreateScheduleRequest is the request body for creating a schedule.
type CreateScheduleRequest struct {
	Name         string             `json:"name" binding:"required"`
	TeamID       *uint              `json:"team_id"`
	Description  string             `json:"description"`
	RotationType model.RotationType `json:"rotation_type" binding:"required"`
	Timezone     string             `json:"timezone"`
	HandoffTime  string             `json:"handoff_time"`
	HandoffDay   int                `json:"handoff_day"`
	IsEnabled    *bool              `json:"is_enabled"`
}

// UpdateScheduleRequest is the request body for updating a schedule.
type UpdateScheduleRequest struct {
	Name         string             `json:"name" binding:"required"`
	Description  string             `json:"description"`
	RotationType model.RotationType `json:"rotation_type" binding:"required"`
	Timezone     string             `json:"timezone"`
	HandoffTime  string             `json:"handoff_time"`
	HandoffDay   int                `json:"handoff_day"`
	IsEnabled    *bool              `json:"is_enabled"`
}

// SetParticipantsRequest is the request body for setting schedule participants.
type SetParticipantsRequest struct {
	UserIDs []uint `json:"user_ids" binding:"required"`
}

// CreateOverrideRequest is the request body for creating a schedule override.
type CreateOverrideRequest struct {
	UserID    uint      `json:"user_id" binding:"required"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
	Reason    string    `json:"reason"`
}

// CreateShiftRequest is the request body for creating an on-call shift.
type CreateShiftRequest struct {
	UserID         uint      `json:"user_id" binding:"required"`
	StartTime      time.Time `json:"start_time" binding:"required"`
	EndTime        time.Time `json:"end_time" binding:"required"`
	SeverityFilter string    `json:"severity_filter"`
	Note           string    `json:"note"`
}

// UpdateShiftRequest is the request body for updating an on-call shift.
type UpdateShiftRequest struct {
	UserID         uint      `json:"user_id" binding:"required"`
	StartTime      time.Time `json:"start_time" binding:"required"`
	EndTime        time.Time `json:"end_time" binding:"required"`
	SeverityFilter string    `json:"severity_filter"`
	Note           string    `json:"note"`
}

// GenerateShiftsRequest is the request body for auto-generating rotation shifts.
type GenerateShiftsRequest struct {
	Weeks int `json:"weeks" binding:"required,min=1,max=52"`
}

// CreateEscalationPolicyRequest is the request body for creating an escalation policy.
type CreateEscalationPolicyRequest struct {
	Name      string `json:"name" binding:"required"`
	TeamID    uint   `json:"team_id" binding:"required"`
	IsEnabled *bool  `json:"is_enabled"`
}

// UpdateEscalationPolicyRequest is the request body for updating an escalation policy.
type UpdateEscalationPolicyRequest struct {
	Name      string `json:"name" binding:"required"`
	TeamID    uint   `json:"team_id" binding:"required"`
	IsEnabled *bool  `json:"is_enabled"`
}

// CreateEscalationStepRequest is the request body for creating an escalation step.
type CreateEscalationStepRequest struct {
	StepOrder       int    `json:"step_order" binding:"required"`
	DelayMinutes    int    `json:"delay_minutes" binding:"required"`
	TargetType      string `json:"target_type" binding:"required"`
	TargetID        uint   `json:"target_id" binding:"required"`
	NotifyChannelID *uint  `json:"notify_channel_id"`
}

// UpdateEscalationStepRequest is the request body for updating an escalation step.
type UpdateEscalationStepRequest struct {
	StepOrder       int    `json:"step_order" binding:"required"`
	DelayMinutes    int    `json:"delay_minutes" binding:"required"`
	TargetType      string `json:"target_type" binding:"required"`
	TargetID        uint   `json:"target_id" binding:"required"`
	NotifyChannelID *uint  `json:"notify_channel_id"`
}

// ---------------------------------------------------------------------------
// Schedule CRUD Handlers
// ---------------------------------------------------------------------------

// CreateSchedule creates a new on-call schedule.
func (h *ScheduleHandler) CreateSchedule(c *gin.Context) {
	var req CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}

	timezone := req.Timezone
	if timezone == "" {
		timezone = "Asia/Shanghai"
	}

	handoffTime := req.HandoffTime
	if handoffTime == "" {
		handoffTime = "09:00"
	}

	schedule := &model.Schedule{
		Name:         req.Name,
		TeamID:       req.TeamID,
		Description:  req.Description,
		RotationType: req.RotationType,
		Timezone:     timezone,
		HandoffTime:  handoffTime,
		HandoffDay:   req.HandoffDay,
		IsEnabled:    isEnabled,
	}

	if err := h.svc.CreateSchedule(c.Request.Context(), schedule); err != nil {
		Error(c, err)
		return
	}

	Success(c, schedule)
}

// GetSchedule returns a schedule by ID.
func (h *ScheduleHandler) GetSchedule(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	schedule, err := h.svc.GetScheduleByID(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, schedule)
}

// ListSchedules returns a paginated list of schedules.
func (h *ScheduleHandler) ListSchedules(c *gin.Context) {
	pq := GetPageQuery(c)

	var teamID uint
	if tidStr := c.Query("team_id"); tidStr != "" {
		if tid, err := strconv.ParseUint(tidStr, 10, 64); err == nil {
			teamID = uint(tid)
		}
	}

	list, total, err := h.svc.ListSchedules(c.Request.Context(), teamID, pq.Page, pq.PageSize)
	if err != nil {
		Error(c, err)
		return
	}

	SuccessPage(c, list, total, pq.Page, pq.PageSize)
}

// UpdateSchedule updates a schedule.
func (h *ScheduleHandler) UpdateSchedule(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	var req UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}

	schedule := &model.Schedule{
		Name:         req.Name,
		Description:  req.Description,
		RotationType: req.RotationType,
		Timezone:     req.Timezone,
		HandoffTime:  req.HandoffTime,
		HandoffDay:   req.HandoffDay,
		IsEnabled:    isEnabled,
	}
	schedule.ID = id

	if err := h.svc.UpdateSchedule(c.Request.Context(), schedule); err != nil {
		Error(c, err)
		return
	}

	Success(c, schedule)
}

// DeleteSchedule deletes a schedule.
func (h *ScheduleHandler) DeleteSchedule(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.svc.DeleteSchedule(c.Request.Context(), id); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// ---------------------------------------------------------------------------
// On-Call
// ---------------------------------------------------------------------------

// GetCurrentOnCall returns the user currently on-call for the given schedule.
func (h *ScheduleHandler) GetCurrentOnCall(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	result, err := h.svc.GetCurrentOnCall(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, result)
}

// ---------------------------------------------------------------------------
// Participants
// ---------------------------------------------------------------------------

// SetParticipants sets the participant list for a schedule.
func (h *ScheduleHandler) SetParticipants(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	var req SetParticipantsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	if err := h.svc.SetParticipants(c.Request.Context(), id, req.UserIDs); err != nil {
		Error(c, err)
		return
	}

	// Return the updated participant list
	participants, err := h.svc.ListParticipants(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, participants)
}

// ---------------------------------------------------------------------------
// Overrides
// ---------------------------------------------------------------------------

// CreateOverride creates a schedule override.
func (h *ScheduleHandler) CreateOverride(c *gin.Context) {
	scheduleID, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	var req CreateOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	override := &model.ScheduleOverride{
		ScheduleID: scheduleID,
		UserID:     req.UserID,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Reason:     req.Reason,
	}

	if err := h.svc.CreateOverride(c.Request.Context(), override); err != nil {
		Error(c, err)
		return
	}

	Success(c, override)
}

// DeleteOverride deletes a schedule override.
func (h *ScheduleHandler) DeleteOverride(c *gin.Context) {
	oid, err := GetIDParam(c, "oid")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.svc.DeleteOverride(c.Request.Context(), oid); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// ---------------------------------------------------------------------------
// Escalation Policy CRUD Handlers
// ---------------------------------------------------------------------------

// CreateEscalationPolicy creates a new escalation policy.
func (h *ScheduleHandler) CreateEscalationPolicy(c *gin.Context) {
	var req CreateEscalationPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}

	policy := &model.EscalationPolicy{
		Name:      req.Name,
		TeamID:    req.TeamID,
		IsEnabled: isEnabled,
	}

	if err := h.svc.CreateEscalationPolicy(c.Request.Context(), policy); err != nil {
		Error(c, err)
		return
	}

	Success(c, policy)
}

// GetEscalationPolicy returns an escalation policy by ID.
func (h *ScheduleHandler) GetEscalationPolicy(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	policy, err := h.svc.GetEscalationPolicyByID(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}

	// Also fetch the steps
	steps, _ := h.svc.ListEscalationSteps(c.Request.Context(), id)

	Success(c, gin.H{
		"policy": policy,
		"steps":  steps,
	})
}

// ListEscalationPolicies returns escalation policies, optionally filtered by team.
func (h *ScheduleHandler) ListEscalationPolicies(c *gin.Context) {
	var teamID uint
	if tidStr := c.Query("team_id"); tidStr != "" {
		if tid, err := strconv.ParseUint(tidStr, 10, 64); err == nil {
			teamID = uint(tid)
		}
	}

	list, err := h.svc.ListEscalationPolicies(c.Request.Context(), teamID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, list)
}

// UpdateEscalationPolicy updates an escalation policy.
func (h *ScheduleHandler) UpdateEscalationPolicy(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	var req UpdateEscalationPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}

	policy := &model.EscalationPolicy{
		Name:      req.Name,
		TeamID:    req.TeamID,
		IsEnabled: isEnabled,
	}
	policy.ID = id

	if err := h.svc.UpdateEscalationPolicy(c.Request.Context(), policy); err != nil {
		Error(c, err)
		return
	}

	Success(c, policy)
}

// DeleteEscalationPolicy deletes an escalation policy.
func (h *ScheduleHandler) DeleteEscalationPolicy(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.svc.DeleteEscalationPolicy(c.Request.Context(), id); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// ---------------------------------------------------------------------------
// Escalation Step Handlers
// ---------------------------------------------------------------------------

// CreateEscalationStep creates a new step in an escalation policy.
func (h *ScheduleHandler) CreateEscalationStep(c *gin.Context) {
	policyID, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	var req CreateEscalationStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	step := &model.EscalationStep{
		PolicyID:        policyID,
		StepOrder:       req.StepOrder,
		DelayMinutes:    req.DelayMinutes,
		TargetType:      req.TargetType,
		TargetID:        req.TargetID,
		NotifyChannelID: req.NotifyChannelID,
	}

	if err := h.svc.CreateEscalationStep(c.Request.Context(), step); err != nil {
		Error(c, err)
		return
	}

	Success(c, step)
}

// UpdateEscalationStep updates a step in an escalation policy.
func (h *ScheduleHandler) UpdateEscalationStep(c *gin.Context) {
	policyID, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	stepID, err := GetIDParam(c, "stepId")
	if err != nil {
		Error(c, err)
		return
	}

	var req UpdateEscalationStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	step := &model.EscalationStep{
		PolicyID:        policyID,
		StepOrder:       req.StepOrder,
		DelayMinutes:    req.DelayMinutes,
		TargetType:      req.TargetType,
		TargetID:        req.TargetID,
		NotifyChannelID: req.NotifyChannelID,
	}
	step.ID = stepID

	if err := h.svc.UpdateEscalationStep(c.Request.Context(), step); err != nil {
		Error(c, err)
		return
	}

	Success(c, step)
}

// DeleteEscalationStep deletes a step from an escalation policy.
func (h *ScheduleHandler) DeleteEscalationStep(c *gin.Context) {
	stepID, err := GetIDParam(c, "stepId")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.svc.DeleteEscalationStep(c.Request.Context(), stepID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// ---------------------------------------------------------------------------
// OnCallShift Handlers
// ---------------------------------------------------------------------------

// ListShifts returns shifts for a schedule in the given time window.
// GET /schedules/:id/shifts?start=<RFC3339>&end=<RFC3339>
func (h *ScheduleHandler) ListShifts(c *gin.Context) {
	scheduleID, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	startStr := c.Query("start")
	endStr := c.Query("end")

	var start, end time.Time
	if startStr != "" {
		if start, err = time.Parse(time.RFC3339, startStr); err != nil {
			ErrorWithMessage(c, 10001, "invalid start time, use RFC3339 format")
			return
		}
	} else {
		start = time.Now().AddDate(0, 0, -7)
	}
	if endStr != "" {
		if end, err = time.Parse(time.RFC3339, endStr); err != nil {
			ErrorWithMessage(c, 10001, "invalid end time, use RFC3339 format")
			return
		}
	} else {
		end = time.Now().AddDate(0, 0, 30)
	}

	shifts, err := h.svc.ListShifts(c.Request.Context(), scheduleID, start, end)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, shifts)
}

// CreateShift creates a new on-call shift for a schedule.
// POST /schedules/:id/shifts
func (h *ScheduleHandler) CreateShift(c *gin.Context) {
	scheduleID, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	var req CreateShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	shift := &model.OnCallShift{
		ScheduleID:     scheduleID,
		UserID:         req.UserID,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		SeverityFilter: req.SeverityFilter,
		Source:         "manual",
		Note:           req.Note,
	}

	if err := h.svc.CreateShift(c.Request.Context(), shift); err != nil {
		Error(c, err)
		return
	}

	Success(c, shift)
}

// UpdateShift updates an existing on-call shift.
// PUT /schedules/:id/shifts/:shiftId
func (h *ScheduleHandler) UpdateShift(c *gin.Context) {
	_, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	shiftID, err := GetIDParam(c, "shiftId")
	if err != nil {
		Error(c, err)
		return
	}

	var req UpdateShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	shift := &model.OnCallShift{
		UserID:         req.UserID,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		SeverityFilter: req.SeverityFilter,
		Note:           req.Note,
	}
	shift.ID = shiftID

	if err := h.svc.UpdateShift(c.Request.Context(), shift); err != nil {
		Error(c, err)
		return
	}

	Success(c, shift)
}

// DeleteShift deletes an on-call shift.
// DELETE /schedules/:id/shifts/:shiftId
func (h *ScheduleHandler) DeleteShift(c *gin.Context) {
	_, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	shiftID, err := GetIDParam(c, "shiftId")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.svc.DeleteShift(c.Request.Context(), shiftID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GenerateShifts auto-generates rotation shifts for a schedule.
// POST /schedules/:id/generate-shifts
func (h *ScheduleHandler) GenerateShifts(c *gin.Context) {
	scheduleID, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	var req GenerateShiftsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	if err := h.svc.GenerateRotationShifts(c.Request.Context(), scheduleID, req.Weeks); err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"message": "shifts generated", "weeks": req.Weeks})
}

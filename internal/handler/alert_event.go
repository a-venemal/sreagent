package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/sreagent/sreagent/internal/model"
	"github.com/sreagent/sreagent/internal/repository"
	"github.com/sreagent/sreagent/internal/service"
)

type AlertEventHandler struct {
	svc *service.AlertEventService
}

func NewAlertEventHandler(svc *service.AlertEventService) *AlertEventHandler {
	return &AlertEventHandler{svc: svc}
}

// List returns paginated alert events with optional filters.
// Supports view_mode=mine|unassigned|all and user_id for role-based visibility.
func (h *AlertEventHandler) List(c *gin.Context) {
	pq := GetPageQuery(c)

	filter := repository.AlertEventFilter{
		Status:   c.Query("status"),
		Severity: c.Query("severity"),
		ViewMode: c.Query("view_mode"),
		Page:     pq.Page,
		PageSize: pq.PageSize,
	}

	// user_id param overrides current user (admin use); default to current user
	if uidStr := c.Query("user_id"); uidStr != "" {
		if uid, err := strconv.ParseUint(uidStr, 10, 64); err == nil {
			filter.UserID = uint(uid)
		}
	}
	if filter.UserID == 0 {
		filter.UserID = GetCurrentUserID(c)
	}

	list, total, err := h.svc.ListWithFilter(c.Request.Context(), filter)
	if err != nil {
		Error(c, err)
		return
	}

	SuccessPage(c, list, total, pq.Page, pq.PageSize)
}

// Get returns a single alert event with its timeline.
func (h *AlertEventHandler) Get(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	event, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, event)
}

// Acknowledge marks an alert as acknowledged by the current user.
func (h *AlertEventHandler) Acknowledge(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	userID := GetCurrentUserID(c)
	if err := h.svc.Acknowledge(c.Request.Context(), id, userID); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// Assign assigns an alert event to a specific user.
func (h *AlertEventHandler) Assign(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	var req struct {
		AssignTo uint   `json:"assign_to" binding:"required"`
		Note     string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	operatorID := GetCurrentUserID(c)
	if err := h.svc.Assign(c.Request.Context(), id, req.AssignTo, operatorID, req.Note); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// Resolve marks an alert as resolved.
func (h *AlertEventHandler) Resolve(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	var req struct {
		Resolution string `json:"resolution"`
	}
	_ = c.ShouldBindJSON(&req)

	userID := GetCurrentUserID(c)
	if err := h.svc.Resolve(c.Request.Context(), id, userID, req.Resolution); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// Close closes an alert event.
func (h *AlertEventHandler) Close(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	var req struct {
		Note string `json:"note"`
	}
	_ = c.ShouldBindJSON(&req)

	userID := GetCurrentUserID(c)
	if err := h.svc.Close(c.Request.Context(), id, userID, req.Note); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// AddComment adds a comment to an alert event timeline.
func (h *AlertEventHandler) AddComment(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	var req struct {
		Note string `json:"note" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	userID := GetCurrentUserID(c)
	if err := h.svc.AddComment(c.Request.Context(), id, userID, req.Note); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetTimeline returns the timeline for an alert event.
func (h *AlertEventHandler) GetTimeline(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	timeline, err := h.svc.GetTimeline(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, timeline)
}

// Silence silences an alert for a specified duration.
func (h *AlertEventHandler) Silence(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	var req struct {
		DurationMinutes int    `json:"duration_minutes" binding:"required,min=1"`
		Reason          string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	userID := GetCurrentUserID(c)
	if err := h.svc.Silence(c.Request.Context(), id, userID, req.DurationMinutes, req.Reason); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// BatchAcknowledge acknowledges multiple alerts at once.
func (h *AlertEventHandler) BatchAcknowledge(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	userID := GetCurrentUserID(c)
	success, failed, err := h.svc.BatchAcknowledge(c.Request.Context(), req.IDs, userID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"success": success, "failed": failed})
}

// BatchClose closes multiple alerts at once.
func (h *AlertEventHandler) BatchClose(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	userID := GetCurrentUserID(c)
	success, failed, err := h.svc.BatchClose(c.Request.Context(), req.IDs, userID)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"success": success, "failed": failed})
}

// WebhookReceive handles incoming alert webhooks (AlertManager compatible).
func (h *AlertEventHandler) WebhookReceive(c *gin.Context) {
	var payload model.AlertManagerPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	if err := h.svc.ProcessWebhook(c.Request.Context(), &payload); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/sreagent/sreagent/internal/model"
	"github.com/sreagent/sreagent/internal/service"
)

// NotificationHandler handles HTTP requests for notification channels and policies.
type NotificationHandler struct {
	svc *service.NotificationService
}

// NewNotificationHandler creates a new NotificationHandler.
func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// --- Channel Requests ---

// CreateChannelRequest is the request body for creating a notification channel.
type CreateChannelRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Type        model.NotifyChannelType `json:"type" binding:"required"`
	Description string                  `json:"description"`
	Labels      model.JSONLabels        `json:"labels"`
	Config      string                  `json:"config" binding:"required"`
	IsEnabled   *bool                   `json:"is_enabled"`
}

// UpdateChannelRequest is the request body for updating a notification channel.
type UpdateChannelRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Type        model.NotifyChannelType `json:"type" binding:"required"`
	Description string                  `json:"description"`
	Labels      model.JSONLabels        `json:"labels"`
	Config      string                  `json:"config"`
	IsEnabled   *bool                   `json:"is_enabled"`
}

// CreateChannel creates a new notification channel.
func (h *NotificationHandler) CreateChannel(c *gin.Context) {
	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}

	channel := &model.NotifyChannel{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Labels:      req.Labels,
		Config:      req.Config,
		IsEnabled:   isEnabled,
	}

	if err := h.svc.CreateChannel(c.Request.Context(), channel); err != nil {
		Error(c, err)
		return
	}

	Success(c, channel)
}

// GetChannel returns a single notification channel by ID.
func (h *NotificationHandler) GetChannel(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	channel, err := h.svc.GetChannel(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, channel)
}

// ListChannels returns a paginated list of notification channels.
func (h *NotificationHandler) ListChannels(c *gin.Context) {
	pq := GetPageQuery(c)

	list, total, err := h.svc.ListChannels(c.Request.Context(), pq.Page, pq.PageSize)
	if err != nil {
		Error(c, err)
		return
	}

	SuccessPage(c, list, total, pq.Page, pq.PageSize)
}

// UpdateChannel updates a notification channel.
func (h *NotificationHandler) UpdateChannel(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	var req UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}

	channel := &model.NotifyChannel{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
		Labels:      req.Labels,
		Config:      req.Config,
		IsEnabled:   isEnabled,
	}
	channel.ID = id

	if err := h.svc.UpdateChannel(c.Request.Context(), channel); err != nil {
		Error(c, err)
		return
	}

	Success(c, channel)
}

// DeleteChannel deletes a notification channel.
func (h *NotificationHandler) DeleteChannel(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.svc.DeleteChannel(c.Request.Context(), id); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// TestChannel sends a test notification to a channel.
func (h *NotificationHandler) TestChannel(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.svc.TestChannel(c.Request.Context(), id); err != nil {
		Error(c, err)
		return
	}

	Success(c, gin.H{"message": "test notification sent"})
}

// --- Policy Requests ---

// CreatePolicyRequest is the request body for creating a notification policy.
type CreatePolicyRequest struct {
	Name            string           `json:"name" binding:"required"`
	Description     string           `json:"description"`
	MatchLabels     model.JSONLabels `json:"match_labels" binding:"required"`
	Severities      string           `json:"severities"`
	ChannelID       uint             `json:"channel_id" binding:"required"`
	ThrottleMinutes int              `json:"throttle_minutes"`
	TemplateName    string           `json:"template_name"`
	IsEnabled       *bool            `json:"is_enabled"`
	Priority        int              `json:"priority"`
}

// UpdatePolicyRequest is the request body for updating a notification policy.
type UpdatePolicyRequest struct {
	Name            string           `json:"name" binding:"required"`
	Description     string           `json:"description"`
	MatchLabels     model.JSONLabels `json:"match_labels" binding:"required"`
	Severities      string           `json:"severities"`
	ChannelID       uint             `json:"channel_id" binding:"required"`
	ThrottleMinutes int              `json:"throttle_minutes"`
	TemplateName    string           `json:"template_name"`
	IsEnabled       *bool            `json:"is_enabled"`
	Priority        int              `json:"priority"`
}

// CreatePolicy creates a new notification policy.
func (h *NotificationHandler) CreatePolicy(c *gin.Context) {
	var req CreatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}

	templateName := req.TemplateName
	if templateName == "" {
		templateName = "default"
	}

	policy := &model.NotifyPolicy{
		Name:            req.Name,
		Description:     req.Description,
		MatchLabels:     req.MatchLabels,
		Severities:      req.Severities,
		ChannelID:       req.ChannelID,
		ThrottleMinutes: req.ThrottleMinutes,
		TemplateName:    templateName,
		IsEnabled:       isEnabled,
		Priority:        req.Priority,
	}

	if err := h.svc.CreatePolicy(c.Request.Context(), policy); err != nil {
		Error(c, err)
		return
	}

	Success(c, policy)
}

// GetPolicy returns a single notification policy by ID.
func (h *NotificationHandler) GetPolicy(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	policy, err := h.svc.GetPolicy(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, policy)
}

// ListPolicies returns a paginated list of notification policies.
func (h *NotificationHandler) ListPolicies(c *gin.Context) {
	pq := GetPageQuery(c)

	list, total, err := h.svc.ListPolicies(c.Request.Context(), pq.Page, pq.PageSize)
	if err != nil {
		Error(c, err)
		return
	}

	SuccessPage(c, list, total, pq.Page, pq.PageSize)
}

// UpdatePolicy updates a notification policy.
func (h *NotificationHandler) UpdatePolicy(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	var req UpdatePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}

	templateName := req.TemplateName
	if templateName == "" {
		templateName = "default"
	}

	policy := &model.NotifyPolicy{
		Name:            req.Name,
		Description:     req.Description,
		MatchLabels:     req.MatchLabels,
		Severities:      req.Severities,
		ChannelID:       req.ChannelID,
		ThrottleMinutes: req.ThrottleMinutes,
		TemplateName:    templateName,
		IsEnabled:       isEnabled,
		Priority:        req.Priority,
	}
	policy.ID = id

	if err := h.svc.UpdatePolicy(c.Request.Context(), policy); err != nil {
		Error(c, err)
		return
	}

	Success(c, policy)
}

// DeletePolicy deletes a notification policy.
func (h *NotificationHandler) DeletePolicy(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.svc.DeletePolicy(c.Request.Context(), id); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

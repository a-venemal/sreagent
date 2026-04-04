package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/sreagent/sreagent/internal/service"
)

// AIHandler handles AI/LLM-related API endpoints.
type AIHandler struct {
	aiSvc    *service.AIService
	eventSvc *service.AlertEventService
}

// NewAIHandler creates a new AIHandler.
func NewAIHandler(aiSvc *service.AIService, eventSvc *service.AlertEventService) *AIHandler {
	return &AIHandler{aiSvc: aiSvc, eventSvc: eventSvc}
}

// GenerateReport generates an AI-powered alert report.
func (h *AIHandler) GenerateReport(c *gin.Context) {
	var req struct {
		EventID uint `json:"event_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	event, err := h.eventSvc.GetByID(c.Request.Context(), req.EventID)
	if err != nil {
		Error(c, err)
		return
	}

	report, err := h.aiSvc.GenerateAlertReport(c.Request.Context(), event)
	if err != nil {
		ErrorWithMessage(c, 50003, err.Error())
		return
	}

	Success(c, gin.H{"report": report, "event_id": req.EventID})
}

// SuggestSOP suggests Standard Operating Procedure steps for an alert.
func (h *AIHandler) SuggestSOP(c *gin.Context) {
	var req struct {
		EventID uint `json:"event_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	event, err := h.eventSvc.GetByID(c.Request.Context(), req.EventID)
	if err != nil {
		Error(c, err)
		return
	}

	sop, err := h.aiSvc.SuggestSOP(c.Request.Context(), event)
	if err != nil {
		ErrorWithMessage(c, 50003, err.Error())
		return
	}

	Success(c, gin.H{"sop": sop, "event_id": req.EventID})
}

// GetConfig returns the current AI configuration with masked API key.
func (h *AIHandler) GetConfig(c *gin.Context) {
	cfg, err := h.aiSvc.GetConfig(c.Request.Context())
	if err != nil {
		ErrorWithMessage(c, 50003, "failed to load AI config: "+err.Error())
		return
	}
	Success(c, cfg)
}

// UpdateConfig updates the AI configuration.
func (h *AIHandler) UpdateConfig(c *gin.Context) {
	var req service.AIConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	if err := h.aiSvc.UpdateConfig(c.Request.Context(), req); err != nil {
		ErrorWithMessage(c, 50003, "failed to save AI config: "+err.Error())
		return
	}
	Success(c, gin.H{"message": "AI configuration updated"})
}

// TestConnection tests connectivity to the configured AI provider.
func (h *AIHandler) TestConnection(c *gin.Context) {
	if err := h.aiSvc.TestConnection(c.Request.Context()); err != nil {
		ErrorWithMessage(c, 50003, "AI connection test failed: "+err.Error())
		return
	}

	Success(c, gin.H{"message": "AI connection successful"})
}

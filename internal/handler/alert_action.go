package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/sreagent/sreagent/internal/repository"
	"github.com/sreagent/sreagent/internal/service"
)

// AlertActionHandler handles no-auth alert action pages (linked from Lark cards).
type AlertActionHandler struct {
	eventSvc  *service.AlertEventService
	userRepo  *repository.UserRepository
	jwtSecret string
	logger    *zap.Logger
}

// NewAlertActionHandler creates a new AlertActionHandler.
func NewAlertActionHandler(
	eventSvc *service.AlertEventService,
	userRepo *repository.UserRepository,
	jwtSecret string,
	logger *zap.Logger,
) *AlertActionHandler {
	return &AlertActionHandler{
		eventSvc:  eventSvc,
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

// ActionPage serves an HTML page for alert operations (no auth required).
// GET /alert-action/:token
func (h *AlertActionHandler) ActionPage(c *gin.Context) {
	token := c.Param("token")

	eventID, err := service.ParseAlertActionToken(token, h.jwtSecret)
	if err != nil {
		h.logger.Warn("invalid alert action token", zap.Error(err))
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusForbidden, renderErrorPage("链接无效或已过期", "该操作链接已过期（24小时有效），请从最新的告警通知中获取链接。"))
		return
	}

	event, err := h.eventSvc.GetByID(c.Request.Context(), eventID)
	if err != nil {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusNotFound, renderErrorPage("告警不存在", "未找到对应的告警事件，可能已被删除。"))
		return
	}

	// Check if there's a pre-selected action from query param
	preAction := c.Query("action")
	durationStr := c.Query("duration")
	duration := 0
	if durationStr != "" {
		duration, _ = strconv.Atoi(durationStr)
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, renderActionPage(event, token, preAction, duration))
}

// ExecuteAction handles the action form submission.
// POST /alert-action/:token
func (h *AlertActionHandler) ExecuteAction(c *gin.Context) {
	token := c.Param("token")

	eventID, err := service.ParseAlertActionToken(token, h.jwtSecret)
	if err != nil {
		h.logger.Warn("invalid alert action token on execute", zap.Error(err))
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusForbidden, renderErrorPage("链接无效或已过期", "该操作链接已过期（24小时有效），请从最新的告警通知中获取链接。"))
		return
	}

	// Parse form data (from HTML form POST)
	action := c.PostForm("action")
	operatorName := c.PostForm("operator_name")
	note := c.PostForm("note")
	durationStr := c.PostForm("duration")

	if action == "" {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusBadRequest, renderErrorPage("操作无效", "请选择一个操作。"))
		return
	}

	// Look up or create a pseudo user ID from operator name
	// For no-auth actions, we use a system user ID (0) and record the operator name in the note
	var userID uint
	if operatorName != "" {
		// Try to find user by display name
		user, findErr := h.userRepo.GetByUsername(c.Request.Context(), operatorName)
		if findErr == nil {
			userID = user.ID
		}
	}

	actionNote := note
	if operatorName != "" && actionNote == "" {
		actionNote = "操作人: " + operatorName
	} else if operatorName != "" {
		actionNote = "操作人: " + operatorName + " | " + note
	}

	var actionErr error
	var successMsg string

	switch action {
	case "acknowledge":
		actionErr = h.eventSvc.Acknowledge(c.Request.Context(), eventID, userID)
		successMsg = "告警已认领"
		if actionNote != "" {
			if commentErr := h.eventSvc.AddComment(c.Request.Context(), eventID, userID, actionNote); commentErr != nil {
				h.logger.Warn("failed to add action comment", zap.Uint("event_id", eventID), zap.Error(commentErr))
			}
		}
	case "silence":
		duration := 60 // default 1 hour
		if durationStr != "" {
			if d, parseErr := strconv.Atoi(durationStr); parseErr == nil && d > 0 {
				duration = d
			}
		}
		reason := actionNote
		if reason == "" {
			reason = "Silenced from Lark card"
		}
		actionErr = h.eventSvc.Silence(c.Request.Context(), eventID, userID, duration, reason)
		successMsg = "告警已静默"
	case "resolve":
		resolution := actionNote
		if resolution == "" {
			resolution = "Resolved from Lark card"
		}
		actionErr = h.eventSvc.Resolve(c.Request.Context(), eventID, userID, resolution)
		successMsg = "告警已解决"
	case "close":
		closeNote := actionNote
		if closeNote == "" {
			closeNote = "Closed from Lark card"
		}
		actionErr = h.eventSvc.Close(c.Request.Context(), eventID, userID, closeNote)
		successMsg = "告警已关闭"
	default:
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusBadRequest, renderErrorPage("操作无效", "不支持的操作类型: "+action))
		return
	}

	if actionErr != nil {
		h.logger.Error("alert action failed",
			zap.Uint("event_id", eventID),
			zap.String("action", action),
			zap.Error(actionErr),
		)
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, renderResultPage(false, "操作失败", actionErr.Error()))
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, renderResultPage(true, successMsg, ""))
}

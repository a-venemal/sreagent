package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/sreagent/sreagent/internal/model"
)

type DashboardHandler struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewDashboardHandler(db *gorm.DB, logger *zap.Logger) *DashboardHandler {
	return &DashboardHandler{db: db, logger: logger}
}

// DashboardStats represents the aggregated dashboard statistics.
type DashboardStats struct {
	TotalDatasources int64 `json:"total_datasources"`
	TotalRules       int64 `json:"total_rules"`
	ActiveAlerts     int64 `json:"active_alerts"`
	ResolvedToday    int64 `json:"resolved_today"`
	TotalUsers       int64 `json:"total_users"`
	TotalTeams       int64 `json:"total_teams"`
}

// GetStats returns aggregated dashboard statistics.
func (h *DashboardHandler) GetStats(c *gin.Context) {
	var stats DashboardStats

	// Total datasources
	if err := h.db.Model(&model.DataSource{}).Count(&stats.TotalDatasources).Error; err != nil {
		h.logger.Error("failed to count datasources", zap.Error(err))
	}

	// Total alert rules
	if err := h.db.Model(&model.AlertRule{}).Count(&stats.TotalRules).Error; err != nil {
		h.logger.Error("failed to count alert rules", zap.Error(err))
	}

	// Active alerts (firing + acknowledged)
	if err := h.db.Model(&model.AlertEvent{}).
		Where("status IN ?", []string{
			string(model.EventStatusFiring),
			string(model.EventStatusAcknowledged),
		}).
		Count(&stats.ActiveAlerts).Error; err != nil {
		h.logger.Error("failed to count active alerts", zap.Error(err))
	}

	// Resolved today
	todayStart := time.Now().Truncate(24 * time.Hour)
	if err := h.db.Model(&model.AlertEvent{}).
		Where("status = ? AND resolved_at >= ?", string(model.EventStatusResolved), todayStart).
		Count(&stats.ResolvedToday).Error; err != nil {
		h.logger.Error("failed to count resolved alerts today", zap.Error(err))
	}

	// Total users
	if err := h.db.Model(&model.User{}).Count(&stats.TotalUsers).Error; err != nil {
		h.logger.Error("failed to count users", zap.Error(err))
	}

	// Total teams
	if err := h.db.Model(&model.Team{}).Count(&stats.TotalTeams).Error; err != nil {
		h.logger.Error("failed to count teams", zap.Error(err))
	}

	Success(c, stats)
}

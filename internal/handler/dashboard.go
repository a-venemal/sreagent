package handler

import (
	"strconv"
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
	TotalDatasources  int64            `json:"total_datasources"`
	TotalRules        int64            `json:"total_rules"`
	ActiveAlerts      int64            `json:"active_alerts"`
	ResolvedToday     int64            `json:"resolved_today"`
	TotalUsers        int64            `json:"total_users"`
	TotalTeams        int64            `json:"total_teams"`
	// SeverityBreakdown holds the count of active (firing+acked) alerts per severity.
	SeverityBreakdown map[string]int64 `json:"severity_breakdown"`
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

	// Severity breakdown of active alerts
	type sevRow struct {
		Severity string
		Cnt      int64
	}
	var sevRows []sevRow
	h.db.Model(&model.AlertEvent{}).
		Select("severity, COUNT(*) AS cnt").
		Where("status IN ?", []string{
			string(model.EventStatusFiring),
			string(model.EventStatusAcknowledged),
		}).
		Group("severity").
		Scan(&sevRows)
	stats.SeverityBreakdown = map[string]int64{
		"critical": 0,
		"warning":  0,
		"info":     0,
	}
	for _, r := range sevRows {
		stats.SeverityBreakdown[r.Severity] = r.Cnt
	}

	Success(c, stats)
}

// MTTRStats holds mean time to acknowledge and mean time to resolve, in seconds.
type MTTRStats struct {
	// Window is the query window in hours.
	WindowHours int `json:"window_hours"`
	// MTTA is the mean time to acknowledge in seconds (-1 if no data).
	MTTA float64 `json:"mtta_seconds"`
	// MTTR is the mean time to resolve in seconds (-1 if no data).
	MTTR float64 `json:"mttr_seconds"`
	// AckedCount is the number of acknowledged events in the window.
	AckedCount int64 `json:"acked_count"`
	// ResolvedCount is the number of resolved events in the window.
	ResolvedCount int64 `json:"resolved_count"`
}

// GetMTTRStats returns MTTA and MTTR aggregated over a configurable time window.
// Query param: hours (default 24, accepts 1/6/24/168/720)
func (h *DashboardHandler) GetMTTRStats(c *gin.Context) {
	hours := 24
	if v := c.Query("hours"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			hours = n
		}
	}

	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	var stats MTTRStats
	stats.WindowHours = hours
	stats.MTTA = -1
	stats.MTTR = -1

	// MTTA: average seconds from fired_at to acked_at
	type aggResult struct {
		AvgSeconds float64
		Cnt        int64
	}

	var mttaRes aggResult
	h.db.Model(&model.AlertEvent{}).
		Select("AVG(TIMESTAMPDIFF(SECOND, fired_at, acked_at)) AS avg_seconds, COUNT(*) AS cnt").
		Where("acked_at IS NOT NULL AND fired_at >= ?", since).
		Scan(&mttaRes)
	if mttaRes.Cnt > 0 {
		stats.MTTA = mttaRes.AvgSeconds
		stats.AckedCount = mttaRes.Cnt
	}

	// MTTR: average seconds from fired_at to resolved_at
	var mttrRes aggResult
	h.db.Model(&model.AlertEvent{}).
		Select("AVG(TIMESTAMPDIFF(SECOND, fired_at, resolved_at)) AS avg_seconds, COUNT(*) AS cnt").
		Where("resolved_at IS NOT NULL AND fired_at >= ?", since).
		Scan(&mttrRes)
	if mttrRes.Cnt > 0 {
		stats.MTTR = mttrRes.AvgSeconds
		stats.ResolvedCount = mttrRes.Cnt
	}

	Success(c, stats)
}

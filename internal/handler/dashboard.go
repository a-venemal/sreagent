package handler

import (
	"sort"
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

// AlertTrendPoint represents a data point for the alert trend chart.
type AlertTrendPoint struct {
	Date          string `json:"date"`
	FiredCount    int64  `json:"fired_count"`
	ResolvedCount int64  `json:"resolved_count"`
}

// GetAlertTrend returns daily fired/resolved counts for trend charts.
// GET /api/v1/dashboard/alert-trend?days=30
func (h *DashboardHandler) GetAlertTrend(c *gin.Context) {
	days := 30
	if v := c.Query("days"); v != "" {
		if n, _ := strconv.Atoi(v); n > 0 && n <= 365 {
			days = n
		}
	}
	since := time.Now().AddDate(0, 0, -days)

	type dateCount struct {
		Date string
		Cnt  int64
	}

	// Query fired counts per day
	var firedRows []dateCount
	h.db.Model(&model.AlertEvent{}).
		Select("DATE(fired_at) AS date, COUNT(*) AS cnt").
		Where("fired_at >= ? AND deleted_at IS NULL", since).
		Group("DATE(fired_at)").Order("date").Scan(&firedRows)

	// Query resolved counts per day
	var resolvedRows []dateCount
	h.db.Model(&model.AlertEvent{}).
		Select("DATE(resolved_at) AS date, COUNT(*) AS cnt").
		Where("resolved_at >= ? AND resolved_at IS NOT NULL AND deleted_at IS NULL", since).
		Group("DATE(resolved_at)").Order("date").Scan(&resolvedRows)

	// Merge into result
	resolvedMap := map[string]int64{}
	for _, r := range resolvedRows {
		resolvedMap[r.Date] = r.Cnt
	}

	result := make([]AlertTrendPoint, 0, len(firedRows))
	for _, f := range firedRows {
		result = append(result, AlertTrendPoint{
			Date: f.Date, FiredCount: f.Cnt, ResolvedCount: resolvedMap[f.Date],
		})
	}
	Success(c, result)
}

// TopRuleItem represents a rule with its alert count for the top-rules endpoint.
type TopRuleItem struct {
	RuleID    *uint  `json:"rule_id"`
	AlertName string `json:"alert_name"`
	Count     int64  `json:"count"`
}

// GetTopRules returns the most frequently firing alert rules.
// GET /api/v1/dashboard/top-rules?days=30&limit=10
func (h *DashboardHandler) GetTopRules(c *gin.Context) {
	days := 30
	if v := c.Query("days"); v != "" {
		if n, _ := strconv.Atoi(v); n > 0 {
			days = n
		}
	}
	limit := 10
	if v := c.Query("limit"); v != "" {
		if n, _ := strconv.Atoi(v); n > 0 && n <= 50 {
			limit = n
		}
	}
	since := time.Now().AddDate(0, 0, -days)

	var items []TopRuleItem
	h.db.Model(&model.AlertEvent{}).
		Select("rule_id, alert_name, COUNT(*) AS count").
		Where("fired_at >= ? AND deleted_at IS NULL", since).
		Group("rule_id, alert_name").
		Order("count DESC").
		Limit(limit).
		Scan(&items)
	Success(c, items)
}

// SeverityHistoryPoint represents per-severity alert counts for a single day.
type SeverityHistoryPoint struct {
	Date   string         `json:"date"`
	Counts map[string]int64 `json:"counts"`
}

// GetSeverityHistory returns daily alert counts broken down by severity.
// GET /api/v1/dashboard/severity-history?days=30
func (h *DashboardHandler) GetSeverityHistory(c *gin.Context) {
	days := 30
	if v := c.Query("days"); v != "" {
		if n, _ := strconv.Atoi(v); n > 0 {
			days = n
		}
	}
	since := time.Now().AddDate(0, 0, -days)

	type row struct {
		Date     string
		Severity string
		Cnt      int64
	}
	var rows []row
	h.db.Model(&model.AlertEvent{}).
		Select("DATE(fired_at) AS date, severity, COUNT(*) AS cnt").
		Where("fired_at >= ? AND deleted_at IS NULL", since).
		Group("DATE(fired_at), severity").
		Order("date").
		Scan(&rows)

	dateMap := map[string]map[string]int64{}
	for _, r := range rows {
		if dateMap[r.Date] == nil {
			dateMap[r.Date] = map[string]int64{"critical": 0, "warning": 0, "info": 0}
		}
		dateMap[r.Date][r.Severity] = r.Cnt
	}

	// Sort dates
	dates := make([]string, 0, len(dateMap))
	for d := range dateMap {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	result := make([]SeverityHistoryPoint, 0, len(dates))
	for _, d := range dates {
		result = append(result, SeverityHistoryPoint{Date: d, Counts: dateMap[d]})
	}
	Success(c, result)
}

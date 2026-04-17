package service

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"github.com/sreagent/sreagent/internal/model"
	apperr "github.com/sreagent/sreagent/internal/pkg/errors"
	"github.com/sreagent/sreagent/internal/repository"
)

type AlertRuleService struct {
	repo        *repository.AlertRuleRepository
	historyRepo *repository.AlertRuleHistoryRepository
	dsRepo      *repository.DataSourceRepository
	logger      *zap.Logger
}

func NewAlertRuleService(
	repo *repository.AlertRuleRepository,
	historyRepo *repository.AlertRuleHistoryRepository,
	dsRepo *repository.DataSourceRepository,
	logger *zap.Logger,
) *AlertRuleService {
	return &AlertRuleService{repo: repo, historyRepo: historyRepo, dsRepo: dsRepo, logger: logger}
}

func (s *AlertRuleService) Create(ctx context.Context, rule *model.AlertRule) error {
	// Verify datasource exists
	if _, err := s.dsRepo.GetByID(ctx, rule.DataSourceID); err != nil {
		return apperr.WithMessage(apperr.ErrDSNotFound, fmt.Sprintf("datasource ID %d not found", rule.DataSourceID))
	}

	rule.Version = 1
	if err := s.repo.Create(ctx, rule); err != nil {
		s.logger.Error("failed to create alert rule", zap.Error(err))
		return apperr.Wrap(apperr.ErrDatabase, err)
	}

	s.recordHistory(ctx, rule, "created")
	return nil
}

func (s *AlertRuleService) GetByID(ctx context.Context, id uint) (*model.AlertRule, error) {
	rule, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperr.ErrRuleNotFound
	}
	return rule, nil
}

func (s *AlertRuleService) List(ctx context.Context, severity, status, groupName, category string, page, pageSize int) ([]model.AlertRule, int64, error) {
	return s.repo.List(ctx, severity, status, groupName, category, page, pageSize)
}

// ListCategories returns all distinct non-empty category values.
func (s *AlertRuleService) ListCategories(ctx context.Context) ([]string, error) {
	return s.repo.ListCategories(ctx)
}

func (s *AlertRuleService) Update(ctx context.Context, rule *model.AlertRule) error {
	existing, err := s.repo.GetByID(ctx, rule.ID)
	if err != nil {
		return apperr.ErrRuleNotFound
	}

	existing.Name = rule.Name
	existing.DisplayName = rule.DisplayName
	existing.Description = rule.Description
	existing.DataSourceID = rule.DataSourceID
	existing.Expression = rule.Expression
	existing.ForDuration = rule.ForDuration
	existing.Severity = rule.Severity
	existing.Labels = rule.Labels
	existing.Annotations = rule.Annotations
	existing.GroupName = rule.GroupName
	existing.Category = rule.Category
	existing.UpdatedBy = rule.UpdatedBy
	existing.EvalInterval = rule.EvalInterval
	existing.RecoveryHold = rule.RecoveryHold
	existing.NoDataEnabled = rule.NoDataEnabled
	existing.NoDataDuration = rule.NoDataDuration
	existing.SuppressEnabled = rule.SuppressEnabled
	existing.BizGroupID = rule.BizGroupID
	existing.Version++

	if err := s.repo.Update(ctx, existing); err != nil {
		s.logger.Error("failed to update alert rule", zap.Error(err))
		return apperr.Wrap(apperr.ErrDatabase, err)
	}

	s.recordHistory(ctx, existing, "updated")
	return nil
}

func (s *AlertRuleService) Delete(ctx context.Context, id uint) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return apperr.ErrRuleNotFound
	}

	s.recordHistory(ctx, existing, "deleted")

	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.Error("failed to delete alert rule", zap.Error(err))
		return apperr.Wrap(apperr.ErrDatabase, err)
	}

	return nil
}

// ImportRules batch-creates alert rules, returning success/failed counts and error details.
func (s *AlertRuleService) ImportRules(ctx context.Context, rules []model.AlertRule) (success, failed int, errors []string) {
	for i, rule := range rules {
		rule.Version = 1
		if err := s.repo.Create(ctx, &rule); err != nil {
			failed++
			errors = append(errors, fmt.Sprintf("rule #%d (%s): %v", i+1, rule.Name, err))
			s.logger.Error("failed to import alert rule",
				zap.String("name", rule.Name),
				zap.Error(err),
			)
		} else {
			success++
			s.recordHistory(ctx, &rule, "created")
		}
	}

	return
}

func (s *AlertRuleService) UpdateStatus(ctx context.Context, id uint, status model.AlertRuleStatus) error {
	rule, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return apperr.ErrRuleNotFound
	}

	rule.Status = status
	rule.Version++
	if err := s.repo.Update(ctx, rule); err != nil {
		return apperr.Wrap(apperr.ErrDatabase, err)
	}

	s.recordHistory(ctx, rule, "updated")
	return nil
}

// recordHistory creates an audit trail entry for an alert rule change.
func (s *AlertRuleService) recordHistory(ctx context.Context, rule *model.AlertRule, changeType string) {
	if s.historyRepo == nil {
		return
	}

	snapshot, err := json.Marshal(rule)
	if err != nil {
		s.logger.Error("failed to marshal rule snapshot for history",
			zap.Uint("rule_id", rule.ID),
			zap.Error(err),
		)
		return
	}

	h := &model.AlertRuleHistory{
		RuleID:     rule.ID,
		Version:    rule.Version,
		ChangeType: changeType,
		Snapshot:   string(snapshot),
		ChangedBy:  rule.UpdatedBy,
	}
	// For create operations, ChangedBy comes from CreatedBy
	if changeType == "created" {
		h.ChangedBy = rule.CreatedBy
	}

	if err := s.historyRepo.Create(ctx, h); err != nil {
		s.logger.Error("failed to record alert rule history",
			zap.Uint("rule_id", rule.ID),
			zap.String("change_type", changeType),
			zap.Error(err),
		)
	}
}

// ListHistory returns paginated history records for a given rule.
func (s *AlertRuleService) ListHistory(ctx context.Context, ruleID uint, page, pageSize int) ([]model.AlertRuleHistory, int64, error) {
	if s.historyRepo == nil {
		return nil, 0, nil
	}
	return s.historyRepo.ListByRuleID(ctx, ruleID, page, pageSize)
}

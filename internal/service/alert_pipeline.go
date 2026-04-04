package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/sreagent/sreagent/internal/model"
)

// AlertPipeline orchestrates the full alert analysis flow:
// 1. Build context (pull metrics)
// 2. Call LLM for analysis
// 3. Return enriched result for notification
type AlertPipeline struct {
	contextBuilder *AlertContextBuilder
	aiSvc          *AIService
	logger         *zap.Logger
}

// NewAlertPipeline creates a new AlertPipeline.
func NewAlertPipeline(contextBuilder *AlertContextBuilder, aiSvc *AIService, logger *zap.Logger) *AlertPipeline {
	return &AlertPipeline{
		contextBuilder: contextBuilder,
		aiSvc:          aiSvc,
		logger:         logger,
	}
}

// AnalyzeAlert runs the full pipeline for an alert event.
// If AI is disabled or fails, it returns nil analysis (not an error) so
// notification can still proceed with just the basic alert card.
func (p *AlertPipeline) AnalyzeAlert(ctx context.Context, event *model.AlertEvent) *AlertAnalysis {
	// 1. Build context
	alertCtx, err := p.contextBuilder.BuildContext(ctx, event)
	if err != nil {
		p.logger.Warn("failed to build alert context, proceeding without metrics",
			zap.Uint("event_id", event.ID),
			zap.Error(err),
		)
		// Fall back to basic context without metrics
		alertCtx = &AlertContext{
			AlertName:   event.AlertName,
			Severity:    string(event.Severity),
			Labels:      event.Labels,
			Annotations: event.Annotations,
			FiredAt:     event.FiredAt,
		}
		alertCtx.ContextText = formatBasicContext(alertCtx)
	}

	// 2. Call LLM
	analysis, err := p.aiSvc.AnalyzeAlertWithContext(ctx, alertCtx.ContextText)
	if err != nil {
		p.logger.Warn("LLM analysis failed, notification will proceed without AI analysis",
			zap.Uint("event_id", event.ID),
			zap.Error(err),
		)
		return nil
	}

	return analysis
}

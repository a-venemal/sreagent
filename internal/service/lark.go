package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/sreagent/sreagent/internal/model"
	"github.com/sreagent/sreagent/internal/pkg/lark"
)

// LarkService wraps the Lark client for sending alert notifications.
type LarkService struct {
	client *lark.Client
	logger *zap.Logger
	// platformBaseURL is the base URL of the SREAgent web UI for deep-linking.
	platformBaseURL string
	// jwtSecret is used to sign alert action tokens.
	jwtSecret string
}

// NewLarkService creates a new LarkService.
func NewLarkService(logger *zap.Logger, platformBaseURL, jwtSecret string) *LarkService {
	return &LarkService{
		client:          lark.NewClient(logger),
		logger:          logger,
		platformBaseURL: platformBaseURL,
		jwtSecret:       jwtSecret,
	}
}

// SendAlertNotification prepares and sends an alert notification via Lark webhook.
func (s *LarkService) SendAlertNotification(ctx context.Context, event *model.AlertEvent, webhookURL string) error {
	// Build the platform link for this alert event
	platformURL := ""
	if s.platformBaseURL != "" {
		platformURL = fmt.Sprintf("%s/alert-events/%d", s.platformBaseURL, event.ID)
	}

	card := lark.BuildAlertCard(
		event.AlertName,
		string(event.Severity),
		string(event.Status),
		event.Labels,
		event.Annotations,
		event.FiredAt,
		platformURL,
	)

	resp, err := s.client.SendWebhook(ctx, webhookURL, card)
	if err != nil {
		s.logger.Error("failed to send lark alert notification",
			zap.Uint("event_id", event.ID),
			zap.String("alert_name", event.AlertName),
			zap.Error(err),
		)
		return fmt.Errorf("lark webhook failed: %w", err)
	}

	s.logger.Info("lark alert notification sent",
		zap.Uint("event_id", event.ID),
		zap.String("alert_name", event.AlertName),
		zap.Int("resp_code", resp.Code),
	)
	return nil
}

// SendEnrichedAlertNotification sends an alert notification with AI analysis via Lark webhook.
func (s *LarkService) SendEnrichedAlertNotification(ctx context.Context, event *model.AlertEvent, analysis *AlertAnalysis, webhookURL string) error {
	// Build the platform link for this alert event
	platformURL := ""
	if s.platformBaseURL != "" {
		platformURL = fmt.Sprintf("%s/alert-events/%d", s.platformBaseURL, event.ID)
	}

	// Generate an action token for no-auth alert action page
	actionBaseURL := ""
	if s.platformBaseURL != "" && s.jwtSecret != "" {
		token, err := GenerateAlertActionToken(event.ID, s.jwtSecret)
		if err != nil {
			s.logger.Warn("failed to generate alert action token",
				zap.Uint("event_id", event.ID),
				zap.Error(err),
			)
		} else {
			actionBaseURL = fmt.Sprintf("%s/alert-action/%s", s.platformBaseURL, token)
		}
	}

	// Convert service.AlertAnalysis to lark.AIAnalysisResult (nil-safe)
	var aiResult *lark.AIAnalysisResult
	if analysis != nil {
		aiResult = &lark.AIAnalysisResult{
			Summary:          analysis.Summary,
			ProbableCauses:   analysis.ProbableCauses,
			Impact:           analysis.Impact,
			RecommendedSteps: analysis.RecommendedSteps,
		}
	}

	card := lark.BuildEnrichedAlertCard(
		event.AlertName,
		string(event.Severity),
		string(event.Status),
		event.Labels,
		event.Annotations,
		event.FiredAt,
		aiResult,
		platformURL,
		actionBaseURL,
	)

	resp, err := s.client.SendWebhook(ctx, webhookURL, card)
	if err != nil {
		s.logger.Error("failed to send enriched lark alert notification",
			zap.Uint("event_id", event.ID),
			zap.String("alert_name", event.AlertName),
			zap.Error(err),
		)
		return fmt.Errorf("lark webhook failed: %w", err)
	}

	s.logger.Info("enriched lark alert notification sent",
		zap.Uint("event_id", event.ID),
		zap.String("alert_name", event.AlertName),
		zap.Int("resp_code", resp.Code),
		zap.Bool("has_ai_analysis", analysis != nil),
	)
	return nil
}

// SendTestNotification sends a test card to the given webhook URL.
func (s *LarkService) SendTestNotification(ctx context.Context, webhookURL string) error {
	card := lark.BuildTestCard()

	_, err := s.client.SendWebhook(ctx, webhookURL, card)
	if err != nil {
		s.logger.Error("failed to send lark test notification", zap.Error(err))
		return fmt.Errorf("lark test webhook failed: %w", err)
	}

	s.logger.Info("lark test notification sent successfully")
	return nil
}

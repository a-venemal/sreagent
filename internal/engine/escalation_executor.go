package engine

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/sreagent/sreagent/internal/model"
	"github.com/sreagent/sreagent/internal/repository"
	"github.com/sreagent/sreagent/internal/service"
)

// EscalationExecutor periodically checks firing alert events and executes escalation steps
// when the configured delay has elapsed and the alert has not yet been resolved or acknowledged.
type EscalationExecutor struct {
	policyRepo   *repository.EscalationPolicyRepository
	stepRepo     *repository.EscalationStepRepository
	eventRepo    *repository.AlertEventRepository
	timelineRepo *repository.AlertTimelineRepository
	channelRepo  *repository.NotifyChannelRepository
	userRepo     *repository.UserRepository
	notifySvc    *service.NotificationService
	logger       *zap.Logger

	interval time.Duration
	stopCh   chan struct{}
	once     sync.Once
}

// NewEscalationExecutor creates a new EscalationExecutor.
func NewEscalationExecutor(
	policyRepo *repository.EscalationPolicyRepository,
	stepRepo *repository.EscalationStepRepository,
	eventRepo *repository.AlertEventRepository,
	timelineRepo *repository.AlertTimelineRepository,
	channelRepo *repository.NotifyChannelRepository,
	userRepo *repository.UserRepository,
	notifySvc *service.NotificationService,
	logger *zap.Logger,
) *EscalationExecutor {
	return &EscalationExecutor{
		policyRepo:   policyRepo,
		stepRepo:     stepRepo,
		eventRepo:    eventRepo,
		timelineRepo: timelineRepo,
		channelRepo:  channelRepo,
		userRepo:     userRepo,
		notifySvc:    notifySvc,
		logger:       logger,
		interval:     60 * time.Second,
		stopCh:       make(chan struct{}),
	}
}

// SetInterval overrides the default 60-second check interval.
func (e *EscalationExecutor) SetInterval(d time.Duration) {
	e.interval = d
}

// Start runs the escalation check loop in a background goroutine.
func (e *EscalationExecutor) Start() {
	go func() {
		ticker := time.NewTicker(e.interval)
		defer ticker.Stop()
		e.logger.Info("escalation executor started", zap.Duration("interval", e.interval))
		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
				e.runOnce(ctx)
				cancel()
			case <-e.stopCh:
				e.logger.Info("escalation executor stopped")
				return
			}
		}
	}()
}

// Stop signals the background goroutine to exit.
func (e *EscalationExecutor) Stop() {
	e.once.Do(func() {
		select {
		case <-e.stopCh:
		default:
			close(e.stopCh)
		}
	})
}

// runOnce performs a single escalation check pass.
func (e *EscalationExecutor) runOnce(ctx context.Context) {
	// Fetch all currently active (firing or acknowledged) events — use a large page.
	events, _, err := e.eventRepo.List(ctx, "", "", 1, 10000)
	if err != nil {
		e.logger.Error("escalation: failed to list events", zap.Error(err))
		return
	}

	now := time.Now()
	for i := range events {
		ev := &events[i]
		// Only escalate firing events that haven't been resolved/closed/silenced.
		switch ev.Status {
		case model.EventStatusFiring:
			// OK — escalate
		default:
			continue
		}

		e.escalateEvent(ctx, ev, now)
	}
}

// escalateEvent evaluates all escalation policies and executes any due steps for the given event.
func (e *EscalationExecutor) escalateEvent(ctx context.Context, event *model.AlertEvent, now time.Time) {
	// Determine which escalation steps have already been executed by inspecting the timeline.
	executedSteps := e.executedStepOrders(ctx, event.ID)

	// Collect all enabled policies across all teams and find matching steps.
	// In a full implementation we would match policies to the event's team/labels.
	// For now we evaluate all enabled policies.
	policies, err := e.listAllEnabledPolicies(ctx)
	if err != nil {
		e.logger.Warn("escalation: failed to list policies", zap.Error(err))
		return
	}

	for _, policy := range policies {
		steps, err := e.stepRepo.ListByPolicyID(ctx, policy.ID)
		if err != nil {
			e.logger.Warn("escalation: failed to list steps",
				zap.Uint("policy_id", policy.ID), zap.Error(err))
			continue
		}

		// Sort by step order to execute in sequence.
		sort.Slice(steps, func(i, j int) bool {
			return steps[i].StepOrder < steps[j].StepOrder
		})

		for _, step := range steps {
			stepKey := fmt.Sprintf("escalation policy '%s' step %d triggered (delay: %dm)",
				policy.Name, step.StepOrder, step.DelayMinutes)
			if executedSteps[stepKey] {
				// Already executed this step for this event.
				continue
			}

			// Check if enough time has passed since the alert fired.
			dueAt := event.FiredAt.Add(time.Duration(step.DelayMinutes) * time.Minute)
			if now.Before(dueAt) {
				// Not due yet; later steps will be even less due.
				break
			}

			// Execute this step.
			if err := e.executeStep(ctx, event, &policy, &step); err != nil {
				e.logger.Error("escalation: failed to execute step",
					zap.Uint("event_id", event.ID),
					zap.Uint("policy_id", policy.ID),
					zap.Int("step_order", step.StepOrder),
					zap.Error(err),
				)
				// Record failure in timeline so we don't retry endlessly this cycle.
				e.recordTimeline(ctx, event.ID, fmt.Sprintf(
					"escalation step %d (policy %s) failed: %v", step.StepOrder, policy.Name, err,
				))
			}
		}
	}
}

// executeStep dispatches a notification for a single escalation step.
func (e *EscalationExecutor) executeStep(ctx context.Context, event *model.AlertEvent, policy *model.EscalationPolicy, step *model.EscalationStep) error {
	// This note is also used as the dedup key in executedStepOrders — keep format in sync.
	note := fmt.Sprintf("escalation policy '%s' step %d triggered (delay: %dm)",
		policy.Name, step.StepOrder, step.DelayMinutes)

	// Resolve the notification channel: prefer the step's override channel, then fall
	// back to notifying the target user/team directly via a system message.
	if step.NotifyChannelID != nil {
		channel, err := e.channelRepo.GetByID(ctx, *step.NotifyChannelID)
		if err != nil {
			return fmt.Errorf("channel %d not found: %w", *step.NotifyChannelID, err)
		}
		if err := e.notifySvc.SendNotification(ctx, event, channel, nil, nil); err != nil {
			return fmt.Errorf("send notification via channel %d: %w", *step.NotifyChannelID, err)
		}
	} else {
		// No channel override — log the escalation (target lookup via user/team/schedule
		// would require those repos; a production implementation would dispatch accordingly).
		e.logger.Info("escalation: no channel configured for step, logging escalation only",
			zap.Uint("event_id", event.ID),
			zap.String("target_type", step.TargetType),
			zap.Uint("target_id", step.TargetID),
		)
	}

	// Record the escalation in the timeline so we don't repeat this step.
	e.recordTimeline(ctx, event.ID, note)

	e.logger.Info("escalation step executed",
		zap.Uint("event_id", event.ID),
		zap.String("policy", policy.Name),
		zap.Int("step_order", step.StepOrder),
	)
	return nil
}

// executedStepOrders returns a set of "policyID:stepOrder" keys already recorded in the
// event's timeline with action=escalated.
func (e *EscalationExecutor) executedStepOrders(ctx context.Context, eventID uint) map[string]bool {
	timelines, err := e.timelineRepo.ListByEventID(ctx, eventID)
	if err != nil {
		return map[string]bool{}
	}
	result := make(map[string]bool)
	for _, t := range timelines {
		if t.Action == model.TimelineActionEscalated {
			// The note encodes the step identity — extract the key from the note prefix.
			// We use the note text as the de-dup key directly.
			result[t.Note] = true
		}
	}
	return result
}

// recordTimeline appends an escalation action to the event timeline.
func (e *EscalationExecutor) recordTimeline(ctx context.Context, eventID uint, note string) {
	t := &model.AlertTimeline{
		EventID: eventID,
		Action:  model.TimelineActionEscalated,
		Note:    note,
	}
	if err := e.timelineRepo.Create(ctx, t); err != nil {
		e.logger.Error("escalation: failed to record timeline",
			zap.Uint("event_id", eventID), zap.Error(err))
	}
}

// listAllEnabledPolicies returns all enabled EscalationPolicy records.
// ListByTeamID with teamID=0 skips the team filter and returns all policies.
func (e *EscalationExecutor) listAllEnabledPolicies(ctx context.Context) ([]model.EscalationPolicy, error) {
	all, err := e.policyRepo.ListByTeamID(ctx, 0)
	if err != nil {
		return nil, err
	}
	enabled := all[:0]
	for _, p := range all {
		if p.IsEnabled {
			enabled = append(enabled, p)
		}
	}
	return enabled, nil
}

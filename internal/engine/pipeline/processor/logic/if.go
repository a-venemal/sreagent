package logic

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sreagent/sreagent/internal/engine/pipeline"
	"github.com/sreagent/sreagent/internal/model"
)

func init() {
	pipeline.RegisterProcessor("if", newIfProcessor)
}

// IfConfig configures the "if" branch processor.
type IfConfig struct {
	Mode       string        `json:"mode"` // "tags" or "expression"
	TagFilters []TagFilter   `json:"tag_filters,omitempty"`
	Expression *ExpressionOp `json:"expression,omitempty"`
}

type TagFilter struct {
	Key   string `json:"key"`
	Op    string `json:"op"` // ==, !=, =~, !~, in, not_in
	Value string `json:"value"`
}

type ExpressionOp struct {
	Field string `json:"field"` // severity, status, source, alert_name
	Op    string `json:"op"`    // ==, !=, =~, !~
	Value string `json:"value"`
}

type ifProcessor struct {
	config IfConfig
}

func newIfProcessor(config map[string]interface{}) (pipeline.Processor, error) {
	cfg := parseIfConfig(config)
	return &ifProcessor{config: cfg}, nil
}

func (p *ifProcessor) Process(wfCtx *pipeline.WorkflowContext) (*pipeline.WorkflowContext, string, error) {
	return wfCtx, "", nil
}

func (p *ifProcessor) ProcessWithBranch(wfCtx *pipeline.WorkflowContext) (*pipeline.NodeOutput, error) {
	result := p.evaluate(wfCtx)
	branchIdx := 0
	if result {
		branchIdx = 0 // true branch
	} else {
		branchIdx = 1 // false branch
	}
	msg := fmt.Sprintf("condition evaluated to %v (branch %d)", result, branchIdx)
	return &pipeline.NodeOutput{
		WfCtx:       wfCtx,
		Message:     msg,
		BranchIndex: branchIdx,
	}, nil
}

func (p *ifProcessor) evaluate(wfCtx *pipeline.WorkflowContext) bool {
	event := wfCtx.Event
	if event == nil {
		return false
	}

	switch p.config.Mode {
	case "expression":
		if p.config.Expression != nil {
			return p.evaluateExpression(event, p.config.Expression)
		}
		return false
	default: // "tags"
		return p.evaluateTagFilters(event)
	}
}

func (p *ifProcessor) evaluateExpression(event *model.AlertEvent, expr *ExpressionOp) bool {
	var fieldValue string
	switch expr.Field {
	case "severity":
		fieldValue = string(event.Severity)
	case "status":
		fieldValue = string(event.Status)
	case "source":
		fieldValue = event.Source
	case "alert_name":
		fieldValue = event.AlertName
	default:
		return false
	}
	return matchOp(fieldValue, expr.Op, expr.Value)
}

func (p *ifProcessor) evaluateTagFilters(event *model.AlertEvent) bool {
	labels := event.Labels
	for _, tf := range p.config.TagFilters {
		val, ok := labels[tf.Key]
		if !ok {
			if tf.Op == "!=" || tf.Op == "!~" || tf.Op == "not_in" {
				continue
			}
			return false
		}
		if !matchOp(val, tf.Op, tf.Value) {
			return false
		}
	}
	return true
}

func matchOp(fieldValue, op, filterValue string) bool {
	switch op {
	case "==":
		return fieldValue == filterValue
	case "!=":
		return fieldValue != filterValue
	case "=~":
		matched, _ := regexp.MatchString(filterValue, fieldValue)
		return matched
	case "!~":
		matched, _ := regexp.MatchString(filterValue, fieldValue)
		return !matched
	case "in":
		for _, v := range strings.Split(filterValue, ",") {
			if fieldValue == strings.TrimSpace(v) {
				return true
			}
		}
		return false
	case "not_in":
		for _, v := range strings.Split(filterValue, ",") {
			if fieldValue == strings.TrimSpace(v) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func parseIfConfig(config map[string]interface{}) IfConfig {
	cfg := IfConfig{Mode: "tags"}
	if m, ok := config["mode"].(string); ok {
		cfg.Mode = m
	}
	if filters, ok := config["tag_filters"].([]interface{}); ok {
		for _, f := range filters {
			if fm, ok := f.(map[string]interface{}); ok {
				tf := TagFilter{}
				if v, ok := fm["key"].(string); ok {
					tf.Key = v
				}
				if v, ok := fm["op"].(string); ok {
					tf.Op = v
				} else {
					tf.Op = "=="
				}
				if v, ok := fm["value"].(string); ok {
					tf.Value = v
				}
				cfg.TagFilters = append(cfg.TagFilters, tf)
			}
		}
	}
	if expr, ok := config["expression"].(map[string]interface{}); ok {
		cfg.Expression = &ExpressionOp{}
		if v, ok := expr["field"].(string); ok {
			cfg.Expression.Field = v
		}
		if v, ok := expr["op"].(string); ok {
			cfg.Expression.Op = v
		} else {
			cfg.Expression.Op = "=="
		}
		if v, ok := expr["value"].(string); ok {
			cfg.Expression.Value = v
		}
	}
	return cfg
}

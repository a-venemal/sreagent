package eventdrop

import (
	"regexp"
	"strings"

	"github.com/sreagent/sreagent/internal/engine/pipeline"
)

func init() {
	pipeline.RegisterProcessor("event_drop", newEventDropProcessor)
}

// EventDropConfig configures the event_drop processor.
type EventDropConfig struct {
	TagFilters []TagFilter   `json:"tag_filters,omitempty"`
	Expression *ExpressionOp `json:"expression,omitempty"`
}

type TagFilter struct {
	Key   string `json:"key"`
	Op    string `json:"op"`
	Value string `json:"value"`
}

type ExpressionOp struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value string `json:"value"`
}

type eventDropProcessor struct {
	config EventDropConfig
}

func newEventDropProcessor(config map[string]interface{}) (pipeline.Processor, error) {
	cfg := parseEventDropConfig(config)
	return &eventDropProcessor{config: cfg}, nil
}

func (p *eventDropProcessor) Process(wfCtx *pipeline.WorkflowContext) (*pipeline.WorkflowContext, string, error) {
	if wfCtx.Event == nil {
		return wfCtx, "no event, skip", nil
	}

	if p.shouldDrop(wfCtx) {
		wfCtx.Event = nil
		return wfCtx, "event dropped by event_drop processor", nil
	}

	return wfCtx, "event not matched, continue", nil
}

func (p *eventDropProcessor) shouldDrop(wfCtx *pipeline.WorkflowContext) bool {
	event := wfCtx.Event

	if p.config.Expression != nil {
		var fieldValue string
		switch p.config.Expression.Field {
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
		if matchOp(fieldValue, p.config.Expression.Op, p.config.Expression.Value) {
			return true
		}
	}

	if len(p.config.TagFilters) > 0 {
		labels := event.Labels
		allMatch := true
		for _, tf := range p.config.TagFilters {
			val, ok := labels[tf.Key]
			if !ok {
				if tf.Op == "==" || tf.Op == "=~" || tf.Op == "in" {
					allMatch = false
					break
				}
				continue
			}
			if !matchOp(val, tf.Op, tf.Value) {
				allMatch = false
				break
			}
		}
		if allMatch {
			return true
		}
	}

	return false
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
		for _, v := range splitCSV(filterValue) {
			if fieldValue == v {
				return true
			}
		}
		return false
	case "not_in":
		for _, v := range splitCSV(filterValue) {
			if fieldValue == v {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func splitCSV(s string) []string {
	return strings.Split(s, ",")
}

func parseEventDropConfig(config map[string]interface{}) EventDropConfig {
	cfg := EventDropConfig{}
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

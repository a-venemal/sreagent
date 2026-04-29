package relabel

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"strings"

	"github.com/sreagent/sreagent/internal/engine/pipeline"
	"github.com/sreagent/sreagent/internal/model"
)

func init() {
	pipeline.RegisterProcessor("relabel", newRelabelProcessor)
}

// RelabelConfig defines a single relabeling rule (Prometheus-style).
type RelabelConfig struct {
	SourceLabels []string `json:"source_labels"`
	Regex        string   `json:"regex"`
	Action       string   `json:"action"` // replace, keep, drop, hashmod, labelmap, labeldrop
	TargetLabel  string   `json:"target_label"`
	Replacement  string   `json:"replacement"`
	Modulus       uint64   `json:"modulus,omitempty"`
	regex        *regexp.Regexp
}

type relabelProcessor struct {
	configs []RelabelConfig
}

func newRelabelProcessor(config map[string]interface{}) (pipeline.Processor, error) {
	configs := parseRelabelConfigs(config)
	// Pre-compile regexes
	for i := range configs {
		if configs[i].Regex != "" {
			re, err := regexp.Compile(configs[i].Regex)
			if err != nil {
				return nil, fmt.Errorf("invalid regex %q: %w", configs[i].Regex, err)
			}
			configs[i].regex = re
		}
	}
	return &relabelProcessor{configs: configs}, nil
}

func (p *relabelProcessor) Process(wfCtx *pipeline.WorkflowContext) (*pipeline.WorkflowContext, string, error) {
	if wfCtx.Event == nil {
		return wfCtx, "no event to relabel", nil
	}

	labels := wfCtx.Event.Labels
	if labels == nil {
		labels = model.JSONLabels{}
	}

	changed := 0
	for _, cfg := range p.configs {
		switch cfg.Action {
		case "replace":
			if p.applyReplace(labels, cfg) {
				changed++
			}
		case "labeldrop":
			for key := range labels {
				if cfg.regex != nil && cfg.regex.MatchString(key) {
					delete(labels, key)
					changed++
				}
			}
		case "labelmap":
			for key, val := range labels {
				if cfg.regex != nil && cfg.regex.MatchString(key) {
					newKey := cfg.regex.ReplaceAllString(key, cfg.TargetLabel)
					if newKey != key {
						labels[newKey] = val
						delete(labels, key)
						changed++
					}
				}
			}
		case "keep":
			// keep only labels matching the regex, drop others
			for key := range labels {
				if !matchSourceLabels(labels, cfg) {
					delete(labels, key)
					changed++
				}
			}
		case "drop":
			if matchSourceLabels(labels, cfg) {
				wfCtx.Event = nil
				return wfCtx, "event dropped by relabel drop action", nil
			}
		case "hashmod":
			val := joinSourceLabels(labels, cfg.SourceLabels)
			hash := md5.Sum([]byte(val))
			mod := uint64(hash[0])<<24 | uint64(hash[1])<<16 | uint64(hash[2])<<8 | uint64(hash[3])
			if cfg.Modulus > 0 {
				labels[cfg.TargetLabel] = fmt.Sprintf("%d", mod%cfg.Modulus)
				changed++
			}
		}
	}

	wfCtx.Event.Labels = labels
	msg := fmt.Sprintf("relabel: %d label changes applied", changed)
	return wfCtx, msg, nil
}

func (p *relabelProcessor) applyReplace(labels model.JSONLabels, cfg RelabelConfig) bool {
	val := joinSourceLabels(labels, cfg.SourceLabels)
	var newVal string
	if cfg.regex != nil {
		newVal = cfg.regex.ReplaceAllString(val, cfg.Replacement)
	} else {
		newVal = val
	}

	if cfg.TargetLabel == "" {
		return false
	}

	if old, exists := labels[cfg.TargetLabel]; !exists || old != newVal {
		labels[cfg.TargetLabel] = newVal
		return true
	}
	return false
}

func matchSourceLabels(labels model.JSONLabels, cfg RelabelConfig) bool {
	val := joinSourceLabels(labels, cfg.SourceLabels)
	if cfg.regex != nil {
		return cfg.regex.MatchString(val)
	}
	return val != ""
}

func joinSourceLabels(labels model.JSONLabels, sourceLabels []string) string {
	parts := make([]string, len(sourceLabels))
	for i, key := range sourceLabels {
		parts[i] = labels[key]
	}
	return strings.Join(parts, ";")
}

func parseRelabelConfigs(config map[string]interface{}) []RelabelConfig {
	var configs []RelabelConfig
	arr, ok := config["configs"].([]interface{})
	if !ok {
		// Single config shorthand
		arr = []interface{}{config}
	}
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		rc := RelabelConfig{}
		if v, ok := m["action"].(string); ok {
			rc.Action = v
		} else {
			rc.Action = "replace"
		}
		if v, ok := m["target_label"].(string); ok {
			rc.TargetLabel = v
		}
		if v, ok := m["regex"].(string); ok {
			rc.Regex = v
		}
		if v, ok := m["replacement"].(string); ok {
			rc.Replacement = v
		} else {
			rc.Replacement = "${1}"
		}
		if v, ok := m["modulus"].(float64); ok {
			rc.Modulus = uint64(v)
		}
		if sl, ok := m["source_labels"].([]interface{}); ok {
			for _, s := range sl {
				if sv, ok := s.(string); ok {
					rc.SourceLabels = append(rc.SourceLabels, sv)
				}
			}
		}
		configs = append(configs, rc)
	}
	return configs
}

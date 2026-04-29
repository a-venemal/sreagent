package callback

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sreagent/sreagent/internal/engine/pipeline"
)

func init() {
	pipeline.RegisterProcessor("callback", newCallbackProcessor)
}

// CallbackConfig configures the callback processor.
type CallbackConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"` // GET or POST, default POST
	Headers map[string]string `json:"headers,omitempty"`
	Timeout int               `json:"timeout,omitempty"` // seconds, default 10
}

type callbackProcessor struct {
	config CallbackConfig
}

func newCallbackProcessor(config map[string]interface{}) (pipeline.Processor, error) {
	cfg := CallbackConfig{Method: "POST", Timeout: 10}
	if v, ok := config["url"].(string); ok {
		cfg.URL = v
	}
	if cfg.URL == "" {
		return nil, fmt.Errorf("callback processor requires a URL")
	}
	if v, ok := config["method"].(string); ok {
		cfg.Method = v
	}
	if v, ok := config["headers"].(map[string]interface{}); ok {
		cfg.Headers = make(map[string]string)
		for k, val := range v {
			if sv, ok := val.(string); ok {
				cfg.Headers[k] = sv
			}
		}
	}
	if v, ok := config["timeout"].(float64); ok {
		cfg.Timeout = int(v)
	}
	return &callbackProcessor{config: cfg}, nil
}

func (p *callbackProcessor) Process(wfCtx *pipeline.WorkflowContext) (*pipeline.WorkflowContext, string, error) {
	if wfCtx.Event == nil {
		return wfCtx, "no event to callback", nil
	}

	// Serialize event as JSON body
	body, err := json.Marshal(wfCtx.Event)
	if err != nil {
		return wfCtx, "", fmt.Errorf("failed to marshal event: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, p.config.Method, p.config.URL, bytes.NewReader(body))
	if err != nil {
		return wfCtx, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range p.config.Headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return wfCtx, "", fmt.Errorf("callback request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	if resp.StatusCode >= 400 {
		return wfCtx, "", fmt.Errorf("callback returned status %d: %s", resp.StatusCode, string(respBody))
	}

	msg := fmt.Sprintf("callback to %s returned %d", p.config.URL, resp.StatusCode)
	return wfCtx, msg, nil
}

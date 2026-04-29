package aisummary

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sreagent/sreagent/internal/engine/pipeline"
)

func init() {
	pipeline.RegisterProcessor("ai_summary", newAISummaryProcessor)
}

// AISummaryConfig configures the ai_summary processor.
type AISummaryConfig struct {
	APIURL  string `json:"api_url"`  // OpenAI-compatible API endpoint
	APIKey  string `json:"api_key"`  // API key
	Model   string `json:"model"`    // model name
	Prompt  string `json:"prompt"`   // custom prompt template (optional)
	Timeout int    `json:"timeout"`  // seconds, default 30
}

type aiSummaryProcessor struct {
	config AISummaryConfig
}

func newAISummaryProcessor(config map[string]interface{}) (pipeline.Processor, error) {
	cfg := AISummaryConfig{Timeout: 30, Model: "gpt-4o-mini"}
	if v, ok := config["api_url"].(string); ok {
		cfg.APIURL = v
	}
	if cfg.APIURL == "" {
		return nil, fmt.Errorf("ai_summary processor requires api_url")
	}
	if v, ok := config["api_key"].(string); ok {
		cfg.APIKey = v
	}
	if v, ok := config["model"].(string); ok {
		cfg.Model = v
	}
	if v, ok := config["prompt"].(string); ok {
		cfg.Prompt = v
	}
	if v, ok := config["timeout"].(float64); ok {
		cfg.Timeout = int(v)
	}
	return &aiSummaryProcessor{config: cfg}, nil
}

func (p *aiSummaryProcessor) Process(wfCtx *pipeline.WorkflowContext) (*pipeline.WorkflowContext, string, error) {
	if wfCtx.Event == nil {
		return wfCtx, "no event for AI summary", nil
	}

	event := wfCtx.Event
	prompt := p.buildPrompt(event)

	// Call OpenAI-compatible API
	summary, err := p.callLLM(prompt)
	if err != nil {
		return wfCtx, "", fmt.Errorf("LLM call failed: %w", err)
	}

	// Write summary into annotations
	if event.Annotations == nil {
		event.Annotations = make(map[string]string)
	}
	event.Annotations["ai_summary"] = summary
	wfCtx.Event = event

	return wfCtx, "AI summary generated", nil
}

func (p *aiSummaryProcessor) buildPrompt(event interface{}) string {
	if p.config.Prompt != "" {
		// Simple template replacement
		prompt := p.config.Prompt
		eventJSON, _ := json.MarshalIndent(event, "", "  ")
		prompt = strings.ReplaceAll(prompt, "{{event}}", string(eventJSON))
		return prompt
	}

	return fmt.Sprintf(`Analyze the following alert event and provide a concise root cause analysis and suggested actions.

Alert Event:
%s

Please provide:
1. Root cause analysis (2-3 sentences)
2. Suggested investigation steps
3. Recommended actions

Keep the response concise and actionable.`, eventJSON(event))
}

func eventJSON(event interface{}) string {
	b, _ := json.MarshalIndent(event, "", "  ")
	return string(b)
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *aiSummaryProcessor) callLLM(userPrompt string) (string, error) {
	reqBody := chatRequest{
		Model: p.config.Model,
		Messages: []chatMessage{
			{Role: "system", Content: "You are an expert SRE assistant. Analyze alerts concisely and provide actionable insights."},
			{Role: "user", Content: userPrompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.APIURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 32*1024))

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse LLM response: %w", err)
	}
	if chatResp.Error != nil {
		return "", fmt.Errorf("LLM error: %s", chatResp.Error.Message)
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("LLM returned no choices")
	}

	return chatResp.Choices[0].Message.Content, nil
}

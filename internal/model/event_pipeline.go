package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// EventPipeline represents a configurable alert processing pipeline.
// Users can define a DAG of processors (if/relabel/event_drop/callback/ai_summary)
// to customize how alerts are processed before notification.
type EventPipeline struct {
	BaseModel
	Name        string `json:"name" gorm:"size:256;not null"`
	Description string `json:"description" gorm:"size:1024"`
	Disabled    bool   `json:"disabled" gorm:"default:false"`
	// Pre-filters: only process events matching these label conditions
	FilterEnable bool          `json:"filter_enable" gorm:"default:false"`
	LabelFilters LabelFilters  `json:"label_filters" gorm:"type:json"`
	// DAG configuration
	Nodes       PipelineNodes  `json:"nodes" gorm:"type:json"`
	Connections Connections     `json:"connections" gorm:"type:json"`
	CreatedBy   uint           `json:"created_by" gorm:"index"`
	UpdatedBy   uint           `json:"updated_by"`
}

func (EventPipeline) TableName() string {
	return "event_pipelines"
}

// PipelineNode represents a single node (processor) in the pipeline DAG.
type PipelineNode struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Type           string                 `json:"type"` // if, relabel, event_drop, callback, ai_summary
	Config         map[string]interface{} `json:"config"`
	Disabled       bool                   `json:"disabled,omitempty"`
	ContinueOnFail bool                   `json:"continue_on_fail,omitempty"`
	RetryOnFail    bool                   `json:"retry_on_fail,omitempty"`
	MaxRetries     int                    `json:"max_retries,omitempty"`
	Position       *NodePosition          `json:"position,omitempty"`
}

type NodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Connections maps source node ID -> output index -> list of target node IDs.
// For linear nodes, output index is always 0.
// For branch nodes (if): 0 = true branch, 1 = false branch.
type Connections map[string]map[int][]string

// LabelFilter defines a label matching condition for pipeline pre-filtering.
type LabelFilter struct {
	Key   string `json:"key"`
	Op    string `json:"op"` // ==, !=, =~, !~, in, not_in
	Value string `json:"value"`
}

// PipelineNodes is a JSON-serializable slice of PipelineNode.
type PipelineNodes []PipelineNode

func (j PipelineNodes) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	return json.Marshal(j)
}

func (j *PipelineNodes) Scan(src interface{}) error {
	if src == nil {
		*j = nil
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("unsupported type for PipelineNodes: %T", src)
	}
	return json.Unmarshal(data, j)
}

// LabelFilters is a JSON-serializable slice of LabelFilter.
type LabelFilters []LabelFilter

func (j LabelFilters) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	return json.Marshal(j)
}

func (j *LabelFilters) Scan(src interface{}) error {
	if src == nil {
		*j = nil
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("unsupported type for LabelFilters: %T", src)
	}
	return json.Unmarshal(data, j)
}

func (c Connections) Value() (driver.Value, error) {
	if c == nil {
		return "{}", nil
	}
	return json.Marshal(c)
}

func (c *Connections) Scan(src interface{}) error {
	if src == nil {
		*c = nil
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("unsupported type for Connections: %T", src)
	}
	return json.Unmarshal(data, c)
}

// PipelineExecution records the result of a single pipeline execution.
type PipelineExecution struct {
	ID           string    `json:"id" gorm:"size:36;primaryKey"`
	PipelineID   uint      `json:"pipeline_id" gorm:"index"`
	EventID      uint      `json:"event_id" gorm:"index"`
	Status       string    `json:"status"` // success, failed, terminated
	NodeResults  string    `json:"node_results" gorm:"type:json"`
	ErrorMessage string    `json:"error_message" gorm:"type:text"`
	DurationMs   int64     `json:"duration_ms"`
	StartedAt    time.Time `json:"started_at"`
	FinishedAt   time.Time `json:"finished_at"`
}

func (PipelineExecution) TableName() string {
	return "pipeline_executions"
}

// NodeExecutionResult records the result of a single node execution.
type NodeExecutionResult struct {
	NodeID     string `json:"node_id"`
	NodeName   string `json:"node_name"`
	NodeType   string `json:"node_type"`
	Status     string `json:"status"` // success, failed, skipped, terminated
	Message    string `json:"message,omitempty"`
	Error      string `json:"error,omitempty"`
	BranchIdx  int    `json:"branch_idx,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}

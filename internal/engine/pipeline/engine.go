package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/sreagent/sreagent/internal/model"
)

// WorkflowResult is the outcome of a full pipeline execution.
type WorkflowResult struct {
	Event       *model.AlertEvent
	Status      string // success, failed, terminated
	Message     string
	NodeResults []model.NodeExecutionResult
	ErrorNode   string
	Terminated  bool
}

// Engine executes event pipelines using DAG topological sort (Kahn's algorithm).
type Engine struct {
	logger   *zap.Logger
	execRepo ExecutionRecorder
}

// ExecutionRecorder abstracts the persistence of pipeline execution records.
type ExecutionRecorder interface {
	Create(ctx context.Context, exec *model.PipelineExecution) error
}

// NewEngine creates a new pipeline execution engine.
func NewEngine(logger *zap.Logger, execRepo ExecutionRecorder) *Engine {
	return &Engine{logger: logger, execRepo: execRepo}
}

// edge represents a directed edge in the DAG.
type edge struct {
	target    string
	outputIdx int
}

// Execute runs a single pipeline against an event.
func (e *Engine) Execute(ctx context.Context, ppl *model.EventPipeline, event *model.AlertEvent) (*WorkflowResult, error) {
	startedAt := time.Now()
	result := &WorkflowResult{
		Event:  event,
		Status: "success",
	}

	// Pre-filter check
	if ppl.FilterEnable && !matchFilters(event, ppl.LabelFilters) {
		result.Message = "event did not match pipeline filters"
		return result, nil
	}

	nodes := ppl.Nodes
	conns := ppl.Connections

	if len(nodes) == 0 {
		result.Message = "pipeline has no nodes"
		return result, nil
	}

	// Build node map
	nodeMap := make(map[string]*model.PipelineNode, len(nodes))
	for i := range nodes {
		nodeMap[nodes[i].ID] = &nodes[i]
	}

	// Build adjacency list and in-degree map
	adjacency := make(map[string][]edge)
	inDegree := make(map[string]int)

	for _, n := range nodes {
		inDegree[n.ID] = 0
	}

	for srcID, outputs := range conns {
		for outIdx, targets := range outputs {
			for _, tgtID := range targets {
				adjacency[srcID] = append(adjacency[srcID], edge{target: tgtID, outputIdx: outIdx})
				inDegree[tgtID]++
			}
		}
	}

	// Kahn's algorithm: seed queue with in-degree 0 nodes
	queue := make([]string, 0)
	for _, n := range nodes {
		if inDegree[n.ID] == 0 {
			queue = append(queue, n.ID)
		}
	}

	wfCtx := &WorkflowContext{
		Event:    event,
		Vars:     make(map[string]interface{}),
		Metadata: map[string]string{"pipeline_id": fmt.Sprintf("%d", ppl.ID)},
	}

	branchResults := make(map[string]int)
	executed := 0

	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]

		node, ok := nodeMap[nodeID]
		if !ok {
			continue
		}

		// Skip disabled nodes
		if node.Disabled {
			result.NodeResults = append(result.NodeResults, model.NodeExecutionResult{
				NodeID: nodeID, NodeName: node.Name, NodeType: node.Type,
				Status: "skipped", Message: "node is disabled",
			})
			executed++
			enqueueSuccessors(nodeID, adjacency, inDegree, &queue, 0)
			continue
		}

		// Execute node
		nodeResult := e.executeNode(ctx, node, wfCtx)
		result.NodeResults = append(result.NodeResults, nodeResult)
		executed++

		if nodeResult.Status == "terminated" {
			result.Status = "terminated"
			result.Terminated = true
			result.Message = fmt.Sprintf("pipeline terminated at node %s", node.Name)
			result.ErrorNode = nodeID
			break
		}

		if nodeResult.Status == "failed" {
			result.Status = "failed"
			result.Message = fmt.Sprintf("pipeline failed at node %s: %s", node.Name, nodeResult.Error)
			result.ErrorNode = nodeID
			break
		}

		if nodeResult.BranchIdx > 0 {
			branchResults[nodeID] = nodeResult.BranchIdx
		}

		branchIdx := branchResults[nodeID]
		enqueueSuccessors(nodeID, adjacency, inDegree, &queue, branchIdx)
	}

	// Detect circular dependency
	if executed < len(nodes) && result.Status == "success" {
		result.Status = "failed"
		result.Message = "circular dependency detected in pipeline"
	}

	// Save execution record
	finishedAt := time.Now()
	execRecord := &model.PipelineExecution{
		ID:         uuid.New().String(),
		PipelineID: ppl.ID,
		Status:     result.Status,
		DurationMs: finishedAt.Sub(startedAt).Milliseconds(),
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
	}
	if result.Event != nil {
		execRecord.EventID = result.Event.ID
	}
	if nodeResultsJSON, err := json.Marshal(result.NodeResults); err == nil {
		execRecord.NodeResults = string(nodeResultsJSON)
	}
	if result.Message != "" {
		execRecord.ErrorMessage = result.Message
	}

	if err := e.execRepo.Create(ctx, execRecord); err != nil {
		e.logger.Warn("failed to save pipeline execution record", zap.Error(err))
	}

	return result, nil
}

// ExecuteMatching finds and executes all matching pipelines for an event.
func (e *Engine) ExecuteMatching(ctx context.Context, pipelines []*model.EventPipeline, event *model.AlertEvent) (*WorkflowResult, error) {
	for _, p := range pipelines {
		if p.Disabled {
			continue
		}
		result, err := e.Execute(ctx, p, event)
		if err != nil {
			e.logger.Error("pipeline execution error",
				zap.Uint("pipeline_id", p.ID),
				zap.Error(err),
			)
			continue
		}
		if result != nil && result.Terminated {
			return result, nil
		}
	}
	return nil, nil
}

func (e *Engine) executeNode(ctx context.Context, node *model.PipelineNode, wfCtx *WorkflowContext) model.NodeExecutionResult {
	startedAt := time.Now()
	result := model.NodeExecutionResult{
		NodeID:   node.ID,
		NodeName: node.Name,
		NodeType: node.Type,
	}

	proc, err := GetProcessor(node.Type, node.Config)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to create processor: %v", err)
		result.DurationMs = time.Since(startedAt).Milliseconds()
		return result
	}

	maxRetries := 0
	if node.RetryOnFail {
		maxRetries = node.MaxRetries
		if maxRetries <= 0 {
			maxRetries = 1
		}
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Second)
		}

		// Branch processor
		if bp, ok := proc.(BranchProcessor); ok {
			output, err := bp.ProcessWithBranch(wfCtx)
			if err != nil {
				lastErr = err
				continue
			}
			if output.Terminate {
				result.Status = "terminated"
				result.Message = output.Message
				result.DurationMs = time.Since(startedAt).Milliseconds()
				return result
			}
			if output.WfCtx != nil {
				*wfCtx = *output.WfCtx
			}
			result.Status = "success"
			result.Message = output.Message
			result.BranchIdx = output.BranchIndex
			result.DurationMs = time.Since(startedAt).Milliseconds()
			return result
		}

		// Regular processor
		newCtx, msg, err := proc.Process(wfCtx)
		if err != nil {
			lastErr = err
			continue
		}
		if newCtx != nil && newCtx.Event == nil {
			result.Status = "terminated"
			result.Message = msg
			result.DurationMs = time.Since(startedAt).Milliseconds()
			return result
		}
		if newCtx != nil {
			*wfCtx = *newCtx
		}
		result.Status = "success"
		result.Message = msg
		result.DurationMs = time.Since(startedAt).Milliseconds()
		return result
	}

	if node.ContinueOnFail {
		result.Status = "failed"
		result.Error = fmt.Sprintf("processor failed after %d attempts: %v", maxRetries+1, lastErr)
		result.DurationMs = time.Since(startedAt).Milliseconds()
		return result
	}

	result.Status = "failed"
	result.Error = fmt.Sprintf("processor failed: %v", lastErr)
	result.DurationMs = time.Since(startedAt).Milliseconds()
	return result
}

func enqueueSuccessors(nodeID string, adjacency map[string][]edge, inDegree map[string]int, queue *[]string, branchIdx int) {
	for _, e := range adjacency[nodeID] {
		if e.outputIdx != branchIdx {
			continue
		}
		inDegree[e.target]--
		if inDegree[e.target] == 0 {
			*queue = append(*queue, e.target)
		}
	}
}

func matchFilters(event *model.AlertEvent, filters model.LabelFilters) bool {
	if len(filters) == 0 {
		return true
	}
	labels := event.Labels
	for _, f := range filters {
		val, ok := labels[f.Key]
		switch f.Op {
		case "==":
			if !ok || val != f.Value {
				return false
			}
		case "!=":
			if ok && val == f.Value {
				return false
			}
		case "=~":
			if !ok || !matchRegexp(f.Value, val) {
				return false
			}
		case "!~":
			if ok && matchRegexp(f.Value, val) {
				return false
			}
		}
	}
	return true
}

func matchRegexp(pattern, s string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

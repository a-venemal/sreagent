package pipeline

import (
	"fmt"
	"sync"

	"github.com/sreagent/sreagent/internal/model"
)

// Processor is the interface that all pipeline processors must implement.
type Processor interface {
	// Process executes the processor logic and returns the (possibly modified) context.
	// The string return value is a human-readable message for execution logs.
	Process(wfCtx *WorkflowContext) (*WorkflowContext, string, error)
}

// BranchProcessor extends Processor for nodes that produce multiple output branches.
// Used by conditional nodes (if, switch) to route events to different downstream paths.
type BranchProcessor interface {
	Processor
	ProcessWithBranch(wfCtx *WorkflowContext) (*NodeOutput, error)
}

// NewProcessorFn is a factory function that creates a Processor from config.
type NewProcessorFn func(config map[string]interface{}) (Processor, error)

// WorkflowContext is the runtime context passed through all pipeline nodes.
type WorkflowContext struct {
	Event    *model.AlertEvent
	Vars     map[string]interface{}
	Metadata map[string]string
}

// NodeOutput is the result of a node execution.
type NodeOutput struct {
	WfCtx       *WorkflowContext
	Message     string
	Terminate   bool // if true, the pipeline stops
	BranchIndex int  // which output branch to follow (for branch processors)
}

// processorRegistry is the global registry of processor factories.
var (
	processorRegistry = map[string]NewProcessorFn{}
	registryMu        sync.RWMutex
)

// RegisterProcessor registers a processor factory by type name.
// Typically called from init() functions in processor packages.
func RegisterProcessor(typ string, fn NewProcessorFn) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := processorRegistry[typ]; exists {
		return // no-op on duplicate
	}
	processorRegistry[typ] = fn
}

// GetProcessor creates a new processor instance by type name and config.
func GetProcessor(typ string, config map[string]interface{}) (Processor, error) {
	registryMu.RLock()
	fn, ok := processorRegistry[typ]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown processor type: %s", typ)
	}
	return fn(config)
}

// RegisteredTypes returns a list of all registered processor type names.
func RegisteredTypes() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	types := make([]string, 0, len(processorRegistry))
	for t := range processorRegistry {
		types = append(types, t)
	}
	return types
}

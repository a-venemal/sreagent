package processor

import (
	"github.com/sreagent/sreagent/internal/engine/pipeline"

	// Register all processors via init() functions.
	_ "github.com/sreagent/sreagent/internal/engine/pipeline/processor/aisummary"
	_ "github.com/sreagent/sreagent/internal/engine/pipeline/processor/callback"
	_ "github.com/sreagent/sreagent/internal/engine/pipeline/processor/eventdrop"
	_ "github.com/sreagent/sreagent/internal/engine/pipeline/processor/logic"
	_ "github.com/sreagent/sreagent/internal/engine/pipeline/processor/relabel"
)

// Init triggers all processor registrations via blank imports.
func Init() {
	// All processors are registered in init() functions.
	// This function exists as an explicit initialization hook.
	_ = pipeline.RegisteredTypes()
}

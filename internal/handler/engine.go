package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/sreagent/sreagent/internal/engine"
)

// EngineHandler handles alert engine status API requests.
type EngineHandler struct {
	evaluator *engine.Evaluator
}

// NewEngineHandler creates a new EngineHandler.
func NewEngineHandler(evaluator *engine.Evaluator) *EngineHandler {
	return &EngineHandler{
		evaluator: evaluator,
	}
}

// GetStatus returns the status of the alert evaluation engine.
func (h *EngineHandler) GetStatus(c *gin.Context) {
	status := h.evaluator.GetStatus()
	Success(c, status)
}

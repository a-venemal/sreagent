package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/sreagent/sreagent/internal/model"
	"github.com/sreagent/sreagent/internal/service"
)

type EventPipelineHandler struct {
	svc *service.EventPipelineService
}

func NewEventPipelineHandler(svc *service.EventPipelineService) *EventPipelineHandler {
	return &EventPipelineHandler{svc: svc}
}

type CreatePipelineRequest struct {
	Name         string            `json:"name" binding:"required"`
	Description  string            `json:"description"`
	Disabled     bool              `json:"disabled"`
	FilterEnable bool              `json:"filter_enable"`
	LabelFilters model.LabelFilters `json:"label_filters"`
	Nodes        model.PipelineNodes `json:"nodes"`
	Connections  model.Connections  `json:"connections"`
}

func (h *EventPipelineHandler) Create(c *gin.Context) {
	var req CreatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	userID := GetCurrentUserID(c)

	p := &model.EventPipeline{
		Name:         req.Name,
		Description:  req.Description,
		Disabled:     req.Disabled,
		FilterEnable: req.FilterEnable,
		LabelFilters: req.LabelFilters,
		Nodes:        req.Nodes,
		Connections:  req.Connections,
		CreatedBy:    userID,
		UpdatedBy:    userID,
	}

	if err := h.svc.Create(c.Request.Context(), p); err != nil {
		Error(c, err)
		return
	}

	Success(c, p)
}

func (h *EventPipelineHandler) Get(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	p, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, p)
}

func (h *EventPipelineHandler) List(c *gin.Context) {
	pq := GetPageQuery(c)
	search := c.Query("search")

	list, total, err := h.svc.List(c.Request.Context(), search, pq.Page, pq.PageSize)
	if err != nil {
		Error(c, err)
		return
	}

	SuccessPage(c, list, total, pq.Page, pq.PageSize)
}

func (h *EventPipelineHandler) Update(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	var req CreatePipelineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	userID := GetCurrentUserID(c)

	p := &model.EventPipeline{
		Name:         req.Name,
		Description:  req.Description,
		Disabled:     req.Disabled,
		FilterEnable: req.FilterEnable,
		LabelFilters: req.LabelFilters,
		Nodes:        req.Nodes,
		Connections:  req.Connections,
		UpdatedBy:    userID,
	}
	p.ID = id

	if err := h.svc.Update(c.Request.Context(), p); err != nil {
		Error(c, err)
		return
	}

	Success(c, p)
}

func (h *EventPipelineHandler) Delete(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

type TryRunRequest struct {
	PipelineID uint               `json:"pipeline_id" binding:"required"`
	Event      model.AlertEvent   `json:"event"`
}

func (h *EventPipelineHandler) TryRun(c *gin.Context) {
	var req TryRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithMessage(c, 10001, err.Error())
		return
	}

	result, err := h.svc.TryRun(c.Request.Context(), req.PipelineID, &req.Event)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, result)
}

func (h *EventPipelineHandler) ListExecutions(c *gin.Context) {
	id, err := GetIDParam(c, "id")
	if err != nil {
		Error(c, err)
		return
	}

	pq := GetPageQuery(c)

	list, total, err := h.svc.ListExecutions(c.Request.Context(), id, pq.Page, pq.PageSize)
	if err != nil {
		Error(c, err)
		return
	}

	SuccessPage(c, list, total, pq.Page, pq.PageSize)
}

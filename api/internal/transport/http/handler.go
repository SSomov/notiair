package http

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"

	"notiair/internal/routing"
	"notiair/internal/templates"
	"notiair/internal/workflow"
)

type NotificationService interface {
	Dispatch(ctx context.Context, workflowID, templateID string, variables map[string]string, payload map[string]any) error
}

type TemplateRepository interface {
	Save(ctx context.Context, tpl templates.Template) (templates.Template, error)
	List(ctx context.Context) ([]templates.Template, error)
}

type WorkflowRepository interface {
	Save(ctx context.Context, wf workflow.Workflow) (workflow.Workflow, error)
	List(ctx context.Context) ([]workflow.Workflow, error)
}

type QueueInspector interface {
	ListPending(ctx context.Context) ([]routing.Task, error)
}

type Handler struct {
	notifications NotificationService
	templates     TemplateRepository
	workflows     WorkflowRepository
	queue         QueueInspector
}

func NewHandler(notificationSvc NotificationService, tplRepo TemplateRepository, wfRepo WorkflowRepository, queueInspector QueueInspector) *Handler {
	return &Handler{notifications: notificationSvc, templates: tplRepo, workflows: wfRepo, queue: queueInspector}
}

func (h *Handler) Register(router fiber.Router) {
	router.Post("/notifications/dispatch", h.dispatchNotification)
	router.Get("/templates", h.listTemplates)
	router.Post("/templates", h.saveTemplate)
	router.Get("/workflows", h.listWorkflows)
	router.Post("/workflows", h.saveWorkflow)
	router.Get("/queues/pending", h.listQueue)
}

type dispatchRequest struct {
	WorkflowID string            `json:"workflowId"`
	TemplateID string            `json:"templateId"`
	Variables  map[string]string `json:"variables"`
	Payload    map[string]any    `json:"payload"`
}

func (h *Handler) dispatchNotification(c *fiber.Ctx) error {
	var req dispatchRequest
	if err := c.BodyParser(&req); err != nil {
		return h.respondErr(c, fiber.StatusBadRequest, err)
	}

	if req.WorkflowID == "" || req.TemplateID == "" {
		return h.respondErr(c, fiber.StatusBadRequest, ErrValidation)
	}

	ctx := h.contextFromFiber(c)

	if err := h.notifications.Dispatch(ctx, req.WorkflowID, req.TemplateID, req.Variables, req.Payload); err != nil {
		return h.respondErr(c, fiber.StatusInternalServerError, err)
	}

	return c.SendStatus(fiber.StatusAccepted)
}

type templateRequest struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Body        string            `json:"body"`
	Variables   map[string]string `json:"variables"`
}

func (h *Handler) saveTemplate(c *fiber.Ctx) error {
	var req templateRequest
	if err := c.BodyParser(&req); err != nil {
		return h.respondErr(c, fiber.StatusBadRequest, err)
	}

	tpl := templates.Template{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Body:        req.Body,
		Variables:   req.Variables,
	}

	saved, err := h.templates.Save(h.contextFromFiber(c), tpl)
	if err != nil {
		return h.respondErr(c, fiber.StatusInternalServerError, err)
	}

	return h.respondJSON(c, fiber.StatusCreated, saved)
}

func (h *Handler) listTemplates(c *fiber.Ctx) error {
	tpls, err := h.templates.List(h.contextFromFiber(c))
	if err != nil {
		return h.respondErr(c, fiber.StatusInternalServerError, err)
	}

	return h.respondJSON(c, fiber.StatusOK, tpls)
}

type workflowRequest struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Nodes       []workflow.Node   `json:"nodes"`
	Edges       []workflow.Edge   `json:"edges"`
	Filters     map[string]string `json:"filters"`
}

func (h *Handler) saveWorkflow(c *fiber.Ctx) error {
	var req workflowRequest
	if err := c.BodyParser(&req); err != nil {
		return h.respondErr(c, fiber.StatusBadRequest, err)
	}

	wf := workflow.Workflow{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Nodes:       req.Nodes,
		Edges:       req.Edges,
		Filters:     req.Filters,
	}

	saved, err := h.workflows.Save(h.contextFromFiber(c), wf)
	if err != nil {
		return h.respondErr(c, fiber.StatusInternalServerError, err)
	}

	return h.respondJSON(c, fiber.StatusCreated, saved)
}

func (h *Handler) listWorkflows(c *fiber.Ctx) error {
	wfs, err := h.workflows.List(h.contextFromFiber(c))
	if err != nil {
		return h.respondErr(c, fiber.StatusInternalServerError, err)
	}

	return h.respondJSON(c, fiber.StatusOK, wfs)
}

func (h *Handler) listQueue(c *fiber.Ctx) error {
	items, err := h.queue.ListPending(h.contextFromFiber(c))
	if err != nil {
		return h.respondErr(c, fiber.StatusInternalServerError, err)
	}

	return h.respondJSON(c, fiber.StatusOK, items)
}

func (h *Handler) respondJSON(c *fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(data)
}

var ErrValidation = errors.New("validation error")

func (h *Handler) respondErr(c *fiber.Ctx, status int, err error) error {
	return h.respondJSON(c, status, map[string]string{"error": err.Error()})
}

func (h *Handler) contextFromFiber(c *fiber.Ctx) context.Context {
	if ctx := c.UserContext(); ctx != nil {
		return ctx
	}
	return context.Background()
}

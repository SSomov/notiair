package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"notiair/internal/routing"
	"notiair/internal/templates"
	"notiair/internal/workflow"
	"notiair/services"
)

type NotificationService interface {
	Dispatch(ctx context.Context, input services.DispatchInput) error
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

type API struct {
	notifications NotificationService
	templates     TemplateRepository
	workflows     WorkflowRepository
	queue         QueueInspector
}

func NewAPI(notificationSvc NotificationService, tplRepo TemplateRepository, wfRepo WorkflowRepository, queueInspector QueueInspector) *API {
	return &API{notifications: notificationSvc, templates: tplRepo, workflows: wfRepo, queue: queueInspector}
}

type dispatchRequest struct {
	WorkflowID string            `json:"workflowId"`
	TemplateID string            `json:"templateId"`
	Variables  map[string]string `json:"variables"`
	Payload    map[string]any    `json:"payload"`
}

func (a *API) DispatchNotification(c *fiber.Ctx) error {
	var req dispatchRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if req.WorkflowID == "" || req.TemplateID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "workflowId and templateId are required")
	}

	if err := a.notifications.Dispatch(c.Context(), services.DispatchInput{
		WorkflowID: req.WorkflowID,
		TemplateID: req.TemplateID,
		Variables:  req.Variables,
		Payload:    req.Payload,
	}); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
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

func (a *API) SaveTemplate(c *fiber.Ctx) error {
	var req templateRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	tpl := templates.Template{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Body:        req.Body,
		Variables:   req.Variables,
	}

	saved, err := a.templates.Save(c.Context(), tpl)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(saved)
}

func (a *API) ListTemplates(c *fiber.Ctx) error {
	tpls, err := a.templates.List(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(tpls)
}

type workflowRequest struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Nodes       []workflow.Node   `json:"nodes"`
	Edges       []workflow.Edge   `json:"edges"`
	Filters     map[string]string `json:"filters"`
}

func (a *API) SaveWorkflow(c *fiber.Ctx) error {
	var req workflowRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	wf := workflow.Workflow{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Nodes:       req.Nodes,
		Edges:       req.Edges,
		Filters:     req.Filters,
	}

	saved, err := a.workflows.Save(c.Context(), wf)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(saved)
}

func (a *API) ListWorkflows(c *fiber.Ctx) error {
	wfs, err := a.workflows.List(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(wfs)
}

func (a *API) ListQueue(c *fiber.Ctx) error {
	items, err := a.queue.ListPending(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(items)
}

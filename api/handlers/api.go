package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"notiair/internal/persistence/channel"
	"notiair/internal/persistence/serviceconfig"
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
	FindByID(ctx context.Context, id string) (workflow.Workflow, error)
	List(ctx context.Context) ([]workflow.Workflow, error)
	Delete(ctx context.Context, id string) error
}

type QueueInspector interface {
	ListPending(ctx context.Context) ([]routing.Task, error)
}

type ServiceConfigRepository interface {
	List(ctx context.Context) ([]serviceconfig.ServiceConfig, error)
	Create(ctx context.Context, input serviceconfig.CreateInput) (serviceconfig.ServiceConfig, error)
	Update(ctx context.Context, id string, input serviceconfig.UpdateInput) (serviceconfig.ServiceConfig, error)
	Delete(ctx context.Context, id string) error
	SetActive(ctx context.Context, id string, active bool) error
}

type ChannelRepository interface {
	ListByConnector(ctx context.Context, connectorID string) ([]channel.Channel, error)
	Create(ctx context.Context, input channel.CreateInput) (channel.Channel, error)
	Update(ctx context.Context, id string, input channel.UpdateInput) (channel.Channel, error)
	Delete(ctx context.Context, id string) error
}

type API struct {
	notifications NotificationService
	templates     TemplateRepository
	workflows     WorkflowRepository
	queue         QueueInspector
	serviceConfig ServiceConfigRepository
	channels      ChannelRepository
}

func NewAPI(notificationSvc NotificationService, tplRepo TemplateRepository, wfRepo WorkflowRepository, queueInspector QueueInspector, serviceConfigRepo ServiceConfigRepository, channelRepo ChannelRepository) *API {
	return &API{
		notifications: notificationSvc,
		templates:     tplRepo,
		workflows:     wfRepo,
		queue:         queueInspector,
		serviceConfig: serviceConfigRepo,
		channels:      channelRepo,
	}
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
	IsActive    bool              `json:"isActive"`
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
		IsActive:    req.IsActive,
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

	if wfs == nil {
		wfs = []workflow.Workflow{}
	}

	return c.JSON(wfs)
}

func (a *API) GetWorkflow(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id is required")
	}

	wf, err := a.workflows.FindByID(c.Context(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "workflow not found")
	}

	return c.JSON(wf)
}

func (a *API) DeleteWorkflow(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id is required")
	}

	if err := a.workflows.Delete(c.Context(), id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (a *API) ListQueue(c *fiber.Ctx) error {
	items, err := a.queue.ListPending(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(items)
}

type telegramTokenRequest struct {
	Name    string `json:"name"`
	Secret  string `json:"secret"`
	Comment string `json:"comment"`
}

type telegramTokenResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Secret  string `json:"secret"`
	Comment string `json:"comment"`
	IsActive bool  `json:"isActive"`
}

func (a *API) ListTelegramTokens(c *fiber.Ctx) error {
	configs, err := a.serviceConfig.List(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	tokens := make([]telegramTokenResponse, 0)
	for _, cfg := range configs {
		if cfg.Type != serviceconfig.TypeTelegram {
			continue
		}
		token, _ := cfg.Settings["token"].(string)
		name, _ := cfg.Settings["name"].(string)
		comment, _ := cfg.Settings["comment"].(string)
		if token != "" {
			tokens = append(tokens, telegramTokenResponse{
				ID:       cfg.ID,
				Name:     name,
				Secret:   token,
				Comment:  comment,
				IsActive: cfg.IsActive,
			})
		}
	}

	return c.JSON(tokens)
}

func (a *API) CreateTelegramToken(c *fiber.Ctx) error {
	var req telegramTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if req.Secret == "" {
		return fiber.NewError(fiber.StatusBadRequest, "secret is required")
	}

	settings := map[string]any{
		"token":   req.Secret,
		"name":    req.Name,
		"comment": req.Comment,
	}

	createInput := serviceconfig.CreateInput{
		Type:      serviceconfig.TypeTelegram,
		IsDefault: false,
		IsActive:  true,
		Settings:  settings,
	}

	created, err := a.serviceConfig.Create(c.Context(), createInput)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	token, _ := created.Settings["token"].(string)
	name, _ := created.Settings["name"].(string)
	comment, _ := created.Settings["comment"].(string)

	return c.Status(fiber.StatusCreated).JSON(telegramTokenResponse{
		ID:       created.ID,
		Name:     name,
		Secret:   token,
		Comment:  comment,
		IsActive: created.IsActive,
	})
}

func (a *API) UpdateTelegramToken(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id is required")
	}

	var req telegramTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if req.Secret == "" {
		return fiber.NewError(fiber.StatusBadRequest, "secret is required")
	}

	settings := map[string]any{
		"token":   req.Secret,
		"name":    req.Name,
		"comment": req.Comment,
	}

	updateInput := serviceconfig.UpdateInput{
		Settings: settings,
	}

	updated, err := a.serviceConfig.Update(c.Context(), id, updateInput)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	token, _ := updated.Settings["token"].(string)
	name, _ := updated.Settings["name"].(string)
	comment, _ := updated.Settings["comment"].(string)

	return c.JSON(telegramTokenResponse{
		ID:       updated.ID,
		Name:     name,
		Secret:   token,
		Comment:  comment,
		IsActive: updated.IsActive,
	})
}

func (a *API) DeleteTelegramToken(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id is required")
	}

	if err := a.serviceConfig.Delete(c.Context(), id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (a *API) ToggleTelegramTokenActive(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id is required")
	}

	var req struct {
		IsActive bool `json:"isActive"`
	}
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if err := a.serviceConfig.SetActive(c.Context(), id, req.IsActive); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	configs, err := a.serviceConfig.List(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	var updatedCfg serviceconfig.ServiceConfig
	for _, cfg := range configs {
		if cfg.ID == id {
			updatedCfg = cfg
			break
		}
	}

	token, _ := updatedCfg.Settings["token"].(string)
	name, _ := updatedCfg.Settings["name"].(string)
	comment, _ := updatedCfg.Settings["comment"].(string)

	return c.JSON(telegramTokenResponse{
		ID:       updatedCfg.ID,
		Name:     name,
		Secret:   token,
		Comment:  comment,
		IsActive: updatedCfg.IsActive,
	})
}

type channelRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Muted       bool   `json:"muted"`
}

type channelResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Muted       bool   `json:"muted"`
}

func (a *API) ListChannels(c *fiber.Ctx) error {
	connectorID := c.Params("connectorId")
	if connectorID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "connectorId is required")
	}

	channels, err := a.channels.ListByConnector(c.Context(), connectorID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	result := make([]channelResponse, len(channels))
	for i, ch := range channels {
		result[i] = channelResponse{
			ID:          ch.ID,
			Name:        ch.Name,
			DisplayName: ch.DisplayName,
			Description: ch.Description,
			Muted:       ch.Muted,
		}
	}

	return c.JSON(result)
}

func (a *API) CreateChannel(c *fiber.Ctx) error {
	connectorID := c.Params("connectorId")
	if connectorID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "connectorId is required")
	}

	var req channelRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}

	created, err := a.channels.Create(c.Context(), channel.CreateInput{
		ConnectorID: connectorID,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Muted:       req.Muted,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(channelResponse{
		ID:          created.ID,
		Name:        created.Name,
		DisplayName: created.DisplayName,
		Description: created.Description,
		Muted:       created.Muted,
	})
}

func (a *API) UpdateChannel(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id is required")
	}

	var req channelRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}

	updated, err := a.channels.Update(c.Context(), id, channel.UpdateInput{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Muted:       req.Muted,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(channelResponse{
		ID:          updated.ID,
		Name:        updated.Name,
		DisplayName: updated.DisplayName,
		Description: updated.Description,
		Muted:       updated.Muted,
	})
}

func (a *API) DeleteChannel(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id is required")
	}

	if err := a.channels.Delete(c.Context(), id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}

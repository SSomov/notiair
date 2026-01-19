package routes

import (
	"github.com/gofiber/fiber/v2"

	"notiair/handlers"
)

type API struct {
	handlers *handlers.API
}

func New(apiHandlers *handlers.API) *API {
	return &API{handlers: apiHandlers}
}

func (a *API) Register(router fiber.Router) {
	router.Post("/notifications/dispatch", a.handlers.DispatchNotification)
	router.Get("/templates", a.handlers.ListTemplates)
	router.Post("/templates", a.handlers.SaveTemplate)
	router.Get("/workflows", a.handlers.ListWorkflows)
	router.Get("/workflows/:id", a.handlers.GetWorkflow)
	router.Post("/workflows", a.handlers.SaveWorkflow)
	router.Delete("/workflows/:id", a.handlers.DeleteWorkflow)
	router.Get("/queues/pending", a.handlers.ListQueue)
	router.Get("/connectors/telegram", a.handlers.ListTelegramTokens)
	router.Post("/connectors/telegram", a.handlers.CreateTelegramToken)
	router.Put("/connectors/telegram/:id", a.handlers.UpdateTelegramToken)
	router.Patch("/connectors/telegram/:id/active", a.handlers.ToggleTelegramTokenActive)
	router.Delete("/connectors/telegram/:id", a.handlers.DeleteTelegramToken)
	router.Get("/connectors/:connectorId/channels", a.handlers.ListChannels)
	router.Post("/connectors/:connectorId/channels", a.handlers.CreateChannel)
	router.Put("/channels/:id", a.handlers.UpdateChannel)
	router.Delete("/channels/:id", a.handlers.DeleteChannel)
}

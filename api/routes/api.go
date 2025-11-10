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
	router.Post("/workflows", a.handlers.SaveWorkflow)
	router.Get("/queues/pending", a.handlers.ListQueue)
}

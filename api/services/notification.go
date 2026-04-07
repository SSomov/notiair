package services

import (
	"context"
	"fmt"

	"notiair/internal/persistence/outbox"
	"notiair/internal/routing"
)

type WorkflowRouter interface {
	ResolveTargets(ctx context.Context, workflowID string, payload map[string]any) ([]routing.Task, error)
}

type QueueClient interface {
	Enqueue(ctx context.Context, task routing.Task) error
}

type OutboxRepository interface {
	CreatePending(ctx context.Context, input outbox.CreateInput) (outbox.Message, error)
	MarkQueued(ctx context.Context, id string) error
}

type NotificationService struct {
	router WorkflowRouter
	queue  QueueClient
	outbox OutboxRepository
}

type DispatchInput struct {
	WorkflowID string
	TemplateID string
	Variables  map[string]string
	Payload    map[string]any
}

func NewNotificationService(router WorkflowRouter, queue QueueClient, outboxRepo OutboxRepository) *NotificationService {
	return &NotificationService{
		router: router,
		queue:  queue,
		outbox: outboxRepo,
	}
}

func (s *NotificationService) Dispatch(ctx context.Context, input DispatchInput) error {
	tasks, err := s.router.ResolveTargets(ctx, input.WorkflowID, input.Payload)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		msg, err := s.outbox.CreatePending(ctx, outbox.CreateInput{
			WorkflowID: input.WorkflowID,
			ChannelID:  task.ChannelID,
			TemplateID: input.TemplateID,
			Payload:    task.Payload,
			Variables:  input.Variables,
		})
		if err != nil {
			return fmt.Errorf("outbox create: %w", err)
		}

		task.TemplateID = input.TemplateID
		task.Variables = input.Variables
		task.MessageID = msg.ID

		if err := s.queue.Enqueue(ctx, task); err != nil {
			return fmt.Errorf("enqueue task: %w", err)
		}

		if err := s.outbox.MarkQueued(ctx, msg.ID); err != nil {
			return fmt.Errorf("outbox mark queued: %w", err)
		}
	}

	return nil
}

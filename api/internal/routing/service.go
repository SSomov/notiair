package routing

import (
	"context"
	"fmt"

	"notiair/internal/workflow"
)

type WorkflowRepository interface {
	FindByID(ctx context.Context, id string) (workflow.Workflow, error)
}

type Task struct {
	WorkflowID string            `json:"workflowId"`
	ChannelID  string            `json:"channelId"`
	Payload    map[string]any    `json:"payload"`
	TemplateID string            `json:"templateId"`
	Variables  map[string]string `json:"variables"`
	MessageID  string            `json:"messageId"`
}

// Service отвечает за выбор каналов на основе workflow и фильтров события.
type Service struct {
	wfRepo WorkflowRepository
}

func NewService(repo WorkflowRepository) *Service {
	return &Service{wfRepo: repo}
}

func (s *Service) ResolveTargets(ctx context.Context, workflowID string, payload map[string]any) ([]Task, error) {
	wf, err := s.wfRepo.FindByID(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	// TODO: внедрить реальную логику фильтрации и маршрутизации
	tasks := make([]Task, 0, len(wf.Filters))
	for channelID := range wf.Filters {
		tasks = append(tasks, Task{
			WorkflowID: workflowID,
			ChannelID:  channelID,
			Payload:    payload,
		})
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("no routing targets for workflow %s", workflowID)
	}

	return tasks, nil
}

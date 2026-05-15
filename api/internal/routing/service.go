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

// Service отвечает за выбор каналов на основе workflow graph.
type Service struct {
	wfRepo     WorkflowRepository
	storageSvc StorageSaver
}

func NewService(repo WorkflowRepository, storageSvc StorageSaver) *Service {
	return &Service{wfRepo: repo, storageSvc: storageSvc}
}

func (s *Service) ResolveTargets(ctx context.Context, workflowID string, payload map[string]any) ([]Task, error) {
	wf, err := s.wfRepo.FindByID(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	if s.storageSvc == nil {
		return nil, fmt.Errorf("storage service not configured")
	}

	tasks, err := executeGraph(ctx, wf, workflowID, payload, s.storageSvc)
	if err != nil {
		if len(wf.Filters) > 0 {
			return s.resolveFromFilters(workflowID, payload, wf), nil
		}
		return nil, err
	}

	return tasks, nil
}

func (s *Service) resolveFromFilters(workflowID string, payload map[string]any, wf workflow.Workflow) []Task {
	tasks := make([]Task, 0, len(wf.Filters))
	for channelID := range wf.Filters {
		tasks = append(tasks, Task{
			WorkflowID: workflowID,
			ChannelID:  channelID,
			Payload:    payload,
		})
	}
	return tasks
}

package queue

import (
	"context"

	"notiair/internal/routing"
)

type Inspector interface {
	ListPending(ctx context.Context) ([]routing.Task, error)
}

type noopInspector struct{}

func NewNoopInspector() Inspector {
	return &noopInspector{}
}

func (n *noopInspector) ListPending(ctx context.Context) ([]routing.Task, error) {
	return []routing.Task{}, nil
}

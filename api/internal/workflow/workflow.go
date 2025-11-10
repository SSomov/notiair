package workflow

import (
	"context"
	"errors"
	"sync"
	"time"
)

type NodeType string

type Workflow struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Nodes       []Node            `json:"nodes"`
	Edges       []Edge            `json:"edges"`
	Filters     map[string]string `json:"filters"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

type Node struct {
	ID       string   `json:"id"`
	Type     NodeType `json:"type"`
	Config   any      `json:"config"`
	Position Position `json:"position"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

const (
	NodeTypeTrigger NodeType = "trigger"
	NodeTypeFilter  NodeType = "filter"
	NodeTypeAction  NodeType = "action"
)

type Repository interface {
	Save(ctx context.Context, wf Workflow) (Workflow, error)
	FindByID(ctx context.Context, id string) (Workflow, error)
	List(ctx context.Context) ([]Workflow, error)
}

type memoryRepository struct {
	mu        sync.RWMutex
	workflows map[string]Workflow
}

func NewMemoryRepository() Repository {
	return &memoryRepository{workflows: make(map[string]Workflow)}
}

func (r *memoryRepository) Save(ctx context.Context, wf Workflow) (Workflow, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	wf.UpdatedAt = time.Now()
	if wf.CreatedAt.IsZero() {
		wf.CreatedAt = wf.UpdatedAt
	}

	r.workflows[wf.ID] = wf
	return wf, nil
}

func (r *memoryRepository) FindByID(ctx context.Context, id string) (Workflow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	wf, ok := r.workflows[id]
	if !ok {
		return Workflow{}, errors.New("workflow not found")
	}
	return wf, nil
}

func (r *memoryRepository) List(ctx context.Context) ([]Workflow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Workflow, 0, len(r.workflows))
	for _, wf := range r.workflows {
		out = append(out, wf)
	}
	return out, nil
}

package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"notiair/internal/persistence/workflow"
)

type NodeType string

type Workflow struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Nodes       []Node            `json:"nodes"`
	Edges       []Edge            `json:"edges"`
	Filters     map[string]string `json:"filters"`
	IsActive    bool              `json:"isActive"`
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
	Delete(ctx context.Context, id string) error
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

func (r *memoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.workflows, id)
	return nil
}

type dbRepository struct {
	repo workflow.Repository
}

func NewDBRepository(repo workflow.Repository) Repository {
	return &dbRepository{repo: repo}
}

func (r *dbRepository) Save(ctx context.Context, wf Workflow) (Workflow, error) {
	nodesJSON, err := json.Marshal(wf.Nodes)
	if err != nil {
		return Workflow{}, err
	}

	edgesJSON, err := json.Marshal(wf.Edges)
	if err != nil {
		return Workflow{}, err
	}

	entity, err := r.repo.Save(ctx, workflow.SaveInput{
		ID:          wf.ID,
		Name:        wf.Name,
		Description: wf.Description,
		Nodes:       nodesJSON,
		Edges:       edgesJSON,
		Filters:     wf.Filters,
		IsActive:    wf.IsActive,
	})
	if err != nil {
		return Workflow{}, err
	}

	var nodes []Node
	if err := json.Unmarshal(entity.Nodes, &nodes); err != nil {
		return Workflow{}, err
	}

	var edges []Edge
	if err := json.Unmarshal(entity.Edges, &edges); err != nil {
		return Workflow{}, err
	}

	filters := make(map[string]string)
	for k, v := range entity.Filters {
		if str, ok := v.(string); ok {
			filters[k] = str
		}
	}

	return Workflow{
		ID:          entity.ID,
		Name:        entity.Name,
		Description: entity.Description,
		Nodes:       nodes,
		Edges:       edges,
		Filters:     filters,
		IsActive:    entity.IsActive,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}, nil
}

func (r *dbRepository) FindByID(ctx context.Context, id string) (Workflow, error) {
	entity, err := r.repo.FindByID(ctx, id)
	if err != nil {
		return Workflow{}, err
	}

	var nodes []Node
	if err := json.Unmarshal(entity.Nodes, &nodes); err != nil {
		return Workflow{}, err
	}

	var edges []Edge
	if err := json.Unmarshal(entity.Edges, &edges); err != nil {
		return Workflow{}, err
	}

	filters := make(map[string]string)
	for k, v := range entity.Filters {
		if str, ok := v.(string); ok {
			filters[k] = str
		}
	}

	return Workflow{
		ID:          entity.ID,
		Name:        entity.Name,
		Description: entity.Description,
		Nodes:       nodes,
		Edges:       edges,
		Filters:     filters,
		IsActive:    entity.IsActive,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}, nil
}

func (r *dbRepository) List(ctx context.Context) ([]Workflow, error) {
	entities, err := r.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	workflows := make([]Workflow, len(entities))
	for i, entity := range entities {
		var nodes []Node
		if err := json.Unmarshal(entity.Nodes, &nodes); err != nil {
			return nil, err
		}

		var edges []Edge
		if err := json.Unmarshal(entity.Edges, &edges); err != nil {
			return nil, err
		}

		filters := make(map[string]string)
		for k, v := range entity.Filters {
			if str, ok := v.(string); ok {
				filters[k] = str
			}
		}

		workflows[i] = Workflow{
			ID:          entity.ID,
			Name:        entity.Name,
			Description: entity.Description,
			Nodes:       nodes,
			Edges:       edges,
			Filters:     filters,
			IsActive:    entity.IsActive,
			CreatedAt:   entity.CreatedAt,
			UpdatedAt:   entity.UpdatedAt,
		}
	}

	return workflows, nil
}

func (r *dbRepository) Delete(ctx context.Context, id string) error {
	return r.repo.Delete(ctx, id)
}

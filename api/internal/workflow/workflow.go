package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	workflowpersist "notiair/internal/persistence/workflow"
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
	CanvasZoom  *float64          `json:"canvasZoom,omitempty"`
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

type VersionMeta struct {
	ID            string    `json:"id"`
	WorkflowID    string    `json:"workflowId"`
	VersionNumber int       `json:"versionNumber"`
	Source        string    `json:"source"`
	CreatedAt     time.Time `json:"createdAt"`
	IsActive      bool      `json:"isActive"`
	Name          string    `json:"name"`
}

type Version struct {
	VersionMeta
	Description           string            `json:"description"`
	Nodes                 []Node            `json:"nodes"`
	Edges                 []Edge            `json:"edges"`
	Filters               map[string]string `json:"filters"`
	CanvasZoom            *float64          `json:"canvasZoom,omitempty"`
	RestoredFromVersionID *string           `json:"restoredFromVersionId,omitempty"`
}

type Repository interface {
	Save(ctx context.Context, wf Workflow) (Workflow, error)
	FindByID(ctx context.Context, id string) (Workflow, error)
	List(ctx context.Context) ([]Workflow, error)
	Delete(ctx context.Context, id string) error
	ListVersions(ctx context.Context, workflowID string) ([]VersionMeta, error)
	GetVersion(ctx context.Context, workflowID, versionID string) (Version, error)
	RestoreVersion(ctx context.Context, workflowID, versionID string) (Workflow, error)
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
	repo workflowpersist.Repository
}

func NewDBRepository(repo workflowpersist.Repository) Repository {
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

	entity, err := r.repo.Save(ctx, workflowpersist.SaveInput{
		ID:          wf.ID,
		Name:        wf.Name,
		Description: wf.Description,
		Nodes:       nodesJSON,
		Edges:       edgesJSON,
		Filters:     wf.Filters,
		IsActive:    wf.IsActive,
		CanvasZoom:  wf.CanvasZoom,
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
		CanvasZoom:  entity.CanvasZoom,
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
		CanvasZoom:  entity.CanvasZoom,
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
			CanvasZoom:  entity.CanvasZoom,
			CreatedAt:   entity.CreatedAt,
			UpdatedAt:   entity.UpdatedAt,
		}
	}

	return workflows, nil
}

func (r *dbRepository) Delete(ctx context.Context, id string) error {
	return r.repo.Delete(ctx, id)
}

func (r *dbRepository) ListVersions(ctx context.Context, workflowID string) ([]VersionMeta, error) {
	metas, err := r.repo.ListVersions(ctx, workflowID)
	if err != nil {
		return nil, err
	}
	out := make([]VersionMeta, len(metas))
	for i, m := range metas {
		out[i] = VersionMeta{
			ID:            m.ID,
			WorkflowID:    m.WorkflowID,
			VersionNumber: m.VersionNumber,
			Source:        m.Source,
			CreatedAt:     m.CreatedAt,
			IsActive:      m.IsActive,
			Name:          m.Name,
		}
	}
	return out, nil
}

func (r *dbRepository) GetVersion(ctx context.Context, workflowID, versionID string) (Version, error) {
	entity, err := r.repo.FindVersionByID(ctx, workflowID, versionID)
	if err != nil {
		return Version{}, err
	}

	var nodes []Node
	if err := json.Unmarshal(entity.Nodes, &nodes); err != nil {
		return Version{}, err
	}

	var edges []Edge
	if err := json.Unmarshal(entity.Edges, &edges); err != nil {
		return Version{}, err
	}

	filters := make(map[string]string)
	for k, v := range entity.Filters {
		if str, ok := v.(string); ok {
			filters[k] = str
		}
	}

	return Version{
		VersionMeta: VersionMeta{
			ID:            entity.ID,
			WorkflowID:    entity.WorkflowID,
			VersionNumber: entity.VersionNumber,
			Source:        entity.Source,
			CreatedAt:     entity.CreatedAt,
			IsActive:      entity.IsActive,
			Name:          entity.Name,
		},
		Description:           entity.Description,
		Nodes:                 nodes,
		Edges:                 edges,
		Filters:               filters,
		CanvasZoom:            entity.CanvasZoom,
		RestoredFromVersionID: entity.RestoredFromVersionID,
	}, nil
}

func (r *dbRepository) RestoreVersion(ctx context.Context, workflowID, versionID string) (Workflow, error) {
	entity, err := r.repo.RestoreVersion(ctx, workflowID, versionID)
	if err != nil {
		return Workflow{}, err
	}
	return entityToWorkflow(entity)
}

func (r *memoryRepository) ListVersions(ctx context.Context, workflowID string) ([]VersionMeta, error) {
	return nil, errors.New("version history not supported in memory repository")
}

func (r *memoryRepository) GetVersion(ctx context.Context, workflowID, versionID string) (Version, error) {
	return Version{}, errors.New("version history not supported in memory repository")
}

func (r *memoryRepository) RestoreVersion(ctx context.Context, workflowID, versionID string) (Workflow, error) {
	return Workflow{}, errors.New("version history not supported in memory repository")
}

func entityToWorkflow(entity workflowpersist.WorkflowEntity) (Workflow, error) {
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
		CanvasZoom:  entity.CanvasZoom,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}, nil
}

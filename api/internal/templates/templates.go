package templates

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Template struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Body        string            `json:"body"`
	Variables   map[string]string `json:"variables"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

type Repository interface {
	Save(ctx context.Context, tpl Template) (Template, error)
	FindByID(ctx context.Context, id string) (Template, error)
	List(ctx context.Context) ([]Template, error)
}

type memoryRepository struct {
	mu        sync.RWMutex
	templates map[string]Template
}

func NewMemoryRepository() Repository {
	return &memoryRepository{templates: make(map[string]Template)}
}

func (r *memoryRepository) Save(ctx context.Context, tpl Template) (Template, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	tpl.UpdatedAt = time.Now()
	if tpl.CreatedAt.IsZero() {
		tpl.CreatedAt = tpl.UpdatedAt
	}

	r.templates[tpl.ID] = tpl
	return tpl, nil
}

func (r *memoryRepository) FindByID(ctx context.Context, id string) (Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tpl, ok := r.templates[id]
	if !ok {
		return Template{}, errors.New("template not found")
	}
	return tpl, nil
}

func (r *memoryRepository) List(ctx context.Context) ([]Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Template, 0, len(r.templates))
	for _, tpl := range r.templates {
		out = append(out, tpl)
	}
	return out, nil
}

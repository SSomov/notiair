package workflow

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type WorkflowEntity struct {
	ID          string            `gorm:"primaryKey"`
	Name        string            `gorm:"type:text;not null"`
	Description string            `gorm:"type:text"`
	Nodes       datatypes.JSON    `gorm:"type:jsonb"`
	Edges       datatypes.JSON    `gorm:"type:jsonb"`
	Filters     datatypes.JSONMap `gorm:"type:jsonb"`
	IsActive    bool              `gorm:"not null;default:false"`
	CreatedAt   time.Time         `gorm:"autoCreateTime"`
	UpdatedAt   time.Time         `gorm:"autoUpdateTime"`
}

type Repository interface {
	Save(ctx context.Context, input SaveInput) (WorkflowEntity, error)
	FindByID(ctx context.Context, id string) (WorkflowEntity, error)
	List(ctx context.Context) ([]WorkflowEntity, error)
	Delete(ctx context.Context, id string) error
}

type SaveInput struct {
	ID          string
	Name        string
	Description string
	Nodes       []byte
	Edges       []byte
	Filters     map[string]string
	IsActive    bool
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Save(ctx context.Context, input SaveInput) (WorkflowEntity, error) {
	var entity WorkflowEntity
	
	// Determine ID: use provided ID or generate new one
	workflowID := input.ID
	if workflowID == "" {
		workflowID = uuid.NewString()
	}

	// Try to find existing workflow
	err := r.db.WithContext(ctx).Where("id = ?", workflowID).First(&entity).Error
	isNew := errors.Is(err, gorm.ErrRecordNotFound)

	if isNew {
		// Create new workflow
		entity = WorkflowEntity{
			ID:          workflowID,
			Name:        input.Name,
			Description: input.Description,
			Nodes:       datatypes.JSON(input.Nodes),
			Edges:       datatypes.JSON(input.Edges),
			Filters:     datatypes.JSONMap{},
			IsActive:    input.IsActive,
		}
		for k, v := range input.Filters {
			entity.Filters[k] = v
		}

		if err := r.db.WithContext(ctx).Create(&entity).Error; err != nil {
			return WorkflowEntity{}, err
		}
	} else {
		// Update existing workflow
		entity.Name = input.Name
		entity.Description = input.Description
		entity.Nodes = datatypes.JSON(input.Nodes)
		entity.Edges = datatypes.JSON(input.Edges)
		entity.IsActive = input.IsActive
		entity.Filters = datatypes.JSONMap{}
		for k, v := range input.Filters {
			entity.Filters[k] = v
		}

		// Use Updates to ensure JSONB fields are properly updated
		if err := r.db.WithContext(ctx).Model(&entity).Updates(map[string]interface{}{
			"name":        entity.Name,
			"description": entity.Description,
			"nodes":       entity.Nodes,
			"edges":       entity.Edges,
			"filters":     entity.Filters,
			"is_active":   entity.IsActive,
		}).Error; err != nil {
			return WorkflowEntity{}, err
		}
	}

	return entity, nil
}

func (r *repository) FindByID(ctx context.Context, id string) (WorkflowEntity, error) {
	var entity WorkflowEntity
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error; err != nil {
		return WorkflowEntity{}, err
	}
	return entity, nil
}

func (r *repository) List(ctx context.Context) ([]WorkflowEntity, error) {
	var entities []WorkflowEntity
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&WorkflowEntity{}).Error; err != nil {
		return err
	}
	return nil
}


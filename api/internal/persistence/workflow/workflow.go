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
	CanvasZoom  *float64          `gorm:"type:double precision"`
	CreatedAt   time.Time         `gorm:"autoCreateTime"`
	UpdatedAt   time.Time         `gorm:"autoUpdateTime"`
}

type Repository interface {
	Save(ctx context.Context, input SaveInput) (WorkflowEntity, error)
	FindByID(ctx context.Context, id string) (WorkflowEntity, error)
	List(ctx context.Context) ([]WorkflowEntity, error)
	Delete(ctx context.Context, id string) error
	ListVersions(ctx context.Context, workflowID string) ([]VersionMeta, error)
	FindVersionByID(ctx context.Context, workflowID, versionID string) (WorkflowVersionEntity, error)
	RestoreVersion(ctx context.Context, workflowID, versionID string) (WorkflowEntity, error)
}

type SaveInput struct {
	ID          string
	Name        string
	Description string
	Nodes       []byte
	Edges       []byte
	Filters     map[string]string
	IsActive    bool
	CanvasZoom  *float64
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Save(ctx context.Context, input SaveInput) (WorkflowEntity, error) {
	var entity WorkflowEntity

	workflowID := input.ID
	if workflowID == "" {
		workflowID = uuid.NewString()
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Where("id = ?", workflowID).First(&entity).Error
		isNew := errors.Is(err, gorm.ErrRecordNotFound)
		if err != nil && !isNew {
			return err
		}

		filters := datatypes.JSONMap{}
		for k, v := range input.Filters {
			filters[k] = v
		}

		if isNew {
			entity = WorkflowEntity{
				ID:          workflowID,
				Name:        input.Name,
				Description: input.Description,
				Nodes:       datatypes.JSON(input.Nodes),
				Edges:       datatypes.JSON(input.Edges),
				Filters:     filters,
				IsActive:    input.IsActive,
				CanvasZoom:  input.CanvasZoom,
			}
			if err := tx.Create(&entity).Error; err != nil {
				return err
			}
		} else {
			entity.Name = input.Name
			entity.Description = input.Description
			entity.Nodes = datatypes.JSON(input.Nodes)
			entity.Edges = datatypes.JSON(input.Edges)
			entity.IsActive = input.IsActive
			entity.CanvasZoom = input.CanvasZoom
			entity.Filters = filters

			if err := tx.Model(&entity).Updates(map[string]interface{}{
				"name":        entity.Name,
				"description": entity.Description,
				"nodes":       entity.Nodes,
				"edges":       entity.Edges,
				"filters":     entity.Filters,
				"is_active":   entity.IsActive,
				"canvas_zoom": entity.CanvasZoom,
			}).Error; err != nil {
				return err
			}
		}

		return r.createVersionSnapshot(ctx, tx, entity, VersionSourceSave, nil)
	})
	if err != nil {
		return WorkflowEntity{}, err
	}

	// Reload to get updated timestamps
	return r.FindByID(ctx, entity.ID)
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
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("workflow_id = ?", id).Delete(&WorkflowVersionEntity{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", id).Delete(&WorkflowEntity{}).Error; err != nil {
			return err
		}
		return nil
	})
}

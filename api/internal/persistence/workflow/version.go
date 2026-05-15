package workflow

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	MaxVersionsPerWorkflow = 100
	VersionSourceSave      = "save"
	VersionSourceRestore   = "restore"
)

type WorkflowVersionEntity struct {
	ID                    string         `gorm:"primaryKey"`
	WorkflowID            string         `gorm:"index:idx_workflow_versions_workflow_number,priority:1;not null;constraint:OnDelete:CASCADE"`
	VersionNumber         int            `gorm:"index:idx_workflow_versions_workflow_number,priority:2;not null"`
	Name                  string         `gorm:"type:text;not null"`
	Description           string         `gorm:"type:text"`
	Nodes                 datatypes.JSON `gorm:"type:jsonb"`
	Edges                 datatypes.JSON `gorm:"type:jsonb"`
	Filters               datatypes.JSONMap `gorm:"type:jsonb"`
	IsActive              bool           `gorm:"not null"`
	CanvasZoom            *float64       `gorm:"type:double precision"`
	Source                string         `gorm:"type:text;not null"`
	RestoredFromVersionID *string        `gorm:"type:text"`
	ContentHash           string         `gorm:"type:text;not null"`
	CreatedAt             time.Time      `gorm:"autoCreateTime"`
}

type VersionMeta struct {
	ID            string    `json:"id"`
	WorkflowID    string    `json:"workflowId"`
	VersionNumber int       `json:"versionNumber"`
	Source        string    `json:"source"`
	CreatedAt     time.Time `json:"createdAt"`
	IsActive      bool      `json:"isActive"`
	Name          string    `json:"name"`
}

type VersionSnapshot struct {
	VersionMeta
	Description           string            `json:"description"`
	Nodes                 []byte            `json:"-"`
	Edges                 []byte            `json:"-"`
	Filters               map[string]string `json:"filters"`
	CanvasZoom            *float64          `json:"canvasZoom,omitempty"`
	RestoredFromVersionID *string           `json:"restoredFromVersionId,omitempty"`
}

func ComputeContentHash(nodes, edges []byte, filters map[string]string) string {
	keys := make([]string, 0, len(filters))
	for k := range filters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sortedFilters := make(map[string]string, len(filters))
	for _, k := range keys {
		sortedFilters[k] = filters[k]
	}

	payload := struct {
		Nodes   json.RawMessage   `json:"nodes"`
		Edges   json.RawMessage   `json:"edges"`
		Filters map[string]string `json:"filters"`
	}{
		Nodes:   json.RawMessage(nodes),
		Edges:   json.RawMessage(edges),
		Filters: sortedFilters,
	}
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func (r *repository) createVersionSnapshot(ctx context.Context, tx *gorm.DB, entity WorkflowEntity, source string, restoredFrom *string) error {
	hash := ComputeContentHash([]byte(entity.Nodes), []byte(entity.Edges), jsonMapToStringMap(entity.Filters))

	var latest WorkflowVersionEntity
	err := tx.WithContext(ctx).
		Where("workflow_id = ?", entity.ID).
		Order("version_number DESC").
		First(&latest).Error
	if err == nil && latest.ContentHash == hash {
		return nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	nextNumber := 1
	if err == nil {
		nextNumber = latest.VersionNumber + 1
	}

	version := WorkflowVersionEntity{
		ID:                    uuid.NewString(),
		WorkflowID:            entity.ID,
		VersionNumber:         nextNumber,
		Name:                  entity.Name,
		Description:           entity.Description,
		Nodes:                 entity.Nodes,
		Edges:                 entity.Edges,
		Filters:               entity.Filters,
		IsActive:              entity.IsActive,
		CanvasZoom:            entity.CanvasZoom,
		Source:                source,
		RestoredFromVersionID: restoredFrom,
		ContentHash:           hash,
	}
	if version.Filters == nil {
		version.Filters = datatypes.JSONMap{}
	}

	if err := tx.WithContext(ctx).Create(&version).Error; err != nil {
		return err
	}

	return pruneOldVersions(ctx, tx, entity.ID)
}

func pruneOldVersions(ctx context.Context, tx *gorm.DB, workflowID string) error {
	var ids []string
	if err := tx.WithContext(ctx).
		Model(&WorkflowVersionEntity{}).
		Select("id").
		Where("workflow_id = ?", workflowID).
		Order("version_number DESC").
		Offset(MaxVersionsPerWorkflow).
		Pluck("id", &ids).Error; err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	return tx.WithContext(ctx).Where("id IN ?", ids).Delete(&WorkflowVersionEntity{}).Error
}

func (r *repository) ListVersions(ctx context.Context, workflowID string) ([]VersionMeta, error) {
	var entities []WorkflowVersionEntity
	if err := r.db.WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Order("version_number DESC").
		Find(&entities).Error; err != nil {
		return nil, err
	}

	out := make([]VersionMeta, len(entities))
	for i, e := range entities {
		out[i] = versionEntityToMeta(e)
	}
	return out, nil
}

func (r *repository) FindVersionByID(ctx context.Context, workflowID, versionID string) (WorkflowVersionEntity, error) {
	var entity WorkflowVersionEntity
	if err := r.db.WithContext(ctx).
		Where("id = ? AND workflow_id = ?", versionID, workflowID).
		First(&entity).Error; err != nil {
		return WorkflowVersionEntity{}, err
	}
	return entity, nil
}

func (r *repository) RestoreVersion(ctx context.Context, workflowID, versionID string) (WorkflowEntity, error) {
	version, err := r.FindVersionByID(ctx, workflowID, versionID)
	if err != nil {
		return WorkflowEntity{}, err
	}

	var entity WorkflowEntity
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", workflowID).First(&entity).Error; err != nil {
			return err
		}

		entity.Name = version.Name
		entity.Description = version.Description
		entity.Nodes = version.Nodes
		entity.Edges = version.Edges
		entity.Filters = version.Filters
		entity.IsActive = version.IsActive
		entity.CanvasZoom = version.CanvasZoom
		if entity.Filters == nil {
			entity.Filters = datatypes.JSONMap{}
		}

		if err := tx.Model(&entity).Updates(map[string]interface{}{
			"name":         entity.Name,
			"description":  entity.Description,
			"nodes":        entity.Nodes,
			"edges":        entity.Edges,
			"filters":      entity.Filters,
			"is_active":    entity.IsActive,
			"canvas_zoom":  entity.CanvasZoom,
		}).Error; err != nil {
			return err
		}

		restoredFrom := version.ID
		return r.createVersionSnapshot(ctx, tx, entity, VersionSourceRestore, &restoredFrom)
	})
	if err != nil {
		return WorkflowEntity{}, err
	}

	return entity, nil
}

func versionEntityToMeta(e WorkflowVersionEntity) VersionMeta {
	return VersionMeta{
		ID:            e.ID,
		WorkflowID:    e.WorkflowID,
		VersionNumber: e.VersionNumber,
		Source:        e.Source,
		CreatedAt:     e.CreatedAt,
		IsActive:      e.IsActive,
		Name:          e.Name,
	}
}

func jsonMapToStringMap(m datatypes.JSONMap) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		if str, ok := v.(string); ok {
			out[k] = str
		}
	}
	return out
}

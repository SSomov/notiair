package storage

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Mode string

const (
	ModeRaw      Mode = "raw"
	ModeRendered Mode = "rendered"
)

type Record struct {
	ID          string            `gorm:"primaryKey"`
	WorkflowID  string            `gorm:"index;not null"`
	NodeID      string            `gorm:"index;not null"`
	Mode        Mode              `gorm:"type:text;not null"`
	ContentType string            `gorm:"type:text;not null"`
	Data        []byte            `gorm:"type:bytea;not null"`
	Metadata    datatypes.JSONMap `gorm:"type:jsonb"`
	CreatedAt   time.Time         `gorm:"autoCreateTime"`
}

func (Record) TableName() string {
	return "workflow_storage_records"
}

type CreateInput struct {
	WorkflowID  string
	NodeID      string
	Mode        Mode
	ContentType string
	Data        []byte
	Metadata    map[string]any
}

type ListFilter struct {
	WorkflowID string
	NodeID     string
	Limit      int
	Offset     int
	Search     string
}

type Repository interface {
	Create(ctx context.Context, input CreateInput) (Record, error)
	ListByNode(ctx context.Context, filter ListFilter) ([]Record, error)
	CountByNode(ctx context.Context, filter ListFilter) (int, error)
	FindByID(ctx context.Context, workflowID, recordID string) (Record, error)
	Delete(ctx context.Context, workflowID, recordID string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, input CreateInput) (Record, error) {
	meta := datatypes.JSONMap{}
	for k, v := range input.Metadata {
		meta[k] = v
	}

	rec := Record{
		ID:          uuid.NewString(),
		WorkflowID:  input.WorkflowID,
		NodeID:      input.NodeID,
		Mode:        input.Mode,
		ContentType: input.ContentType,
		Data:        input.Data,
		Metadata:    meta,
	}

	if err := r.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return Record{}, err
	}
	return rec, nil
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return limit
}

func escapeILIKE(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

func applyListFilter(q *gorm.DB, filter ListFilter) *gorm.DB {
	q = q.Where("workflow_id = ? AND node_id = ?", filter.WorkflowID, filter.NodeID)
	search := strings.TrimSpace(filter.Search)
	if search == "" {
		return q
	}
	pattern := "%" + escapeILIKE(search) + "%"
	return q.Where(
		`id ILIKE ? ESCAPE '\' OR COALESCE(metadata->>'preview', '') ILIKE ? ESCAPE '\' OR (
			content_type NOT LIKE 'application/octet-stream%%' ESCAPE '\'
			AND convert_from(data, 'UTF8') ILIKE ? ESCAPE '\'
		)`,
		pattern, pattern, pattern,
	)
}

func (r *repository) ListByNode(ctx context.Context, filter ListFilter) ([]Record, error) {
	limit := normalizeLimit(filter.Limit)

	var records []Record
	q := applyListFilter(r.db.WithContext(ctx).Model(&Record{}), filter).
		Order("created_at DESC").
		Limit(limit).
		Offset(filter.Offset)

	if err := q.Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *repository) CountByNode(ctx context.Context, filter ListFilter) (int, error) {
	var count int64
	q := applyListFilter(r.db.WithContext(ctx).Model(&Record{}), filter)
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *repository) FindByID(ctx context.Context, workflowID, recordID string) (Record, error) {
	var rec Record
	err := r.db.WithContext(ctx).
		Where("id = ? AND workflow_id = ?", recordID, workflowID).
		First(&rec).Error
	return rec, err
}

func (r *repository) Delete(ctx context.Context, workflowID, recordID string) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND workflow_id = ?", recordID, workflowID).
		Delete(&Record{}).Error
}

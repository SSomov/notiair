package outbox

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusQueued    Status = "queued"
	StatusDelivered Status = "delivered"
	StatusFailed    Status = "failed"
)

type Message struct {
	ID         string            `gorm:"primaryKey"`
	WorkflowID string            `gorm:"index;not null"`
	ChannelID  string            `gorm:"index;not null"`
	TemplateID string            `gorm:"index;not null"`
	Payload    datatypes.JSONMap `gorm:"type:jsonb"`
	Variables  datatypes.JSONMap `gorm:"type:jsonb"`
	Status     Status            `gorm:"type:text;not null"`
	RetryCount int               `gorm:"not null;default:0"`
	LastError  string            `gorm:"type:text"`
	CreatedAt  time.Time         `gorm:"autoCreateTime"`
	UpdatedAt  time.Time         `gorm:"autoUpdateTime"`
}

type CreateInput struct {
	WorkflowID string
	ChannelID  string
	TemplateID string
	Payload    map[string]any
	Variables  map[string]string
}

type Repository interface {
	CreatePending(ctx context.Context, input CreateInput) (Message, error)
	MarkQueued(ctx context.Context, id string) error
	MarkDelivered(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string, lastError string, retryCount int) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreatePending(ctx context.Context, input CreateInput) (Message, error) {
	payload := datatypes.JSONMap{}
	for k, v := range input.Payload {
		payload[k] = v
	}

	vars := datatypes.JSONMap{}
	for k, v := range input.Variables {
		vars[k] = v
	}

	msg := Message{
		ID:         uuid.NewString(),
		WorkflowID: input.WorkflowID,
		ChannelID:  input.ChannelID,
		TemplateID: input.TemplateID,
		Payload:    payload,
		Variables:  vars,
		Status:     StatusPending,
	}

	if err := r.db.WithContext(ctx).Create(&msg).Error; err != nil {
		return Message{}, err
	}

	return msg, nil
}

func (r *repository) MarkQueued(ctx context.Context, id string) error {
	return r.updateStatus(ctx, id, StatusQueued, "", nil)
}

func (r *repository) MarkDelivered(ctx context.Context, id string) error {
	return r.updateStatus(ctx, id, StatusDelivered, "", nil)
}

func (r *repository) MarkFailed(ctx context.Context, id string, lastError string, retryCount int) error {
	return r.updateStatus(ctx, id, StatusFailed, lastError, &retryCount)
}

func (r *repository) updateStatus(ctx context.Context, id string, status Status, lastError string, retryCount *int) error {
	updates := map[string]any{
		"status":     status,
		"last_error": lastError,
	}
	if retryCount != nil {
		updates["retry_count"] = *retryCount
	}

	return r.db.WithContext(ctx).
		Model(&Message{}).
		Where("id = ?", id).
		Updates(updates).Error
}

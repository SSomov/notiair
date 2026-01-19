package channel

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Channel struct {
	ID          string    `gorm:"primaryKey"`
	ConnectorID string    `gorm:"index;not null"`
	Name        string    `gorm:"type:text;not null"`
	DisplayName string    `gorm:"type:text"`
	Description string    `gorm:"type:text"`
	Muted       bool      `gorm:"not null;default:false"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

type CreateInput struct {
	ConnectorID string
	Name        string
	DisplayName string
	Description string
	Muted       bool
}

type UpdateInput struct {
	Name        string
	DisplayName string
	Description string
	Muted       bool
}

type Repository interface {
	ListByConnector(ctx context.Context, connectorID string) ([]Channel, error)
	Create(ctx context.Context, input CreateInput) (Channel, error)
	Update(ctx context.Context, id string, input UpdateInput) (Channel, error)
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) ListByConnector(ctx context.Context, connectorID string) ([]Channel, error) {
	var channels []Channel
	if err := r.db.WithContext(ctx).
		Where("connector_id = ?", connectorID).
		Order("created_at ASC").
		Find(&channels).Error; err != nil {
		return nil, err
	}
	return channels, nil
}

func (r *repository) Create(ctx context.Context, input CreateInput) (Channel, error) {
	channel := Channel{
		ID:          uuid.NewString(),
		ConnectorID: input.ConnectorID,
		Name:        input.Name,
		DisplayName: input.DisplayName,
		Description: input.Description,
		Muted:       input.Muted,
	}

	if err := r.db.WithContext(ctx).Create(&channel).Error; err != nil {
		return Channel{}, err
	}

	return channel, nil
}

func (r *repository) Update(ctx context.Context, id string, input UpdateInput) (Channel, error) {
	var channel Channel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&channel).Error; err != nil {
		return Channel{}, err
	}

	channel.Name = input.Name
	channel.DisplayName = input.DisplayName
	channel.Description = input.Description
	channel.Muted = input.Muted

	if err := r.db.WithContext(ctx).Save(&channel).Error; err != nil {
		return Channel{}, err
	}

	return channel, nil
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&Channel{}).Error
}


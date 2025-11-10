package serviceconfig

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Type string

const (
	TypeTelegram Type = "telegram"
	TypeDefault  Type = "default"
)

type ServiceConfig struct {
	ID        string            `gorm:"primaryKey"`
	Type      Type              `gorm:"type:text;not null"`
	IsDefault bool              `gorm:"not null;default:false"`
	IsActive  bool              `gorm:"not null;default:true"`
	Settings  datatypes.JSONMap `gorm:"type:jsonb"`
	CreatedAt time.Time         `gorm:"autoCreateTime"`
	UpdatedAt time.Time         `gorm:"autoUpdateTime"`
}

type Repository interface {
	List(ctx context.Context) ([]ServiceConfig, error)
	Create(ctx context.Context, input CreateInput) (ServiceConfig, error)
	SetActive(ctx context.Context, id string, active bool) error
	SetDefault(ctx context.Context, id string) error
	EnsureDefault(ctx context.Context, svcType Type) (ServiceConfig, error)
}

type CreateInput struct {
	Type      Type
	IsDefault bool
	IsActive  bool
	Settings  map[string]any
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List(ctx context.Context) ([]ServiceConfig, error) {
	var result []ServiceConfig
	if err := r.db.WithContext(ctx).Order("created_at ASC").Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (r *repository) Create(ctx context.Context, input CreateInput) (ServiceConfig, error) {
	settings := datatypes.JSONMap{}
	for k, v := range input.Settings {
		settings[k] = v
	}

	cfg := ServiceConfig{
		ID:        uuid.NewString(),
		Type:      input.Type,
		IsDefault: input.IsDefault,
		IsActive:  input.IsActive,
		Settings:  settings,
	}

	return cfg, r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if input.IsDefault {
			if err := tx.Model(&ServiceConfig{}).Where("type = ?", input.Type).Update("is_default", false).Error; err != nil {
				return err
			}
		}

		if err := tx.Create(&cfg).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *repository) SetActive(ctx context.Context, id string, active bool) error {
	return r.db.WithContext(ctx).Model(&ServiceConfig{}).
		Where("id = ?", id).
		Update("is_active", active).Error
}

func (r *repository) SetDefault(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cfg ServiceConfig
		if err := tx.Where("id = ?", id).First(&cfg).Error; err != nil {
			return err
		}

		if err := tx.Model(&ServiceConfig{}).Where("type = ?", cfg.Type).Update("is_default", false).Error; err != nil {
			return err
		}

		return tx.Model(&ServiceConfig{}).
			Where("id = ?", id).
			Update("is_default", true).Error
	})
}

func (r *repository) EnsureDefault(ctx context.Context, svcType Type) (ServiceConfig, error) {
	var cfg ServiceConfig
	err := r.db.WithContext(ctx).
		Where("type = ? AND is_default = TRUE", svcType).
		First(&cfg).Error

	if err == nil {
		return cfg, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return ServiceConfig{}, err
	}

	settings := datatypes.JSONMap{}
	if svcType == TypeTelegram {
		settings["token"] = ""
	}

	cfg = ServiceConfig{
		ID:        uuid.NewString(),
		Type:      svcType,
		IsDefault: true,
		IsActive:  true,
		Settings:  settings,
	}

	if err := r.db.WithContext(ctx).Create(&cfg).Error; err != nil {
		return ServiceConfig{}, err
	}

	return cfg, nil
}

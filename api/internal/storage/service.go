package storage

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	persiststorage "notiair/internal/persistence/storage"
	"notiair/internal/template"
)

const previewMaxLen = 200

type Repository interface {
	Create(ctx context.Context, input persiststorage.CreateInput) (persiststorage.Record, error)
	ListByNode(ctx context.Context, filter persiststorage.ListFilter) ([]persiststorage.Record, error)
	FindByID(ctx context.Context, workflowID, recordID string) (persiststorage.Record, error)
	Delete(ctx context.Context, workflowID, recordID string) error
}

type SaveInput struct {
	WorkflowID   string
	NodeID       string
	Mode         persiststorage.Mode
	Payload      map[string]any
	TemplateBody string
	// Data and ContentType come from the upstream block output (preferred).
	Data        []byte
	ContentType string
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Save(ctx context.Context, input SaveInput) (persiststorage.Record, error) {
	data, contentType, err := buildData(input)
	if err != nil {
		return persiststorage.Record{}, err
	}

	preview := buildPreview(data, contentType)
	meta := map[string]any{
		"preview": preview,
		"size":    len(data),
	}
	if eventID, ok := input.Payload["event_id"].(string); ok && eventID != "" {
		meta["event_id"] = eventID
	}

	return s.repo.Create(ctx, persiststorage.CreateInput{
		WorkflowID:  input.WorkflowID,
		NodeID:      input.NodeID,
		Mode:        input.Mode,
		ContentType: contentType,
		Data:        data,
		Metadata:    meta,
	})
}

func buildData(input SaveInput) ([]byte, string, error) {
	if input.ContentType != "" {
		return input.Data, input.ContentType, nil
	}

	switch input.Mode {
	case persiststorage.ModeRaw:
		if b64, ok := input.Payload["_binary"].(string); ok && b64 != "" {
			raw, err := base64.StdEncoding.DecodeString(b64)
			if err != nil {
				return nil, "", fmt.Errorf("decode _binary: %w", err)
			}
			return raw, "application/octet-stream", nil
		}
		data, err := json.Marshal(input.Payload)
		if err != nil {
			return nil, "", err
		}
		return data, "application/json", nil

	case persiststorage.ModeRendered:
		body := input.TemplateBody
		if body == "" {
			data, err := json.Marshal(input.Payload)
			if err != nil {
				return nil, "", err
			}
			return data, "application/json", nil
		}
		rendered := template.Render(body, input.Payload)
		return []byte(rendered), "text/plain; charset=utf-8", nil

	default:
		return nil, "", fmt.Errorf("unknown storage mode: %s", input.Mode)
	}
}

func buildPreview(data []byte, contentType string) string {
	if strings.HasPrefix(contentType, "application/octet-stream") {
		return fmt.Sprintf("[binary %d bytes]", len(data))
	}
	s := string(data)
	if !utf8.ValidString(s) {
		return fmt.Sprintf("[binary %d bytes]", len(data))
	}
	if len(s) > previewMaxLen {
		return s[:previewMaxLen] + "…"
	}
	return s
}

func (s *Service) ListByNode(ctx context.Context, filter persiststorage.ListFilter) ([]persiststorage.Record, error) {
	return s.repo.ListByNode(ctx, filter)
}

func (s *Service) GetByID(ctx context.Context, workflowID, recordID string) (persiststorage.Record, error) {
	return s.repo.FindByID(ctx, workflowID, recordID)
}

func (s *Service) Delete(ctx context.Context, workflowID, recordID string) error {
	return s.repo.Delete(ctx, workflowID, recordID)
}

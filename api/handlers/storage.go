package handlers

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"

	persiststorage "notiair/internal/persistence/storage"
)

type StorageReader interface {
	ListByNode(ctx context.Context, filter persiststorage.ListFilter) ([]persiststorage.Record, error)
	CountByNode(ctx context.Context, filter persiststorage.ListFilter) (int, error)
	GetByID(ctx context.Context, workflowID, recordID string) (persiststorage.Record, error)
	Delete(ctx context.Context, workflowID, recordID string) error
}

type storageListResponse struct {
	Items []storageRecordResponse `json:"items"`
	Total int                     `json:"total"`
}

type storageRecordResponse struct {
	ID          string         `json:"id"`
	WorkflowID  string         `json:"workflowId"`
	NodeID      string         `json:"nodeId"`
	Mode        string         `json:"mode"`
	ContentType string         `json:"contentType"`
	Size        int            `json:"size"`
	Preview     string         `json:"preview"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	CreatedAt   string         `json:"createdAt"`
}

func recordToListItem(rec persiststorage.Record) storageRecordResponse {
	preview := ""
	size := len(rec.Data)
	if rec.Metadata != nil {
		if p, ok := rec.Metadata["preview"].(string); ok {
			preview = p
		}
		if s, ok := rec.Metadata["size"].(float64); ok {
			size = int(s)
		}
	}

	meta := make(map[string]any)
	for k, v := range rec.Metadata {
		if k != "preview" && k != "size" {
			meta[k] = v
		}
	}

	return storageRecordResponse{
		ID:          rec.ID,
		WorkflowID:  rec.WorkflowID,
		NodeID:      rec.NodeID,
		Mode:        string(rec.Mode),
		ContentType: rec.ContentType,
		Size:        size,
		Preview:     preview,
		Metadata:    meta,
		CreatedAt:   rec.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (a *API) ListStorageRecords(c *fiber.Ctx) error {
	workflowID := c.Params("id")
	nodeID := c.Query("nodeId")
	if workflowID == "" || nodeID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "workflow id and nodeId query are required")
	}

	limit := 20
	offset := 0
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			offset = n
		}
	}

	listFilter := persiststorage.ListFilter{
		WorkflowID: workflowID,
		NodeID:     nodeID,
		Limit:      limit,
		Offset:     offset,
		Search:     c.Query("q"),
	}

	total, err := a.storage.CountByNode(c.Context(), listFilter)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	records, err := a.storage.ListByNode(c.Context(), listFilter)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	items := make([]storageRecordResponse, len(records))
	for i, rec := range records {
		items[i] = recordToListItem(rec)
	}

	return c.JSON(storageListResponse{Items: items, Total: total})
}

func (a *API) GetStorageRecord(c *fiber.Ctx) error {
	workflowID := c.Params("id")
	recordID := c.Params("recordId")
	if workflowID == "" || recordID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id and recordId are required")
	}

	rec, err := a.storage.GetByID(c.Context(), workflowID, recordID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "record not found")
	}

	preview := ""
	if rec.Metadata != nil {
		if p, ok := rec.Metadata["preview"].(string); ok {
			preview = p
		}
	}

	return c.JSON(fiber.Map{
		"id":          rec.ID,
		"workflowId":  rec.WorkflowID,
		"nodeId":      rec.NodeID,
		"mode":        rec.Mode,
		"contentType": rec.ContentType,
		"size":        len(rec.Data),
		"preview":     preview,
		"data":        string(rec.Data),
		"metadata":    rec.Metadata,
		"createdAt":   rec.CreatedAt,
	})
}

func (a *API) DeleteStorageRecord(c *fiber.Ctx) error {
	workflowID := c.Params("id")
	recordID := c.Params("recordId")
	if workflowID == "" || recordID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "id and recordId are required")
	}

	if err := a.storage.Delete(c.Context(), workflowID, recordID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}

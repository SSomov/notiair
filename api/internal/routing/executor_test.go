package routing

import (
	"context"
	"encoding/json"
	"testing"

	persiststorage "notiair/internal/persistence/storage"
	"notiair/internal/storage"
	"notiair/internal/workflow"
)

type mockStorage struct {
	saved []storage.SaveInput
}

func (m *mockStorage) Save(ctx context.Context, input storage.SaveInput) (persiststorage.Record, error) {
	m.saved = append(m.saved, input)
	return persiststorage.Record{ID: "rec-1"}, nil
}

func TestExecuteGraph_StorageAfterTemplateStoresRendered(t *testing.T) {
	mock := &mockStorage{}
	wf := workflow.Workflow{
		ID: "wf-1",
		Nodes: []workflow.Node{
			{ID: "tr", Type: workflow.NodeTypeTrigger, Config: map[string]any{"variant": "trigger"}},
			{ID: "tpl", Type: workflow.NodeTypeAction, Config: map[string]any{
				"variant":      "template",
				"templateBody": "Hello {{name}}!",
			}},
			{ID: "st", Type: workflow.NodeTypeAction, Config: map[string]any{
				"variant":     "storage",
				"storageMode": "raw",
			}},
		},
		Edges: []workflow.Edge{
			{From: "tr", To: "tpl"},
			{From: "tpl", To: "st"},
		},
	}

	tasks, err := executeGraph(context.Background(), wf, "wf-1", map[string]any{"name": "World"}, mock)
	if err != nil {
		t.Fatalf("executeGraph: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected no channel tasks, got %d", len(tasks))
	}
	if len(mock.saved) != 1 {
		t.Fatalf("expected 1 save, got %d", len(mock.saved))
	}
	got := string(mock.saved[0].Data)
	if got != "Hello World!" {
		t.Fatalf("stored %q, want rendered text", got)
	}
	if mock.saved[0].Mode != persiststorage.ModeRendered {
		t.Fatalf("mode %s, want rendered", mock.saved[0].Mode)
	}
}

func TestExecuteGraph_StorageAfterTriggerStoresJSON(t *testing.T) {
	mock := &mockStorage{}
	wf := workflow.Workflow{
		ID: "wf-1",
		Nodes: []workflow.Node{
			{ID: "tr", Type: workflow.NodeTypeTrigger, Config: map[string]any{"variant": "trigger"}},
			{ID: "st", Type: workflow.NodeTypeAction, Config: map[string]any{
				"variant":     "storage",
				"storageMode": "raw",
			}},
		},
		Edges: []workflow.Edge{{From: "tr", To: "st"}},
	}

	_, err := executeGraph(context.Background(), wf, "wf-1", map[string]any{"x": 1}, mock)
	if err != nil {
		t.Fatalf("executeGraph: %v", err)
	}
	if len(mock.saved) != 1 {
		t.Fatalf("expected 1 save, got %d", len(mock.saved))
	}
	var decoded map[string]any
	if err := json.Unmarshal(mock.saved[0].Data, &decoded); err != nil {
		t.Fatalf("expected json bytes: %v", err)
	}
	if decoded["x"].(float64) != 1 {
		t.Fatalf("unexpected payload %v", decoded)
	}
}

func TestExecuteGraph_TemplateStorageChannelPayloadRendered(t *testing.T) {
	mock := &mockStorage{}
	wf := workflow.Workflow{
		ID: "wf-1",
		Nodes: []workflow.Node{
			{ID: "tr", Type: workflow.NodeTypeTrigger, Config: map[string]any{"variant": "trigger"}},
			{ID: "tpl", Type: workflow.NodeTypeAction, Config: map[string]any{
				"variant":      "template",
				"templateBody": "Hi {{name}}",
			}},
			{ID: "st", Type: workflow.NodeTypeAction, Config: map[string]any{"variant": "storage"}},
			{ID: "ch", Type: workflow.NodeTypeAction, Config: map[string]any{
				"variant":   "channel",
				"channelId": "chan-1",
			}},
		},
		Edges: []workflow.Edge{
			{From: "tr", To: "tpl"},
			{From: "tpl", To: "st"},
			{From: "st", To: "ch"},
		},
	}

	tasks, err := executeGraph(context.Background(), wf, "wf-1", map[string]any{"name": "Ann"}, mock)
	if err != nil {
		t.Fatalf("executeGraph: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	body, ok := tasks[0].Payload["body"].(string)
	if !ok || body != "Hi Ann" {
		t.Fatalf("channel payload %v, want body=Hi Ann", tasks[0].Payload)
	}
	if len(tasks[0].Payload) != 1 {
		t.Fatalf("expected only rendered body in payload, got %v", tasks[0].Payload)
	}
}

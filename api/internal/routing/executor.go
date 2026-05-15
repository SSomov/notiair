package routing

import (
	"context"
	"encoding/json"
	"fmt"

	persiststorage "notiair/internal/persistence/storage"
	"notiair/internal/storage"
	tplrender "notiair/internal/template"
	"notiair/internal/workflow"
)

type nodeConfig struct {
	Variant         string         `json:"variant"`
	ChannelID       string         `json:"channelId"`
	StorageMode     string         `json:"storageMode"`
	TemplateBody    string         `json:"templateBody"`
	TemplatePayload map[string]any `json:"templatePayload"`
}

// flowData is the output passed along edges (from the block on the left).
type flowData struct {
	Data        []byte
	ContentType string
	Mode        persiststorage.Mode
	Payload     map[string]any
}

type StorageSaver interface {
	Save(ctx context.Context, input storage.SaveInput) (persiststorage.Record, error)
}

func parseNodeConfig(node workflow.Node) nodeConfig {
	var cfg nodeConfig
	b, err := json.Marshal(node.Config)
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(b, &cfg)
	return cfg
}

func buildAdjacency(edges []workflow.Edge) map[string][]string {
	adj := make(map[string][]string)
	for _, e := range edges {
		adj[e.From] = append(adj[e.From], e.To)
	}
	return adj
}

func findTriggerIDs(nodes []workflow.Node) []string {
	var ids []string
	for _, n := range nodes {
		if n.Type == workflow.NodeTypeTrigger {
			ids = append(ids, n.ID)
		}
	}
	return ids
}

func nodeByID(nodes []workflow.Node) map[string]workflow.Node {
	m := make(map[string]workflow.Node, len(nodes))
	for _, n := range nodes {
		m[n.ID] = n
	}
	return m
}

func initialFlow(payload map[string]any) (flowData, error) {
	if payload == nil {
		payload = map[string]any{}
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return flowData{}, err
	}
	return flowData{
		Data:        data,
		ContentType: "application/json",
		Mode:        persiststorage.ModeRaw,
		Payload:     payload,
	}, nil
}

func templateOutput(in flowData, templateBody string) flowData {
	rendered := tplrender.Render(templateBody, in.Payload)
	return flowData{
		Data:        []byte(rendered),
		ContentType: "text/plain; charset=utf-8",
		Mode:        persiststorage.ModeRendered,
		Payload:     map[string]any{"body": rendered},
	}
}

func renderedBody(in flowData) string {
	if in.Payload != nil {
		if b, ok := in.Payload["body"].(string); ok {
			return b
		}
	}
	return string(in.Data)
}

// payloadForChannel sends rendered template text to delivery, not the raw trigger JSON.
func payloadForChannel(in flowData) map[string]any {
	if in.Mode == persiststorage.ModeRendered {
		return map[string]any{"body": renderedBody(in)}
	}
	if in.Payload != nil {
		return in.Payload
	}
	return map[string]any{}
}

// executeGraph walks from triggers; each node receives the left block's output.
func executeGraph(
	ctx context.Context,
	wf workflow.Workflow,
	workflowID string,
	payload map[string]any,
	storageSvc StorageSaver,
) ([]Task, error) {
	adj := buildAdjacency(wf.Edges)
	nodes := nodeByID(wf.Nodes)
	triggerIDs := findTriggerIDs(wf.Nodes)

	if len(triggerIDs) == 0 {
		return nil, fmt.Errorf("workflow %s has no trigger nodes", workflowID)
	}

	start, err := initialFlow(payload)
	if err != nil {
		return nil, err
	}

	visited := make(map[string]bool)
	var tasks []Task

	var walk func(nodeID string, in flowData) error
	walk = func(nodeID string, in flowData) error {
		if visited[nodeID] {
			return nil
		}
		visited[nodeID] = true

		node, ok := nodes[nodeID]
		if !ok {
			return nil
		}

		cfg := parseNodeConfig(node)
		out := in

		switch cfg.Variant {
		case "template":
			out = templateOutput(in, cfg.TemplateBody)

		case "storage":
			saveMode := in.Mode
			if cfg.StorageMode == "raw" && in.Mode != persiststorage.ModeRendered {
				saveMode = persiststorage.ModeRaw
			}
			if _, err := storageSvc.Save(ctx, storage.SaveInput{
				WorkflowID:  workflowID,
				NodeID:      nodeID,
				Mode:        saveMode,
				Payload:     in.Payload,
				Data:        in.Data,
				ContentType: in.ContentType,
			}); err != nil {
				return fmt.Errorf("storage save node %s: %w", nodeID, err)
			}
			out = in

		case "channel":
			if cfg.ChannelID != "" {
				tasks = append(tasks, Task{
					WorkflowID: workflowID,
					ChannelID:  cfg.ChannelID,
					Payload:    payloadForChannel(in),
				})
			}
		}

		for _, nextID := range adj[nodeID] {
			if err := walk(nextID, out); err != nil {
				return err
			}
		}
		return nil
	}

	for _, tid := range triggerIDs {
		if err := walk(tid, start); err != nil {
			return nil, err
		}
	}

	// Storage-only (no channel downstream): success with no delivery tasks.
	if len(tasks) == 0 {
		return []Task{}, nil
	}

	return tasks, nil
}

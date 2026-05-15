package routing

import (
	"context"
	"encoding/json"
	"fmt"

	persiststorage "notiair/internal/persistence/storage"
	"notiair/internal/storage"
	"notiair/internal/workflow"
)

type nodeConfig struct {
	Variant       string `json:"variant"`
	ChannelID     string `json:"channelId"`
	StorageMode   string `json:"storageMode"`
	TemplateBody  string `json:"templateBody"`
	TemplatePayload map[string]any `json:"templatePayload"`
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

func findUpstreamTemplate(wf workflow.Workflow, storageNodeID string) *workflow.Node {
	incoming := make(map[string]bool)
	for _, e := range wf.Edges {
		if e.To == storageNodeID {
			incoming[e.From] = true
		}
	}
	for id := range incoming {
		n, ok := nodeByID(wf.Nodes)[id]
		if !ok {
			continue
		}
		cfg := parseNodeConfig(n)
		if cfg.Variant == "template" {
			return &n
		}
	}
	return nil
}

func findFirstTemplate(wf workflow.Workflow) *workflow.Node {
	for i := range wf.Nodes {
		cfg := parseNodeConfig(wf.Nodes[i])
		if cfg.Variant == "template" {
			return &wf.Nodes[i]
		}
	}
	return nil
}

// executeGraph walks the workflow graph from triggers, persists storage nodes, and returns channel tasks.
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

	visited := make(map[string]bool)
	var tasks []Task

	var walk func(nodeID string) error
	walk = func(nodeID string) error {
		if visited[nodeID] {
			return nil
		}
		visited[nodeID] = true

		node, ok := nodes[nodeID]
		if !ok {
			return nil
		}

		cfg := parseNodeConfig(node)

		switch cfg.Variant {
		case "storage":
			mode := persiststorage.ModeRaw
			if cfg.StorageMode == "rendered" {
				mode = persiststorage.ModeRendered
			}
			tplBody := ""
			tplNode := findUpstreamTemplate(wf, nodeID)
			if tplNode == nil {
				tplNode = findFirstTemplate(wf)
			}
			if tplNode != nil {
				tplCfg := parseNodeConfig(*tplNode)
				tplBody = tplCfg.TemplateBody
			}
			if _, err := storageSvc.Save(ctx, storage.SaveInput{
				WorkflowID:   workflowID,
				NodeID:       nodeID,
				Mode:         mode,
				Payload:      payload,
				TemplateBody: tplBody,
			}); err != nil {
				return fmt.Errorf("storage save node %s: %w", nodeID, err)
			}

		case "channel":
			if cfg.ChannelID != "" {
				tasks = append(tasks, Task{
					WorkflowID: workflowID,
					ChannelID:  cfg.ChannelID,
					Payload:    payload,
				})
			}
		}

		for _, nextID := range adj[nodeID] {
			if err := walk(nextID); err != nil {
				return err
			}
		}
		return nil
	}

	for _, tid := range triggerIDs {
		if err := walk(tid); err != nil {
			return nil, err
		}
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("no routing targets for workflow %s", workflowID)
	}

	return tasks, nil
}

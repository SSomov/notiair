import type {
	WorkflowDraft,
	WorkflowEdge,
	WorkflowNode,
} from "$lib/types/workflow";

export const WORKSPACE_ZOOM_MIN = 0.25;
export const WORKSPACE_ZOOM_MAX = 2;

export function clampWorkspaceZoom(z: number): number {
	return Math.min(WORKSPACE_ZOOM_MAX, Math.max(WORKSPACE_ZOOM_MIN, z));
}

export type WorkflowEditorEdge = {
	from?: { nodeId?: string };
	to?: { nodeId?: string };
};

/** Поля ноды редактора, попадающие в payload сохранения */
export type WorkflowEditorCanvasNode = {
	id: string;
	label: string;
	description: string;
	variant: "trigger" | "template" | "storage" | "channel";
	position: { x: number; y: number };
	selectedChannelId?: string;
	selectedChannelName?: string;
	selectedChannelConnectorId?: string;
	selectedChannelConnectorType?: "telegram" | "slack" | "smtp";
	templateBody?: string;
	templatePayload?: Record<string, unknown>;
	triggerPayload?: Record<string, unknown>;
	eventTypes?: string[];
	storageMode?: "raw" | "rendered";
};

export type WorkflowEditorPersistInput = {
	workflowName: string;
	workflowDescription: string;
	isActive: boolean;
	workspaceZoom: number;
	defaultNewWorkflowName: string;
	nodes: WorkflowEditorCanvasNode[];
	edges: WorkflowEditorEdge[];
};

function buildNodeConfig(
	node: WorkflowEditorCanvasNode,
): Record<string, unknown> {
	const config: Record<string, unknown> = {
		label: node.label,
		description: node.description,
		variant: node.variant,
	};

	if (node.variant === "channel" && node.selectedChannelId) {
		config.channelId = node.selectedChannelId;
		config.channelName = node.selectedChannelName;
		config.connectorId = node.selectedChannelConnectorId;
		config.connectorType = node.selectedChannelConnectorType;
	}

	if (
		node.variant === "template" &&
		(node.templateBody || node.templatePayload)
	) {
		config.templateBody = node.templateBody;
		config.templatePayload = node.templatePayload;
	}

	if (node.variant === "trigger" && node.triggerPayload) {
		config.triggerPayload = node.triggerPayload;
	}

	if (node.variant === "trigger" && node.eventTypes) {
		config.eventTypes = node.eventTypes;
	}

	if (node.variant === "storage") {
		config.storageMode = node.storageMode || "raw";
	}

	return config;
}

function mapNodesToWorkflowNodes(
	nodes: WorkflowEditorCanvasNode[],
): WorkflowNode[] {
	const mapped = nodes.map((node) => ({
		id: node.id,
		type: (node.variant === "trigger"
			? "trigger"
			: "action") as WorkflowNode["type"],
		position: {
			x: node.position.x,
			y: node.position.y,
		},
		config: buildNodeConfig(node),
	}));
	mapped.sort((a, b) => a.id.localeCompare(b.id));
	return mapped;
}

function mapEdgesToWorkflowEdges(edges: WorkflowEditorEdge[]): WorkflowEdge[] {
	const list: WorkflowEdge[] = [];
	for (const edge of edges) {
		const from = edge.from?.nodeId;
		const to = edge.to?.nodeId;
		if (from && to) {
			list.push({ from, to });
		}
	}
	list.sort((a, b) => {
		const c = a.from.localeCompare(b.from);
		return c !== 0 ? c : a.to.localeCompare(b.to);
	});
	return list;
}

/**
 * Тело запроса сохранения без `id` — совпадает по полям с тем, что собирает редактор перед saveWorkflow.
 */
export function buildWorkflowPersistPayload(
	input: WorkflowEditorPersistInput,
): Omit<WorkflowDraft, "id" | "createdAt" | "updatedAt" | "activeNode"> {
	const name = input.workflowName.trim() || input.defaultNewWorkflowName;
	return {
		name,
		description: input.workflowDescription.trim(),
		nodes: mapNodesToWorkflowNodes(input.nodes),
		edges: mapEdgesToWorkflowEdges(input.edges),
		filters: {},
		isActive: input.isActive,
		canvasZoom: clampWorkspaceZoom(input.workspaceZoom),
	};
}

export function workflowEditorStateFingerprint(
	input: WorkflowEditorPersistInput,
): string {
	return JSON.stringify(buildWorkflowPersistPayload(input));
}

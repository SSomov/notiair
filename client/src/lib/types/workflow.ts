export type WorkflowNodeType = 'trigger' | 'filter' | 'action';

export type WorkflowNode = {
	id: string;
	type: WorkflowNodeType;
	position: { x: number; y: number };
	config: Record<string, unknown>;
};

export type WorkflowEdge = {
	from: string;
	to: string;
};

export type WorkflowDraft = {
	id: string;
	name: string;
	nodes: WorkflowNode[];
	edges: WorkflowEdge[];
	filters: Record<string, string>;
	activeNode: WorkflowNode | null;
};

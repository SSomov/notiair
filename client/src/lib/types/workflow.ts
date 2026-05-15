export type WorkflowNodeType = "trigger" | "filter" | "action";

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

export type WorkflowVersionMeta = {
	id: string;
	workflowId: string;
	versionNumber: number;
	source: "save" | "restore";
	createdAt: string;
	isActive: boolean;
	name: string;
};

export type WorkflowVersion = WorkflowVersionMeta & {
	description?: string;
	nodes: WorkflowNode[];
	edges: WorkflowEdge[];
	filters: Record<string, string>;
	canvasZoom?: number;
	restoredFromVersionId?: string;
};

export type WorkflowDraft = {
	id: string;
	name: string;
	description?: string;
	nodes: WorkflowNode[];
	edges: WorkflowEdge[];
	filters: Record<string, string>;
	isActive?: boolean;
	/** Масштаб холста редактора (например 1, 1.25); опционально для старых записей */
	canvasZoom?: number;
	createdAt?: string;
	updatedAt?: string;
	activeNode?: WorkflowNode | null;
};

import { writable } from 'svelte/store';
import type { WorkflowDraft } from '$lib/types/workflow';

export const workflowStore = writable<WorkflowDraft[]>([
  {
    id: 'demo-workflow',
    name: 'Новый workflow',
    nodes: [],
    edges: [],
    filters: {},
    activeNode: null
  }
]);


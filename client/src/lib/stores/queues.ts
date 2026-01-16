import { writable } from "svelte/store";
import type { QueueItem } from "$lib/types/queue";

export const queueStore = writable<QueueItem[]>([
	{
		taskId: "task-1",
		workflowId: "demo-workflow",
		channelId: "@notiair-demo",
		attempts: 0,
		status: "pending",
	},
]);

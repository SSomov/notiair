export type QueueStatus = 'pending' | 'processing' | 'failed' | 'completed';

export type QueueItem = {
	taskId: string;
	workflowId: string;
	channelId: string;
	attempts: number;
	status: QueueStatus;
};

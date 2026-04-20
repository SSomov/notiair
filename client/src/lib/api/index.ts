import type { QueueItem } from "$lib/types/queue";
import type { WorkflowDraft } from "$lib/types/workflow";

const API_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080/api/v1";

export async function listWorkflows(): Promise<WorkflowDraft[]> {
	const res = await fetch(`${API_URL}/workflows`);
	if (!res.ok) throw new Error("errors.loadWorkflows");
	const data = await res.json();
	return Array.isArray(data) ? data : [];
}

export async function getWorkflow(id: string): Promise<WorkflowDraft> {
	const res = await fetch(`${API_URL}/workflows/${id}`);
	if (!res.ok) throw new Error("errors.loadWorkflow");
	return res.json();
}

export async function saveWorkflow(
	payload: WorkflowDraft,
): Promise<WorkflowDraft> {
	const res = await fetch(`${API_URL}/workflows`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(payload),
	});
	if (!res.ok) throw new Error("errors.saveWorkflow");
	return res.json();
}

export async function deleteWorkflow(id: string): Promise<void> {
	const res = await fetch(`${API_URL}/workflows/${id}`, {
		method: "DELETE",
	});
	if (!res.ok) throw new Error("errors.deleteWorkflow");
}

export async function dispatchNotification(input: {
	workflowId: string;
	templateId: string;
	variables: Record<string, string>;
	payload: Record<string, unknown>;
}): Promise<void> {
	const res = await fetch(`${API_URL}/notifications/dispatch`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(input),
	});
	if (!res.ok) throw new Error("errors.dispatchNotification");
}

export async function listQueue(): Promise<QueueItem[]> {
	const res = await fetch(`${API_URL}/queues/pending`);
	if (!res.ok) throw new Error("errors.loadQueue");
	return res.json();
}

export type TelegramToken = {
	id: string;
	name: string;
	secret: string;
	comment: string;
	isActive: boolean;
};

export async function listTelegramTokens(): Promise<TelegramToken[]> {
	const res = await fetch(`${API_URL}/connectors/telegram`);
	if (!res.ok) throw new Error("errors.loadTokens");
	const data = await res.json();
	return Array.isArray(data) ? data : [];
}

export async function createTelegramToken(
	payload: { name: string; secret: string; comment: string },
): Promise<TelegramToken> {
	const res = await fetch(`${API_URL}/connectors/telegram`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(payload),
	});
	if (!res.ok) throw new Error("errors.createToken");
	return res.json();
}

export async function updateTelegramToken(
	id: string,
	payload: { name: string; secret: string; comment: string },
): Promise<TelegramToken> {
	const res = await fetch(`${API_URL}/connectors/telegram/${id}`, {
		method: "PUT",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(payload),
	});
	if (!res.ok) throw new Error("errors.updateToken");
	return res.json();
}

export async function deleteTelegramToken(id: string): Promise<void> {
	const res = await fetch(`${API_URL}/connectors/telegram/${id}`, {
		method: "DELETE",
	});
	if (!res.ok) throw new Error("errors.deleteToken");
}

export async function toggleTelegramTokenActive(
	id: string,
	isActive: boolean,
): Promise<TelegramToken> {
	const res = await fetch(`${API_URL}/connectors/telegram/${id}/active`, {
		method: "PATCH",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ isActive }),
	});
	if (!res.ok) throw new Error("errors.toggleTokenStatus");
	return res.json();
}

export type SmtpAccount = {
	id: string;
	name: string;
	host: string;
	port: number;
	username: string;
	password: string;
	from: string;
	comment: string;
	useTls: boolean;
	useStartTls: boolean;
	isActive: boolean;
};

export async function listSmtpAccounts(): Promise<SmtpAccount[]> {
	const res = await fetch(`${API_URL}/connectors/smtp`);
	if (!res.ok) throw new Error("errors.loadSmtpAccounts");
	const data = await res.json();
	return Array.isArray(data) ? data : [];
}

export async function createSmtpAccount(
	payload: Omit<SmtpAccount, "id" | "isActive">,
): Promise<SmtpAccount> {
	const res = await fetch(`${API_URL}/connectors/smtp`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(payload),
	});
	if (!res.ok) throw new Error("errors.createSmtpAccount");
	return res.json();
}

export async function updateSmtpAccount(
	id: string,
	payload: Omit<SmtpAccount, "id" | "isActive">,
): Promise<SmtpAccount> {
	const res = await fetch(`${API_URL}/connectors/smtp/${id}`, {
		method: "PUT",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(payload),
	});
	if (!res.ok) throw new Error("errors.updateSmtpAccount");
	return res.json();
}

export async function deleteSmtpAccount(id: string): Promise<void> {
	const res = await fetch(`${API_URL}/connectors/smtp/${id}`, {
		method: "DELETE",
	});
	if (!res.ok) throw new Error("errors.deleteSmtpAccount");
}

export async function toggleSmtpAccountActive(
	id: string,
	isActive: boolean,
): Promise<SmtpAccount> {
	const res = await fetch(`${API_URL}/connectors/smtp/${id}/active`, {
		method: "PATCH",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ isActive }),
	});
	if (!res.ok) throw new Error("errors.toggleSmtpAccountStatus");
	return res.json();
}

export type Channel = {
	id: string;
	name: string;
	displayName?: string;
	description: string;
	muted: boolean;
};

export async function listChannels(connectorId: string): Promise<Channel[]> {
	const res = await fetch(`${API_URL}/connectors/${connectorId}/channels`);
	if (!res.ok) throw new Error("errors.loadChannels");
	const data = await res.json();
	return Array.isArray(data) ? data : [];
}

export async function createChannel(
	connectorId: string,
	payload: { name: string; displayName?: string; description: string; muted: boolean },
): Promise<Channel> {
	const res = await fetch(`${API_URL}/connectors/${connectorId}/channels`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(payload),
	});
	if (!res.ok) throw new Error("errors.createChannel");
	return res.json();
}

export async function updateChannel(
	id: string,
	payload: { name: string; displayName?: string; description: string; muted: boolean },
): Promise<Channel> {
	const res = await fetch(`${API_URL}/channels/${id}`, {
		method: "PUT",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(payload),
	});
	if (!res.ok) throw new Error("errors.updateChannel");
	return res.json();
}

export async function deleteChannel(id: string): Promise<void> {
	const res = await fetch(`${API_URL}/channels/${id}`, {
		method: "DELETE",
	});
	if (!res.ok) throw new Error("errors.deleteChannel");
}

export type StreamEvent = {
	event_id: string;
	event_type: string;
	occurred_at: string;
	context: Record<string, any>;
	metadata: Record<string, any>;
};

export async function getStreamMessages(eventTypes: string[] = [], limit: number = 10): Promise<StreamEvent[]> {
	const params = new URLSearchParams();
	if (eventTypes.length > 0) {
		eventTypes.forEach(et => params.append("eventTypes", et));
	}
	params.append("limit", limit.toString());
	
	const res = await fetch(`${API_URL}/stream/messages?${params.toString()}`);
	if (!res.ok) throw new Error("errors.loadMessages");
	return res.json();
}

import type { QueueItem } from "$lib/types/queue";
import type { TemplateDraft } from "$lib/types/template";
import type { WorkflowDraft } from "$lib/types/workflow";

const API_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080/api/v1";

export async function listTemplates(): Promise<TemplateDraft[]> {
	const res = await fetch(`${API_URL}/templates`);
	if (!res.ok) throw new Error("Не удалось загрузить шаблоны");
	return res.json();
}

export async function saveTemplate(
	payload: TemplateDraft,
): Promise<TemplateDraft> {
	const res = await fetch(`${API_URL}/templates`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(payload),
	});
	if (!res.ok) throw new Error("Не удалось сохранить шаблон");
	return res.json();
}

export async function listWorkflows(): Promise<WorkflowDraft[]> {
	const res = await fetch(`${API_URL}/workflows`);
	if (!res.ok) throw new Error("Не удалось загрузить workflow");
	const data = await res.json();
	return Array.isArray(data) ? data : [];
}

export async function getWorkflow(id: string): Promise<WorkflowDraft> {
	const res = await fetch(`${API_URL}/workflows/${id}`);
	if (!res.ok) throw new Error("Не удалось загрузить workflow");
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
	if (!res.ok) throw new Error("Не удалось сохранить workflow");
	return res.json();
}

export async function deleteWorkflow(id: string): Promise<void> {
	const res = await fetch(`${API_URL}/workflows/${id}`, {
		method: "DELETE",
	});
	if (!res.ok) throw new Error("Не удалось удалить workflow");
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
	if (!res.ok) throw new Error("Не удалось отправить уведомление");
}

export async function listQueue(): Promise<QueueItem[]> {
	const res = await fetch(`${API_URL}/queues/pending`);
	if (!res.ok) throw new Error("Не удалось загрузить очередь");
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
	if (!res.ok) throw new Error("Не удалось загрузить токены");
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
	if (!res.ok) throw new Error("Не удалось создать токен");
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
	if (!res.ok) throw new Error("Не удалось обновить токен");
	return res.json();
}

export async function deleteTelegramToken(id: string): Promise<void> {
	const res = await fetch(`${API_URL}/connectors/telegram/${id}`, {
		method: "DELETE",
	});
	if (!res.ok) throw new Error("Не удалось удалить токен");
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
	if (!res.ok) throw new Error("Не удалось изменить статус токена");
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
	if (!res.ok) throw new Error("Не удалось загрузить каналы");
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
	if (!res.ok) throw new Error("Не удалось создать канал");
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
	if (!res.ok) throw new Error("Не удалось обновить канал");
	return res.json();
}

export async function deleteChannel(id: string): Promise<void> {
	const res = await fetch(`${API_URL}/channels/${id}`, {
		method: "DELETE",
	});
	if (!res.ok) throw new Error("Не удалось удалить канал");
}

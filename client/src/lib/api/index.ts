import type { TemplateDraft } from '$lib/types/template';
import type { WorkflowDraft } from '$lib/types/workflow';
import type { QueueItem } from '$lib/types/queue';

const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080/v1';

export async function listTemplates(): Promise<TemplateDraft[]> {
  const res = await fetch(`${API_URL}/templates`);
  if (!res.ok) throw new Error('Не удалось загрузить шаблоны');
  return res.json();
}

export async function saveTemplate(payload: TemplateDraft): Promise<TemplateDraft> {
  const res = await fetch(`${API_URL}/templates`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  if (!res.ok) throw new Error('Не удалось сохранить шаблон');
  return res.json();
}

export async function listWorkflows(): Promise<WorkflowDraft[]> {
  const res = await fetch(`${API_URL}/workflows`);
  if (!res.ok) throw new Error('Не удалось загрузить workflow');
  return res.json();
}

export async function saveWorkflow(payload: WorkflowDraft): Promise<WorkflowDraft> {
  const res = await fetch(`${API_URL}/workflows`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  if (!res.ok) throw new Error('Не удалось сохранить workflow');
  return res.json();
}

export async function dispatchNotification(input: {
  workflowId: string;
  templateId: string;
  variables: Record<string, string>;
  payload: Record<string, unknown>;
}): Promise<void> {
  const res = await fetch(`${API_URL}/notifications/dispatch`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(input)
  });
  if (!res.ok) throw new Error('Не удалось отправить уведомление');
}

export async function listQueue(): Promise<QueueItem[]> {
  const res = await fetch(`${API_URL}/queues/pending`);
  if (!res.ok) throw new Error('Не удалось загрузить очередь');
  return res.json();
}


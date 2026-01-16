import { writable } from "svelte/store";
import type { TemplateDraft } from "$lib/types/template";

export const templateStore = writable<TemplateDraft[]>([
	{
		id: "welcome-msg",
		name: "Приветственное сообщение",
		description: "Уведомление о запуске нового workflow",
		body: "Привет! Workflow {{workflowId}} активирован.",
		variables: {
			workflowId: "Идентификатор workflow",
		},
	},
]);

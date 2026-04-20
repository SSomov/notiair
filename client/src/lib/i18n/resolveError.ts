import { get } from "svelte/store";
import { t } from "$lib/i18n";

/** Translates i18n error keys (`errors.*`). Passes through other messages (e.g. network). */
export function resolveI18nError(message: string): string {
	if (message.startsWith("errors.")) {
		return get(t)(message);
	}
	return message;
}

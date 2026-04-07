import type { Handle } from "@sveltejs/kit";
import { base } from "$app/paths";

export const handle: Handle = async ({ event, resolve }) => {
	const pathname = event.url.pathname;
	const appPath =
		base && pathname.startsWith(base) ? pathname.slice(base.length) || "/" : pathname;

	// Если путь начинается с /ru, пропускаем как есть
	if (appPath === "/ru" || appPath.startsWith("/ru/")) {
		return resolve(event);
	}

	// Английский язык без префикса /ru
	return resolve(event);
};

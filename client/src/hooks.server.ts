import type { Handle } from "@sveltejs/kit";

export const handle: Handle = async ({ event, resolve }) => {
	const { url } = event;
	const pathname = url.pathname;

	// Если путь начинается с /ru, пропускаем как есть
	if (pathname.startsWith("/ru")) {
		return resolve(event);
	}

	// Если путь не начинается с /ru и не корневой, пропускаем
	// (английский язык без префикса)
	return resolve(event);
};

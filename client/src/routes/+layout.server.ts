import { loadTranslations } from "$lib/i18n";
import { getLocaleFromPath, removeLocaleFromPath } from "$lib/i18n/utils";

export async function load({ url, cookies }) {
	const pathname = url.pathname;

	// Определяем локаль из URL
	const currentLocale = getLocaleFromPath(pathname);
	const pathWithoutLocale = removeLocaleFromPath(pathname);

	// Сохраняем локаль в cookie
	cookies.set("locale", currentLocale, {
		path: "/",
		maxAge: 60 * 60 * 24 * 365,
	});

	await loadTranslations(currentLocale, pathWithoutLocale);

	return {
		locale: currentLocale,
		pathWithoutLocale,
	};
}

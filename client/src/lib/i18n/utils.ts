import { goto } from "$app/navigation";

/**
 * Получает локаль из URL
 */
export function getLocaleFromPath(pathname: string): "en" | "ru" {
	if (pathname.startsWith("/ru")) {
		return "ru";
	}
	return "en";
}

/**
 * Удаляет префикс локали из пути
 */
export function removeLocaleFromPath(pathname: string): string {
	if (pathname.startsWith("/ru")) {
		return pathname.slice(3) || "/";
	}
	return pathname;
}

/**
 * Добавляет префикс локали к пути
 */
export function addLocaleToPath(path: string, locale: "en" | "ru"): string {
	const cleanPath = path.startsWith("/") ? path : `/${path}`;

	if (locale === "ru") {
		return `/ru${cleanPath === "/" ? "" : cleanPath}`;
	}

	// Для английского убираем префикс, если он есть
	return cleanPath;
}

/**
 * Переключает язык и перенаправляет на правильный URL
 */
export async function switchLocale(
	currentPath: string,
	newLocale: "en" | "ru",
	_currentLocale: "en" | "ru",
) {
	const pathWithoutLocale = removeLocaleFromPath(currentPath);
	const newPath = addLocaleToPath(pathWithoutLocale, newLocale);

	await goto(newPath, { invalidateAll: true });
}

import { base } from "$app/paths";
import { goto } from "$app/navigation";

function withoutAppBase(pathname: string): string {
	if (base && pathname.startsWith(base)) {
		return pathname.slice(base.length) || "/";
	}
	return pathname;
}

/**
 * Получает локаль из URL
 */
export function getLocaleFromPath(pathname: string): "en" | "ru" {
	const p = withoutAppBase(pathname);
	if (p === "/ru" || p.startsWith("/ru/")) {
		return "ru";
	}
	return "en";
}

/**
 * Удаляет префикс локали из пути
 */
export function removeLocaleFromPath(pathname: string): string {
	const p = withoutAppBase(pathname);
	if (p.startsWith("/ru")) {
		return p.slice(3) || "/";
	}
	return p;
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

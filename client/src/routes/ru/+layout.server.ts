import { loadTranslations } from '$lib/i18n';
import { removeLocaleFromPath } from '$lib/i18n/utils';

export async function load({ url, cookies }) {
	const pathname = url.pathname;
	const pathWithoutLocale = removeLocaleFromPath(pathname);
	
	// Сохраняем локаль в cookie
	cookies.set('locale', 'ru', { path: '/', maxAge: 60 * 60 * 24 * 365 });

	await loadTranslations('ru', pathWithoutLocale);

	return {
		locale: 'ru',
		pathWithoutLocale
	};
}
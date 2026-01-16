import { loadTranslations, locale } from "$lib/i18n";
import { getLocaleFromPath, removeLocaleFromPath } from "$lib/i18n/utils";
import { browser } from "$app/environment";
import { get } from "svelte/store";

export async function load({ url, data, depends }) {
	if (browser) {
		depends("app:locale");
		const currentLocale = data?.locale || getLocaleFromPath(url.pathname) || "en";
		const pathWithoutLocale = removeLocaleFromPath(url.pathname);
		
		if (get(locale) !== currentLocale) {
			locale.set(currentLocale);
		}
		
		await loadTranslations(currentLocale, pathWithoutLocale);
	}

	return {
		...data,
	};
}
<script lang="ts">
import "../app.css";
import { t, locale } from "$lib/i18n";
import { page } from "$app/stores";
import { switchLocale } from "$lib/i18n/utils";
import type { PageData } from "./$types";

export let data: PageData;

$: currentLocale = data.locale || "en";

async function switchLanguage() {
	const newLocale = currentLocale === "ru" ? "en" : "ru";

	locale.set(newLocale);
	await switchLocale($page.url.pathname, newLocale, currentLocale);
}
</script>

<header class="sticky top-0 z-40 w-full border-b border-border bg-surface/95 shadow-sm backdrop-blur">
  <div class="mx-auto flex w-full max-w-7xl items-center justify-between px-4 py-4 md:px-10">
    <a href={currentLocale === 'ru' ? '/ru/' : '/'} class="flex items-center gap-4 transition hover:opacity-80">
      <span class="inline-flex h-11 w-11 items-center justify-center rounded-xl bg-accent/10 text-2xl text-accent">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="h-6 w-6">
          <path d="M13.5 2a1 1 0 0 1 .894.553l6 12A1 1 0 0 1 19.5 16h-5.382l.618 5.543a1 1 0 0 1-1.872.654l-9-14A1 1 0 0 1 4.75 6H10.5l-1.118-3.276A1 1 0 0 1 10.341 2H13.5Z" />
        </svg>
      </span>
      <div>
        <p class="text-xs font-semibold uppercase tracking-[0.45em] text-accent">{$t('header.title')}</p>
        <p class="text-xs uppercase tracking-[0.3em] text-muted">{$t('header.subtitle')}</p>
      </div>
    </a>

    <div class="flex items-center gap-3">
      <a href={currentLocale === 'ru' ? '/ru/connectors' : '/connectors'} class="btn-primary hidden md:inline-flex">{$t('common.connectors')}</a>
      <a href={currentLocale === 'ru' ? '/ru/channels' : '/channels'} class="btn-primary hidden md:inline-flex">{$t('common.channels')}</a>
      <a href={currentLocale === 'ru' ? '/ru/workflows' : '/workflows'} class="btn-primary hidden md:inline-flex">{$t('common.workflows')}</a>
      <span class="pill hidden md:inline-flex">{$t('common.beta')}</span>
      <span class="hidden text-sm text-muted md:inline">redis · postgres · telegram</span>
      <button
        type="button"
        class="inline-flex h-8 w-8 items-center justify-center rounded-full border border-border bg-surface text-muted transition hover:text-accent"
        onclick={switchLanguage}
        aria-label="Switch language"
        title="Switch language"
      >
        <span class="text-xs font-semibold">{currentLocale === 'ru' ? 'EN' : 'RU'}</span>
      </button>
      <button
        type="button"
        class="inline-flex h-10 w-10 items-center justify-center rounded-full border border-border bg-surface text-muted transition hover:text-accent"
        aria-label="{$t('header.openSettings')}"
      >
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="h-5 w-5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M10.343 3.94c.09-.542.56-.94 1.11-.94h1.093c.55 0 1.02.398 1.11.94l.149.894c.07.424.384.764.78.93.398.164.855.142 1.205-.108l.737-.527a1.125 1.125 0 0 1 1.45.12l.773.774c.39.389.44 1.002.12 1.45l-.527.737c-.25.35-.272.806-.107 1.204.165.397.505.71.93.78l.893.15c.543.09.94.56.94 1.109v1.094c0 .55-.397 1.02-.94 1.11l-.893.149c-.425.07-.765.383-.93.78-.165.398-.143.854.107 1.204l.527.738c.32.447.269 1.06-.12 1.45l-.774.773a1.125 1.125 0 0 1-1.449.12l-.738-.527c-.35-.25-.806-.272-1.203-.107-.397.165-.71.505-.781.929l-.149.894c-.09.542-.56.94-1.11.94h-1.094c-.55 0-1.019-.398-1.11-.94l-.148-.894c-.071-.424-.384-.764-.781-.93-.398-.164-.854-.142-1.204.108l-.738.527c-.447.32-1.06.269-1.45-.12l-.773-.774a1.125 1.125 0 0 1-.12-1.45l.527-.737c.25-.35.273-.806.108-1.204-.165-.397-.505-.71-.93-.78l-.894-.15c-.542-.09-.94-.56-.94-1.109v-1.094c0-.55.398-1.02.94-1.11l.894-.149c.424-.07.765-.383.93-.78.165-.398.142-.854-.108-1.204l-.527-.738a1.125 1.125 0 0 1 .12-1.45l.773-.773c.389-.39 1.002-.44 1.45-.12l.737.527c.35.25.807.272 1.204.107.397-.165.71-.505.78-.929l.15-.894Z" />
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
        </svg>
      </button>
    </div>
  </div>
</header>

<main class="px-4 pb-12 pt-6 md:px-12 md:pt-10">
  <div class="mx-auto w-full max-w-6xl space-y-8">
    <slot />
  </div>
</main>


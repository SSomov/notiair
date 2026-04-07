<script lang="ts">
	import { t } from "$lib/i18n";

	let generatedKey: string | null = null;
	let copied = false;

	function generateKey() {
		const bytes = new Uint8Array(16);
		crypto.getRandomValues(bytes);
		const hex = Array.from(bytes)
			.map((b) => b.toString(16).padStart(2, "0"))
			.join("");
		generatedKey = `nak_${hex}`;
		copied = false;
	}

	async function copyKey() {
		if (!generatedKey) return;
		try {
			await navigator.clipboard.writeText(generatedKey);
			copied = true;
			setTimeout(() => (copied = false), 2000);
		} catch {
			/* ignore */
		}
	}
</script>

<section class="space-y-8 px-4 pb-12 pt-2 md:px-12 md:pt-4">
	<header class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
		<div class="space-y-2">
			<span class="pill">{$t('header.settingsApiKeys')}</span>
			<h1 class="text-3xl font-semibold">{$t('apiKeys.title')}</h1>
			<p class="text-sm text-muted max-w-2xl">{$t('apiKeys.description')}</p>
		</div>
	</header>

	<article class="glass-card max-w-2xl space-y-6">
		<div class="flex flex-col gap-4 sm:flex-row sm:items-end">
			<div class="flex-1">
				<label for="api-key-input" class="mb-2 block text-sm font-medium text-muted">
					{$t('apiKeys.title')}
				</label>
				{#if generatedKey}
					<div class="flex gap-2">
						<input
							id="api-key-input"
							type="text"
							readonly
							value={generatedKey}
							class="flex-1 rounded-xl border border-border bg-surfaceMuted/60 px-4 py-2.5 font-mono text-sm text-text"
						/>
						<button
							type="button"
							class="btn-primary shrink-0"
							onclick={copyKey}
						>
							{copied ? $t('apiKeys.copied') : $t('apiKeys.copy')}
						</button>
					</div>
					<p class="mt-2 text-xs text-amber-600 dark:text-amber-400">
						{$t('apiKeys.warning')}
					</p>
				{:else}
					<p class="rounded-xl border border-dashed border-border bg-surfaceMuted/40 px-4 py-6 text-center text-sm text-muted">
						{$t('apiKeys.placeholder')}
					</p>
				{/if}
			</div>
			<button type="button" class="btn-primary" onclick={generateKey}>
				{$t('apiKeys.generate')}
			</button>
		</div>
	</article>
</section>

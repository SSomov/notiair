<script lang="ts">
	import { templateStore } from '@stores/templates';
	import { t } from '$lib/i18n';

	const templates = templateStore;
</script>

<section class="space-y-8 px-4 pb-12 pt-2 md:px-12 md:pt-4">
	<header class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
		<div class="space-y-2">
			<span class="pill">{$t('common.templates')}</span>
			<div>
				<h1 class="text-3xl font-semibold">{$t('templates.title')}</h1>
				<p class="text-sm text-muted">
					{$t('templates.description')}
				</p>
			</div>
		</div>
		<button class="btn-primary">{$t('templates.newTemplate')}</button>
	</header>

	<div class="grid gap-4">
		{#each $templates as template}
			<article class="glass-card">
				<div class="flex items-start justify-between gap-3">
					<div>
						<h2 class="text-lg font-semibold text-text">{template.name}</h2>
						<p class="text-sm text-muted">{template.description}</p>
					</div>
					<span class="pill">id: {template.id}</span>
				</div>
				<pre class="mt-4 rounded-xl bg-surfaceMuted/60 p-4 text-sm text-text">
{template.body}
        </pre>
			</article>
		{/each}
	</div>
</section>

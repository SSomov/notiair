<script lang="ts">
import { t } from "$lib/i18n";
import { workflowStore } from "@stores/workflows";

const draftBlueprints = [
	{
		id: "draft-approval-flow",
		name: "Approval flow v2",
		hint: "Пауза + ручное подтверждение перед отправкой",
		updated: "2 часа назад",
	},
	{
		id: "draft-personalized",
		name: "Personalized onboarding",
		hint: "Сегменты по гео и активности",
		updated: "вчера",
	},
];

$: activeWorkflows = $workflowStore;
</script>

<section class="space-y-8 px-4 pb-12 pt-2 md:px-12 md:pt-4">
	<header class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
		<div class="space-y-2">
			<span class="pill">{$t('common.workflows')}</span>
			<p class="text-sm text-muted max-w-2xl">
				{$t('workflows.description')}
			</p>
		</div>
		<div class="flex flex-wrap gap-3">
			<a href="/templates" class="btn-primary">{$t('workflows.goToTemplates')}</a>
			<a
				href="/workflows/new"
				class="btn-primary bg-surfaceMuted text-text shadow-none hover:shadow-sm">{$t('workflows.newWorkflow')}</a
			>
		</div>
	</header>

	<div class="grid gap-8 lg:grid-cols-2">
		<div class="space-y-4">
			<div class="flex items-center justify-between">
				<div>
					<h2 class="text-lg font-semibold">{$t('workflows.activeScenarios.title')}</h2>
					<p class="text-xs uppercase tracking-wide text-muted">
						{$t('workflows.activeScenarios.subtitle')} • {activeWorkflows.length}
					</p>
				</div>
			</div>
			<div class="grid gap-4 sm:grid-cols-2">
				{#each activeWorkflows as workflow}
					<article class="glass-card h-full">
						<div class="flex items-start justify-between gap-3">
							<div class="space-y-1">
								<h3 class="text-base font-semibold text-text">{workflow.name}</h3>
								<p class="text-xs uppercase tracking-wide text-muted">ID: {workflow.id}</p>
							</div>
							<div class="flex items-center gap-2">
								<button
									type="button"
									class="inline-flex h-8 w-8 items-center justify-center rounded-full border border-border text-muted transition hover:text-accent"
									aria-label="{$t('workflows.activeScenarios.copy')}"
									title="{$t('workflows.activeScenarios.copy')}"
								>
									<svg
										xmlns="http://www.w3.org/2000/svg"
										viewBox="0 0 24 24"
										fill="none"
										stroke="currentColor"
										stroke-width="1.5"
										class="h-4 w-4"
									>
										<rect x="9" y="9" width="12" height="12" rx="2" ry="2" />
										<path d="M5 15H4a2 2 0 0 1-2-2V4c0-1.1.9-2 2-2h9a2 2 0 0 1 2 2v1" />
									</svg>
								</button>
								<span
									class="inline-flex rounded-full bg-positive/10 px-3 py-1 text-xs font-medium text-positive"
									>{$t('common.active')}</span
								>
							</div>
						</div>
						<p class="mt-3 text-sm text-muted">
							{$t('workflows.activeScenarios.description')}
						</p>
						<div class="mt-4 flex items-center justify-between">
							<button
								type="button"
								class="inline-flex text-sm font-semibold text-accent hover:underline"
								>{$t('workflows.activeScenarios.openBuilder')}</button
							>
						</div>
					</article>
				{/each}
			</div>
		</div>

		<div class="space-y-4">
			<div class="flex items-center justify-between">
				<div>
					<h2 class="text-lg font-semibold">{$t('workflows.drafts.title')}</h2>
					<p class="text-xs uppercase tracking-wide text-muted">{$t('workflows.drafts.subtitle')}</p>
				</div>
				<button type="button" class="text-sm font-semibold text-accent hover:underline"
					>{$t('workflows.drafts.allDrafts')}</button
				>
			</div>
			<div class="grid gap-4 sm:grid-cols-2">
				{#each draftBlueprints as draft}
					<article class="glass-card h-full">
						<div class="flex items-start justify-between gap-3">
							<div class="space-y-2">
								<h3 class="text-base font-semibold text-text">{draft.name}</h3>
								<p class="text-sm text-muted">{draft.hint}</p>
								<p class="text-xs text-muted">{$t('workflows.drafts.updated')} {draft.updated}</p>
							</div>
							<button
								type="button"
								class="inline-flex h-8 w-8 items-center justify-center rounded-full border border-border text-muted transition hover:text-accent"
								aria-label="{$t('workflows.drafts.copy')}"
								title="{$t('workflows.drafts.copy')}"
							>
								<svg
									xmlns="http://www.w3.org/2000/svg"
									viewBox="0 0 24 24"
									fill="none"
									stroke="currentColor"
									stroke-width="1.5"
									class="h-4 w-4"
								>
									<rect x="9" y="9" width="12" height="12" rx="2" ry="2" />
									<path d="M5 15H4a2 2 0 0 1-2-2V4c0-1.1.9-2 2-2h9a2 2 0 0 1 2 2v1" />
								</svg>
							</button>
						</div>
						<button
							type="button"
							class="mt-4 inline-flex text-sm font-semibold text-accent hover:underline"
							>{$t('workflows.drafts.openDraft')}</button
						>
					</article>
				{/each}
				<article
					class="glass-card flex h-full items-center justify-center border-dashed border-border text-muted"
				>
					<div class="text-center">
						<p class="font-semibold">{$t('workflows.drafts.createDraft')}</p>
						<p class="text-sm text-muted">{$t('workflows.drafts.createDraftDescription')}</p>
					</div>
				</article>
			</div>
		</div>
	</div>
</section>

<script lang="ts">
	import { onMount } from "svelte";
	import { t } from "$lib/i18n";
	import { listWorkflows, deleteWorkflow } from "$lib/api";
	import type { WorkflowDraft } from "$lib/types/workflow";

	let workflows: WorkflowDraft[] = [];
	let loading = true;
	let error: string | null = null;
	let deletingId: string | null = null;

	onMount(async () => {
		try {
			workflows = await listWorkflows();
		} catch (e) {
			error = e instanceof Error ? e.message : "Не удалось загрузить workflow";
			workflows = [];
		} finally {
			loading = false;
		}
	});

	async function handleDelete(id: string) {
		if (!confirm("Вы уверены, что хотите удалить этот черновик?")) {
			return;
		}

		deletingId = id;
		try {
			await deleteWorkflow(id);
			workflows = workflows.filter((w) => w.id !== id);
		} catch (e) {
			error = e instanceof Error ? e.message : "Не удалось удалить workflow";
		} finally {
			deletingId = null;
		}
	}

	$: activeWorkflows = workflows.filter((w) => w.isActive === true);
	$: draftWorkflows = workflows.filter((w) => w.isActive !== true);

	function formatUpdatedDate(date: string | undefined): string {
		if (!date) return "";
		const d = new Date(date);
		const now = new Date();
		const diffMs = now.getTime() - d.getTime();
		const diffMins = Math.floor(diffMs / 60000);
		const diffHours = Math.floor(diffMs / 3600000);
		const diffDays = Math.floor(diffMs / 86400000);

		if (diffMins < 60) {
			return `${diffMins} ${diffMins === 1 ? "минуту" : diffMins < 5 ? "минуты" : "минут"} назад`;
		}
		if (diffHours < 24) {
			return `${diffHours} ${diffHours === 1 ? "час" : diffHours < 5 ? "часа" : "часов"} назад`;
		}
		if (diffDays === 1) {
			return "вчера";
		}
		if (diffDays < 7) {
			return `${diffDays} ${diffDays < 5 ? "дня" : "дней"} назад`;
		}
		return d.toLocaleDateString("ru-RU");
	}
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
				{#if loading}
					<p class="text-sm text-muted">Загрузка...</p>
				{:else if error}
					<p class="text-sm text-negative">{error}</p>
				{:else if activeWorkflows.length === 0}
					<p class="text-sm text-muted">Нет активных workflow</p>
				{:else}
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
							<a
								href="/workflows/edit?id={workflow.id}"
								class="inline-flex text-sm font-semibold text-accent hover:underline"
								>{$t('workflows.activeScenarios.openBuilder')}</a
							>
							</div>
						</article>
					{/each}
				{/if}
			</div>
		</div>

		<div class="space-y-4">
			<div class="flex items-center justify-between">
				<div>
					<h2 class="text-lg font-semibold">{$t('workflows.drafts.title')}</h2>
					<p class="text-xs uppercase tracking-wide text-muted">
						{$t('workflows.drafts.subtitle')} • {draftWorkflows.length}
					</p>
				</div>
				{#if draftWorkflows.length > 0}
					<button type="button" class="text-sm font-semibold text-accent hover:underline"
						>{$t('workflows.drafts.allDrafts')}</button
					>
				{/if}
			</div>
			<div class="grid gap-4 sm:grid-cols-2">
				{#if loading}
					<p class="text-sm text-muted">Загрузка...</p>
				{:else if error}
					<p class="text-sm text-negative">{error}</p>
				{:else if draftWorkflows.length === 0}
					<article
						class="glass-card flex h-full items-center justify-center border-dashed border-border text-muted"
					>
						<div class="text-center">
							<p class="font-semibold">{$t('workflows.drafts.createDraft')}</p>
							<p class="text-sm text-muted">{$t('workflows.drafts.createDraftDescription')}</p>
						</div>
					</article>
				{:else}
					{#each draftWorkflows as draft}
						<article class="glass-card h-full">
							<div class="flex items-start justify-between gap-3">
								<div class="space-y-2">
									<h3 class="text-base font-semibold text-text">{draft.name || "Без названия"}</h3>
									{#if draft.description}
										<p class="text-sm text-muted">{draft.description}</p>
									{/if}
									<p class="text-xs text-muted">
										{$t('workflows.drafts.updated')} {formatUpdatedDate(draft.updatedAt)}
									</p>
								</div>
								<button
									type="button"
									class="inline-flex h-8 w-8 items-center justify-center rounded-full border border-border text-muted transition hover:text-negative disabled:opacity-50"
									aria-label="{$t('workflows.drafts.delete')}"
									title="{$t('workflows.drafts.delete')}"
									on:click={() => handleDelete(draft.id)}
									disabled={deletingId === draft.id}
								>
									<svg
										xmlns="http://www.w3.org/2000/svg"
										viewBox="0 0 24 24"
										fill="none"
										stroke="currentColor"
										stroke-width="1.5"
										class="h-4 w-4"
									>
										<path
											d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
										/>
									</svg>
								</button>
							</div>
							<a
								href="/workflows/edit?id={draft.id}"
								class="mt-4 inline-flex text-sm font-semibold text-accent hover:underline"
								>{$t('workflows.drafts.openDraft')}</a
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
				{/if}
			</div>
		</div>
	</div>
</section>

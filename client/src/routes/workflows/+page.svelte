<script lang="ts">
	import { onMount } from "svelte";
	import { page } from "$app/stores";
	import { resolve } from "$app/paths";
	import { locale, t } from "$lib/i18n";
	import { listWorkflows, deleteWorkflow } from "$lib/api";
	import { getLocaleFromPath, addLocaleToPath } from "$lib/i18n/utils";
	import { resolveI18nError } from "$lib/i18n/resolveError";
	import type { WorkflowDraft } from "$lib/types/workflow";
	import { get } from "svelte/store";

	$: loc = getLocaleFromPath($page.url.pathname);
	$: hrefNew = resolve(addLocaleToPath("/workflows/new", loc));
	const hrefEdit = (id: string) =>
		resolve(addLocaleToPath(`/workflows/edit?id=${encodeURIComponent(id)}`, loc));

	let workflows: WorkflowDraft[] = [];
	let loading = true;
	let error: string | null = null;
	let deletingId: string | null = null;
	let errorDisplay: string | null = null;

	onMount(async () => {
		try {
			workflows = await listWorkflows();
		} catch (e) {
			error = e instanceof Error ? e.message : "errors.loadWorkflows";
			workflows = [];
		} finally {
			loading = false;
		}
	});

	async function handleDelete(id: string) {
		if (!confirm(get(t)("workflows.confirmDeleteDraft"))) {
			return;
		}

		deletingId = id;
		try {
			await deleteWorkflow(id);
			workflows = workflows.filter((w) => w.id !== id);
		} catch (e) {
			error = e instanceof Error ? e.message : "errors.deleteWorkflow";
		} finally {
			deletingId = null;
		}
	}

	$: activeWorkflows = workflows.filter((w) => w.isActive === true);
	$: draftWorkflows = workflows.filter((w) => w.isActive !== true);

	function formatUpdatedDate(date: string | undefined, localeCode: string): string {
		if (!date) return "";
		const d = new Date(date);
		const now = new Date();
		const diffMs = now.getTime() - d.getTime();
		const diffMins = Math.floor(diffMs / 60000);
		const diffHours = Math.floor(diffMs / 3600000);
		const diffDays = Math.floor(diffMs / 86400000);
		const locTag = localeCode === "ru" ? "ru-RU" : "en-US";
		const rtf = new Intl.RelativeTimeFormat(locTag, { numeric: "auto" });
		if (diffMins < 60) {
			return rtf.format(-diffMins, "minute");
		}
		if (diffHours < 24) {
			return rtf.format(-diffHours, "hour");
		}
		if (diffDays < 7) {
			return rtf.format(-diffDays, "day");
		}
		return d.toLocaleDateString(locTag);
	}

	$: {
		$locale;
		errorDisplay = error ? resolveI18nError(error) : null;
	}

	async function handleCopyWorkflowId(id: string) {
		try {
			await navigator.clipboard.writeText(id);
		} catch {
			// ignore — clipboard may be unavailable
		}
	}
</script>

<section class="space-y-8 px-4 pb-12 pt-2 md:px-12 md:pt-4">
	<header class="space-y-2">
		<span class="pill">{$t('common.workflows')}</span>
		<p class="text-sm text-muted max-w-2xl">
			{$t('workflows.description')}
		</p>
	</header>

	<div class="grid gap-8 lg:grid-cols-2">
		<div class="space-y-4">
			<div class="flex items-center justify-between">
				<div>
					<h2 class="text-xl font-bold tracking-tight text-text">
						{$t('workflows.activeScenarios.title')}
					</h2>
					<p class="mt-1 text-xs uppercase tracking-wide text-muted">
						{$t('workflows.activeScenarios.subtitle')} • {activeWorkflows.length}
					</p>
				</div>
			</div>
			<div class="flex flex-col gap-4">
				{#if loading}
					<p class="text-sm text-muted">{$t('workflows.list.loading')}</p>
				{:else if error}
					<p class="text-sm text-negative">{errorDisplay}</p>
				{:else if activeWorkflows.length === 0}
					<p class="text-sm text-muted">{$t('workflows.list.noActive')}</p>
				{:else}
					{#each activeWorkflows as workflow}
						<article
							class="rounded-3xl border border-border bg-surface p-6 shadow-glass md:p-7"
						>
							<div class="space-y-4">
								<div class="flex items-start justify-between gap-4">
									<h3 class="min-w-0 text-base font-semibold text-text md:text-lg">
										{workflow.name}
									</h3>
									<div class="flex shrink-0 items-center gap-2">
										<button
											type="button"
											class="inline-flex h-9 w-9 items-center justify-center rounded-full border border-border bg-surface text-muted transition hover:border-accent/40 hover:text-accent"
											aria-label="{$t('workflows.activeScenarios.copy')}"
											title="{$t('workflows.activeScenarios.copy')}"
											on:click={() => handleCopyWorkflowId(workflow.id)}
										>
											<svg
												xmlns="http://www.w3.org/2000/svg"
												viewBox="0 0 24 24"
												fill="none"
												stroke="currentColor"
												stroke-width="1.5"
												class="h-4 w-4"
												aria-hidden="true"
											>
												<rect x="9" y="9" width="12" height="12" rx="2" ry="2" />
												<path
													d="M5 15H4a2 2 0 0 1-2-2V4c0-1.1.9-2 2-2h9a2 2 0 0 1 2 2v1"
												/>
											</svg>
										</button>
										<span
											class="inline-flex rounded-full bg-positive/12 px-3.5 py-1.5 text-xs font-semibold text-positive"
										>
											{$t('common.active')}
										</span>
									</div>
								</div>
								<p class="break-all text-xs leading-relaxed text-muted">
									ID: {workflow.id}
								</p>
								<p class="text-sm leading-relaxed text-muted">
									{$t('workflows.activeScenarios.description')}
								</p>
								<a
									href={hrefEdit(workflow.id)}
									class="inline-flex text-sm font-semibold text-accent hover:underline"
								>
									{$t('workflows.activeScenarios.openBuilder')}
								</a>
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
					<p class="text-sm text-muted">{$t('workflows.list.loading')}</p>
				{:else if error}
					<p class="text-sm text-negative">{errorDisplay}</p>
				{:else if draftWorkflows.length === 0}
					<a
						href={hrefNew}
						class="glass-card flex h-full min-h-[8rem] cursor-pointer items-center justify-center border-dashed border-border text-muted no-underline transition hover:border-accent/40 hover:text-text focus:outline-none focus-visible:ring-2 focus-visible:ring-accent"
					>
						<div class="text-center">
							<p class="font-semibold">{$t('workflows.drafts.createDraft')}</p>
							<p class="text-sm text-muted">{$t('workflows.drafts.createDraftDescription')}</p>
						</div>
					</a>
				{:else}
					{#each draftWorkflows as draft}
						<article class="glass-card h-full">
							<div class="flex items-start justify-between gap-3">
								<div class="space-y-2">
									<h3 class="text-base font-semibold text-text">{draft.name || $t('common.noName')}</h3>
									{#if draft.description}
										<p class="text-sm text-muted">{draft.description}</p>
									{/if}
									<p class="text-xs text-muted">
										{$t('workflows.drafts.updated')} {formatUpdatedDate(draft.updatedAt, loc)}
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
								href={hrefEdit(draft.id)}
								class="mt-4 inline-flex text-sm font-semibold text-accent hover:underline"
								>{$t('workflows.drafts.openDraft')}</a
							>
						</article>
					{/each}
					<a
						href={hrefNew}
						class="glass-card flex h-full min-h-[8rem] cursor-pointer items-center justify-center border-dashed border-border text-muted no-underline transition hover:border-accent/40 hover:text-text focus:outline-none focus-visible:ring-2 focus-visible:ring-accent"
					>
						<div class="text-center">
							<p class="font-semibold">{$t('workflows.drafts.createDraft')}</p>
							<p class="text-sm text-muted">{$t('workflows.drafts.createDraftDescription')}</p>
						</div>
					</a>
				{/if}
			</div>
		</div>
	</div>
</section>

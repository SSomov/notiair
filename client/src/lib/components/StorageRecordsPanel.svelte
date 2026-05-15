<script lang="ts">
	import { createEventDispatcher, onDestroy, onMount } from 'svelte';
	import { t } from '$lib/i18n';
	import {
		listStorageRecords,
		getStorageRecord,
		deleteStorageRecord,
		type StorageRecordListItem,
	} from '$lib/api';

	export let workflowId: string;
	export let nodeId: string;
	export let nodeLabel: string;

	const dispatch = createEventDispatcher<{
		close: void;
		labelChange: string;
		error: string;
	}>();

	const pageSize = 20;

	let editingLabel = false;
	let labelDraft = nodeLabel;

	let searchQuery = '';
	let searchDebounced = '';
	let searchTimer: ReturnType<typeof setTimeout> | null = null;

	let page = 1;
	let total = 0;
	let items: StorageRecordListItem[] = [];
	let loading = false;

	let selectedRecordId: string | null = null;
	let previewContent = '';
	let previewLoading = false;

	$: pageCount = Math.max(1, Math.ceil(total / pageSize));
	$: page = Math.min(page, pageCount);
	$: rangeFrom = total === 0 ? 0 : (page - 1) * pageSize + 1;
	$: rangeTo = total === 0 ? 0 : Math.min(page * pageSize, total);
	$: canPrev = page > 1;
	$: canNext = page < pageCount;

	function formatStorageSize(bytes: number): string {
		if (bytes < 1024) return `${bytes} B`;
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
	}

	async function loadRecords() {
		loading = true;
		try {
			const result = await listStorageRecords(workflowId, nodeId, {
				limit: pageSize,
				offset: (page - 1) * pageSize,
				q: searchDebounced,
			});
			items = result.items;
			total = result.total;
			const maxPage = Math.max(1, Math.ceil(total / pageSize));
			if (page > maxPage) {
				page = maxPage;
				loading = false;
				await loadRecords();
				return;
			}
			if (selectedRecordId && !items.some((r) => r.id === selectedRecordId)) {
				selectedRecordId = null;
				previewContent = '';
			}
		} catch (e) {
			dispatch('error', e instanceof Error ? e.message : 'errors.loadStorageRecords');
			items = [];
			total = 0;
		} finally {
			loading = false;
		}
	}

	async function selectRecord(recordId: string) {
		if (selectedRecordId === recordId && previewContent) return;
		selectedRecordId = recordId;
		previewLoading = true;
		previewContent = '';
		try {
			const rec = await getStorageRecord(workflowId, recordId);
			previewContent = rec.data;
		} catch (e) {
			dispatch('error', e instanceof Error ? e.message : 'errors.loadStorageRecord');
			previewContent = '';
		} finally {
			previewLoading = false;
		}
	}

	async function downloadRecord(record: StorageRecordListItem, event: MouseEvent) {
		event.stopPropagation();
		try {
			const rec = await getStorageRecord(workflowId, record.id);
			const blob = new Blob([rec.data], {
				type: rec.contentType || 'application/octet-stream',
			});
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = `storage-${record.id}.txt`;
			a.click();
			URL.revokeObjectURL(url);
		} catch (e) {
			dispatch('error', e instanceof Error ? e.message : 'errors.loadStorageRecord');
		}
	}

	async function removeRecord(recordId: string, event: MouseEvent) {
		event.stopPropagation();
		try {
			await deleteStorageRecord(workflowId, recordId);
			if (selectedRecordId === recordId) {
				selectedRecordId = null;
				previewContent = '';
			}
			await loadRecords();
		} catch (e) {
			dispatch('error', e instanceof Error ? e.message : 'errors.deleteStorageRecord');
		}
	}

	function toggleEditLabel() {
		editingLabel = !editingLabel;
		if (editingLabel) {
			labelDraft = nodeLabel;
		}
	}

	function saveLabel() {
		const trimmed = labelDraft.trim();
		if (trimmed) {
			dispatch('labelChange', trimmed);
		}
		editingLabel = false;
	}

	function scheduleSearch() {
		if (searchTimer) clearTimeout(searchTimer);
		searchTimer = setTimeout(() => {
			searchDebounced = searchQuery.trim();
			page = 1;
			void loadRecords();
		}, 300);
	}

	function goPrev() {
		if (!canPrev) return;
		page -= 1;
		void loadRecords();
	}

	function goNext() {
		if (!canNext) return;
		page += 1;
		void loadRecords();
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			dispatch('close');
		}
	}

	onMount(() => {
		void loadRecords();
	});

	onDestroy(() => {
		if (searchTimer) clearTimeout(searchTimer);
	});
</script>

<svelte:window on:keydown={handleKeydown} />

<!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
<div
	class="storage-records-panel"
	role="dialog"
	aria-label={$t('workflowBuilder.modalStorageRecords')}
	on:click|stopPropagation
	on:keydown|stopPropagation
>
	<header class="storage-records-header">
		<div class="storage-records-title-wrap">
			{#if editingLabel}
				<input
					type="text"
					class="storage-records-title-input"
					bind:value={labelDraft}
					on:blur={saveLabel}
					on:keydown={(e) => e.key === 'Enter' && saveLabel()}
					autofocus
				/>
			{:else}
				<h2 class="storage-records-title">{nodeLabel}</h2>
				<button
					type="button"
					class="storage-records-edit-label"
					title={$t('workflowBuilder.storageEditNodeLabelTitle')}
					aria-label={$t('workflowBuilder.storageEditNodeLabelAria')}
					on:click={toggleEditLabel}
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
							d="M16.862 3.487 20.51 7.136a1.5 1.5 0 0 1 0 2.121l-9.193 9.193-4.593.511a1 1 0 0 1-1.1-1.1l.511-4.593 9.193-9.193a1.5 1.5 0 0 1 2.121 0Z"
						/>
						<path d="M19 11.5 12.5 5" />
					</svg>
				</button>
			{/if}
		</div>
		<div class="storage-records-header-actions">
			<button
				type="button"
				class="btn-secondary storage-records-refresh"
				on:click={() => loadRecords()}
				aria-label={$t('workflowBuilder.storageRecordsRefresh')}
			>
				↻
			</button>
			<button
				type="button"
				class="storage-records-close"
				on:click={() => dispatch('close')}
				aria-label={$t('common.close')}
			>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					class="h-5 w-5"
				>
					<line x1="18" y1="6" x2="6" y2="18" />
					<line x1="6" y1="6" x2="18" y2="18" />
				</svg>
			</button>
		</div>
	</header>

	<div class="storage-records-body">
		<div class="storage-records-list-pane">
			<input
				type="search"
				class="storage-records-search"
				placeholder={$t('workflowBuilder.storageSearchPlaceholder')}
				bind:value={searchQuery}
				on:input={scheduleSearch}
			/>

			{#if loading}
				<p class="storage-records-status">{$t('common.loading')}</p>
			{:else if items.length === 0}
				<p class="storage-records-status">{$t('workflowBuilder.storageRecordsEmpty')}</p>
			{:else}
				<div class="storage-records-table-wrap">
					<table class="storage-records-table">
						<thead>
							<tr>
								<th>{$t('workflowBuilder.storageColDate')}</th>
								<th>{$t('workflowBuilder.storageColMode')}</th>
								<th>{$t('workflowBuilder.storageColSize')}</th>
								<th>{$t('workflowBuilder.storageColId')}</th>
								<th></th>
							</tr>
						</thead>
						<tbody>
							{#each items as record (record.id)}
								<tr
									class:selected={selectedRecordId === record.id}
									class="storage-records-row"
									on:click={() => selectRecord(record.id)}
								>
									<td class="storage-cell-date"
										>{new Date(record.createdAt).toLocaleString()}</td
									>
									<td>{record.mode}</td>
									<td>{formatStorageSize(record.size)}</td>
									<td class="storage-cell-id" title={record.id}>{record.id}</td>
									<td class="storage-cell-actions">
										<button
											type="button"
											class="btn-link"
											on:click={(e) => downloadRecord(record, e)}
										>
											{$t('workflowBuilder.storageDownload')}
										</button>
										<button
											type="button"
											class="btn-link text-red-600"
											on:click={(e) => removeRecord(record.id, e)}
										>
											{$t('workflowBuilder.storageDelete')}
										</button>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}

			{#if total > 0}
				<footer class="storage-records-pagination">
					<button type="button" class="btn-secondary" disabled={!canPrev} on:click={goPrev}>
						{$t('workflowBuilder.storagePaginationPrev')}
					</button>
					<span class="storage-records-range">
						{$t('workflowBuilder.storagePaginationRange', {
							from: rangeFrom,
							to: rangeTo,
							total,
						})}
					</span>
					<button type="button" class="btn-secondary" disabled={!canNext} on:click={goNext}>
						{$t('workflowBuilder.storagePaginationNext')}
					</button>
				</footer>
			{/if}
		</div>

		<div class="storage-records-preview-pane">
			{#if previewLoading}
				<p class="storage-records-status">{$t('common.loading')}</p>
			{:else if selectedRecordId && previewContent}
				<p class="storage-preview-id font-mono text-xs text-muted">{selectedRecordId}</p>
				<pre class="storage-record-preview">{previewContent}</pre>
			{:else}
				<p class="storage-records-status storage-records-preview-empty">
					{$t('workflowBuilder.storagePreviewPlaceholder')}
				</p>
			{/if}
		</div>
	</div>
</div>

<style>
	.storage-records-panel {
		position: absolute;
		inset: 0;
		z-index: 20;
		display: flex;
		flex-direction: column;
		min-height: 0;
		background: rgba(255, 255, 255, 0.97);
		backdrop-filter: blur(6px);
	}

	.storage-records-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		padding: 1rem 1.25rem;
		border-bottom: 1px solid rgba(226, 232, 240, 0.9);
		background: rgba(255, 255, 255, 0.9);
		flex-shrink: 0;
	}

	.storage-records-title-wrap {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		min-width: 0;
	}

	.storage-records-title {
		margin: 0;
		font-size: 1.125rem;
		font-weight: 600;
		color: #1e293b;
	}

	.storage-records-title-input {
		font-size: 1.125rem;
		font-weight: 600;
		border-radius: 0.5rem;
		border: 1px solid #e2e8f0;
		padding: 0.25rem 0.5rem;
		min-width: 12rem;
	}

	.storage-records-edit-label {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border: none;
		background: transparent;
		color: #64748b;
		cursor: pointer;
		padding: 0.25rem;
		border-radius: 0.375rem;
	}

	.storage-records-edit-label:hover {
		color: #1e293b;
		background: rgba(248, 250, 252, 0.9);
	}

	.storage-records-header-actions {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		flex-shrink: 0;
	}

	.storage-records-refresh {
		min-width: 2.5rem;
		padding: 0.375rem 0.75rem;
	}

	.storage-records-close {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border-radius: 999px;
		border: 1px solid rgba(148, 163, 184, 0.3);
		background: #fff;
		color: #64748b;
		cursor: pointer;
	}

	.storage-records-close:hover {
		color: #1e293b;
		border-color: rgba(148, 163, 184, 0.5);
	}

	.storage-records-body {
		display: grid;
		grid-template-columns: minmax(280px, 2fr) minmax(320px, 3fr);
		flex: 1;
		min-height: 0;
		overflow: hidden;
	}

	.storage-records-list-pane,
	.storage-records-preview-pane {
		display: flex;
		flex-direction: column;
		min-height: 0;
		overflow: hidden;
	}

	.storage-records-list-pane {
		border-right: 1px solid rgba(226, 232, 240, 0.9);
		padding: 1rem;
		gap: 0.75rem;
	}

	.storage-records-preview-pane {
		padding: 1rem;
		overflow: auto;
	}

	.storage-records-search {
		width: 100%;
		border-radius: 0.5rem;
		border: 1px solid #e2e8f0;
		padding: 0.5rem 0.75rem;
		font-size: 0.875rem;
		flex-shrink: 0;
	}

	.storage-records-status {
		font-size: 0.875rem;
		color: #64748b;
		margin: 0;
	}

	.storage-records-preview-empty {
		display: flex;
		align-items: center;
		justify-content: center;
		flex: 1;
		min-height: 8rem;
	}

	.storage-records-table-wrap {
		flex: 1;
		min-height: 0;
		overflow: auto;
	}

	.storage-records-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.8125rem;
	}

	.storage-records-table th,
	.storage-records-table td {
		padding: 0.5rem 0.75rem;
		border-bottom: 1px solid #e2e8f0;
		text-align: left;
		vertical-align: top;
	}

	.storage-records-row {
		cursor: pointer;
	}

	.storage-records-row:hover {
		background: rgba(248, 250, 252, 0.9);
	}

	.storage-records-row.selected {
		background: rgba(37, 99, 235, 0.08);
	}

	.storage-cell-id {
		font-family: ui-monospace, monospace;
		font-size: 0.75rem;
		max-width: 10rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.storage-cell-actions {
		white-space: nowrap;
	}

	.storage-cell-actions .btn-link {
		margin-right: 0.5rem;
		font-size: 0.75rem;
		color: #2563eb;
		background: none;
		border: none;
		cursor: pointer;
		padding: 0;
	}

	.storage-records-pagination {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem;
		flex-shrink: 0;
		padding-top: 0.5rem;
		border-top: 1px solid #e2e8f0;
	}

	.storage-records-range {
		font-size: 0.8125rem;
		color: #64748b;
	}

	.storage-preview-id {
		margin: 0 0 0.5rem;
		word-break: break-all;
	}

	.storage-record-preview {
		margin: 0;
		flex: 1;
		overflow: auto;
		font-size: 0.75rem;
		white-space: pre-wrap;
		word-break: break-word;
		background: #f8fafc;
		padding: 1rem;
		border-radius: 0.5rem;
		min-height: 12rem;
	}
</style>

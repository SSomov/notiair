<script lang="ts">
	import { t } from "$lib/i18n";
	import type { WorkflowDraft } from "@types/workflow";
	import type { Writable } from "svelte/store";

	export let workflows: Writable<WorkflowDraft[]>;

	$: workflowCount = $workflows.length;
</script>

<div class="glass-card h-[640px] w-full border-dashed border-border/60 bg-surface/60 p-0">
	<div class="flex h-full flex-col items-center justify-center gap-4 text-center text-muted">
		<span
			class="inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-surfaceMuted/50 text-3xl"
		>
			🧩
		</span>
		<div class="space-y-2">
			<p class="text-sm uppercase tracking-widest">{$t('workflowCanvas.wip')}</p>
			<p class="max-w-md text-sm text-muted">
				{$t('workflowCanvas.hint')}
			</p>
			<p class="text-xs text-muted/70">
				{$t('workflowCanvas.activeConfigs', { count: workflowCount })}
			</p>
		</div>
	</div>
</div>

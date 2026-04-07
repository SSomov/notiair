<script lang="ts">
import { t } from "$lib/i18n";
import { queueStore } from "@stores/queues";

const queue = queueStore;
</script>

<section class="space-y-6">
	<header class="space-y-2">
		<span class="pill">{$t('common.queues')}</span>
		<div>
			<h1 class="text-3xl font-semibold">{$t('queues.title')}</h1>
			<p class="text-sm text-muted">{$t('queues.description')}</p>
		</div>
	</header>

	<div class="glass-card overflow-hidden">
		<table class="min-w-full divide-y divide-border/60 text-sm">
			<thead class="bg-surfaceMuted/60 text-xs uppercase tracking-wide text-muted">
				<tr>
					<th class="px-4 py-3 text-left">{$t('queues.taskId')}</th>
					<th class="px-4 py-3 text-left">{$t('queues.workflow')}</th>
					<th class="px-4 py-3 text-left">{$t('queues.channel')}</th>
					<th class="px-4 py-3 text-left">{$t('queues.attempts')}</th>
					<th class="px-4 py-3 text-left">{$t('queues.status')}</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-border/40">
				{#each $queue as item}
					<tr class="hover:bg-surfaceMuted/40">
						<td class="px-4 py-3 text-text">{item.taskId}</td>
						<td class="px-4 py-3 text-muted">{item.workflowId}</td>
						<td class="px-4 py-3 text-muted">{item.channelId}</td>
						<td class="px-4 py-3 text-muted">{item.attempts}</td>
						<td class="px-4 py-3">
							<span class="pill capitalize">{item.status}</span>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</section>

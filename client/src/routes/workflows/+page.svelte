<script lang="ts">
	import { workflowStore } from '@stores/workflows';

	const draftBlueprints = [
		{
			id: 'draft-approval-flow',
			name: 'Approval flow v2',
			hint: 'Пауза + ручное подтверждение перед отправкой',
			updated: '2 часа назад',
		},
		{
			id: 'draft-personalized',
			name: 'Personalized onboarding',
			hint: 'Сегменты по гео и активности',
			updated: 'вчера',
		},
	];

	$: activeWorkflows = $workflowStore;
</script>

<section class="space-y-8 px-4 pb-12 pt-2 md:px-12 md:pt-4">
	<header class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
		<div class="space-y-2">
			<span class="pill">workflows</span>
			<p class="text-sm text-muted max-w-2xl">
				Собирайте автоматизации из триггеров, фильтров и действий. Управляйте версиями, создавайте
				черновики и переиспользуемые блоки для команд.
			</p>
		</div>
		<div class="flex flex-wrap gap-3">
			<a href="/templates" class="btn-primary">Перейти к шаблонам</a>
			<a
				href="/workflows/new"
				class="btn-primary bg-surfaceMuted text-text shadow-none hover:shadow-sm">Новый workflow</a
			>
		</div>
	</header>

	<div class="grid gap-8 lg:grid-cols-2">
		<div class="space-y-4">
			<div class="flex items-center justify-between">
				<div>
					<h2 class="text-lg font-semibold">Активные сценарии</h2>
					<p class="text-xs uppercase tracking-wide text-muted">
						Готовы к использованию • {activeWorkflows.length}
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
									aria-label="Скопировать workflow"
									title="Скопировать workflow"
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
									>active</span
								>
							</div>
						</div>
						<p class="mt-3 text-sm text-muted">
							Маршрутизация к каналам на основе фильтров. Можно обновить без остановки доставки.
						</p>
						<div class="mt-4 flex items-center justify-between">
							<button
								type="button"
								class="inline-flex text-sm font-semibold text-accent hover:underline"
								>Открыть конструктор</button
							>
						</div>
					</article>
				{/each}
			</div>
		</div>

		<div class="space-y-4">
			<div class="flex items-center justify-between">
				<div>
					<h2 class="text-lg font-semibold">Черновики</h2>
					<p class="text-xs uppercase tracking-wide text-muted">Для проверки и A/B</p>
				</div>
				<button type="button" class="text-sm font-semibold text-accent hover:underline"
					>Все черновики</button
				>
			</div>
			<div class="grid gap-4 sm:grid-cols-2">
				{#each draftBlueprints as draft}
					<article class="glass-card h-full">
						<div class="flex items-start justify-between gap-3">
							<div class="space-y-2">
								<h3 class="text-base font-semibold text-text">{draft.name}</h3>
								<p class="text-sm text-muted">{draft.hint}</p>
								<p class="text-xs text-muted">Обновлено {draft.updated}</p>
							</div>
							<button
								type="button"
								class="inline-flex h-8 w-8 items-center justify-center rounded-full border border-border text-muted transition hover:text-accent"
								aria-label="Скопировать черновик"
								title="Скопировать черновик"
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
							>Открыть черновик</button
						>
					</article>
				{/each}
				<article
					class="glass-card flex h-full items-center justify-center border-dashed border-border text-muted"
				>
					<div class="text-center">
						<p class="font-semibold">Создать черновик</p>
						<p class="text-sm text-muted">Скопируйте готовый сценарий или начните с пустого</p>
					</div>
				</article>
			</div>
		</div>
	</div>
</section>

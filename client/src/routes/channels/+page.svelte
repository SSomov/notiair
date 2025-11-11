<section class="space-y-8 px-4 pb-12 pt-2 md:px-12 md:pt-4">
	<header class="space-y-2">
		<span class="pill">channels</span>
		<p class="text-sm text-muted">
			Управляйте привязками каналов, назначайте workflow и контролируйте доступы.
		</p>
	</header>

	<div class="space-y-6">
		{#each [{ title: 'Telegram', description: 'Официальные боты и каналы', icon: '📣', channels: ['@product-updates', '@support'], color: 'text-accent' }, { title: 'SMTP', description: 'Почтовые рассылки и уведомления', icon: '✉️', channels: ['ops@notiair', 'alerts@notiair'], color: 'text-positive' }, { title: 'Slack', description: 'Интеграции со Slack workspace', icon: '💬', channels: ['#on-call', '#marketing'], color: 'text-warning' }] as group}
			<article class="glass-card">
				<div class="flex items-center justify-between">
					<div class="space-y-1">
						<p class="text-sm uppercase tracking-wide text-muted">{group.description}</p>
						<h2 class="flex items-center gap-2 text-2xl font-semibold">
							<span class={group.color}>{group.icon}</span>
							{group.title}
						</h2>
					</div>
					<button type="button" class="btn-primary hidden md:inline-flex">+</button>
				</div>

				<div class="mt-6 grid gap-4 md:grid-cols-3">
					{#each group.channels as channel}
						<div class="rounded-xl border border-border bg-surfaceMuted p-4">
							<p class="text-sm font-semibold text-text">{channel}</p>
							<p class="mt-2 text-xs text-muted">Привязан к 2 workflow</p>
							<button type="button" class="mt-3 text-sm font-semibold text-accent hover:underline">
								Настроить routing
							</button>
						</div>
					{/each}
					<button
						type="button"
						class="flex h-full items-center justify-center rounded-xl border border-dashed border-border text-muted transition hover:border-accent/60 hover:text-accent"
					>
						+ Добавить канал
					</button>
				</div>
			</article>
		{/each}
	</div>
</section>

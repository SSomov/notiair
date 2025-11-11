<section class="space-y-8 px-4 pb-12 pt-2 md:px-12 md:pt-4">
	<header class="space-y-2">
		<span class="pill">connectors</span>
		<p class="text-sm text-muted max-w-2xl">
			Управляйте интеграциями NotiAir: подключайте каналы, масштабируйте доставку и переиспользуйте
			общие настройки сервисов.
		</p>
	</header>

	<div class="grid gap-6 md:grid-cols-3">
		{#each [{ name: 'Telegram', description: 'Токены', badge: 'messanger', color: 'text-accent' }, { name: 'Slack', description: 'Вебхуки и workspace токены', badge: 'chat', color: 'text-warning' }, { name: 'SMTP', description: 'SMTP Credentials', badge: 'mail', color: 'text-positive' }] as connector}
			<article class="glass-card h-full space-y-4">
				<div class="flex items-center justify-between">
					<h2 class={`text-xl font-semibold ${connector.color}`}>{connector.name}</h2>
					<span class="pill capitalize">{connector.badge}</span>
				</div>
				<div class="space-y-2 text-sm text-muted">
					<p class="font-semibold text-text/80">{connector.subtitle}</p>
					<p>{connector.description}</p>
				</div>
				<div class="flex items-center gap-3">
					<button
						type="button"
						class="btn-primary bg-surfaceMuted text-text shadow-none hover:shadow-sm"
					>
						Добавить
					</button>
				</div>
			</article>
		{/each}
	</div>
</section>

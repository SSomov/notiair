<section class="space-y-8">
  <header class="space-y-3">
    <span class="pill">dashboard</span>
    <p class="text-muted max-w-2xl">
      Базовый обзор активности сервисов: обработанные события, задержки и распределение workflow по каналам.
    </p>
  </header>

  <div class="grid gap-6 lg:grid-cols-2">
    <article class="glass-card">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-lg font-semibold">События в очереди</h2>
          <p class="text-sm text-muted">Последние 24 часа</p>
        </div>
        <span class="pill">demo</span>
      </div>
      <div class="mt-6 h-52 rounded-xl bg-gradient-to-br from-accent/10 via-accent/5 to-surfaceMuted relative overflow-hidden">
        <div class="absolute inset-0 flex items-end gap-1 px-4 pb-4">
          {#each Array(12).fill(0).map((_, i) => 20 + (i % 4) * 12) as height}
            <span class="flex-1 rounded-t-md bg-accent/50" style={`height: ${height}%`}></span>
          {/each}
        </div>
      </div>
    </article>

    <article class="glass-card">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-lg font-semibold">Latency</h2>
          <p class="text-sm text-muted">Среднее время доставки по шагам</p>
        </div>
        <span class="pill">ms</span>
      </div>
      <div class="mt-6 h-52 rounded-xl bg-gradient-to-br from-amber-100 via-white to-surfaceMuted relative overflow-hidden">
        <div class="absolute inset-0 px-4 py-6">
          <svg viewBox="0 0 200 100" class="h-full w-full stroke-accent" fill="none">
            <path d="M0 80 C 40 60, 70 40, 100 60 S 160 30, 200 50" stroke-width="3" stroke-linejoin="round" stroke-linecap="round" />
          </svg>
        </div>
      </div>
    </article>
  </div>

  <article class="glass-card">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-lg font-semibold">Распределение по каналам</h2>
        <p class="text-sm text-muted">Статистика последних отправок</p>
      </div>
    </div>
    <div class="mt-6 grid gap-4 md:grid-cols-4">
      {#each ['@product-updates', '@support', '@ops', '@vip'] as channel, idx}
        <div class="rounded-xl border border-border bg-surfaceMuted p-4">
          <p class="text-sm font-semibold text-text">{channel}</p>
          <p class="mt-2 text-3xl font-bold text-accent">{(idx + 1) * 12}%</p>
          <p class="text-xs text-muted">доля от всех уведомлений</p>
        </div>
      {/each}
    </div>
  </article>
</section>


<script lang="ts">
  type ChannelLink = {
    name: string;
    description: string;
    muted: boolean;
  };

  type ConnectorEntry = {
    id: string;
    comment: string;
    channels: ChannelLink[];
  };

  type ChannelGroup = {
    slug: 'telegram' | 'slack' | 'smtp';
    name: string;
    icon: string;
    color: string;
    description: string;
    connectors: ConnectorEntry[];
  };


  let groups: ChannelGroup[] = [
    {
      slug: 'telegram',
      name: 'Telegram',
      icon: '✈️',
      color: 'text-accent',
      description: 'Боты и каналы Telegram.',
      connectors: [
        {
          id: '#TG-001',
          comment: 'Основной токен для product-уведомлений',
          channels: [
            { name: '@product-updates', description: 'Редакционные апдейты и релизы', muted: false },
            { name: '@support', description: 'Поддержка клиентов и эскалации', muted: false }
          ]
        }
      ]
    },
    {
      slug: 'slack',
      name: 'Slack',
      icon: '💬',
      color: 'text-warning',
      description: 'Каналы Slack.',
      connectors: [
        {
          id: '#SL-001',
          comment: 'Workspace alerts',
          channels: [
            { name: '#on-call', description: 'Дежурная смена и инциденты', muted: false },
            { name: '#marketing', description: 'Коммуникации маркетинга', muted: false }
          ]
        }
      ]
    },
    {
      slug: 'smtp',
      name: 'SMTP',
      icon: '📧',
      color: 'text-positive',
      description: 'Почтовые рассылки и транзакционные письма.',
      connectors: [
        {
          id: '#SM-001',
          comment: 'Primary transactional SMTP',
          channels: [
            { name: 'alerts@notiair', description: 'Оповещения инфраструктуры', muted: false },
            { name: 'billing@notiair', description: 'Счета и биллинговые уведомления', muted: false }
          ]
        }
      ]
    }
  ];

  let activeGroup: ChannelGroup | null = null;
  let activeEntry: ConnectorEntry | null = null;
  let channelInput = '';
  let channelDescription = '';
  let channelModalOpen = false;


  function openChannelModal(group: ChannelGroup, entry: ConnectorEntry) {
    activeGroup = group;
    activeEntry = entry;
    channelInput = '';
    channelDescription = '';
    channelModalOpen = true;
  }

  function closeModal() {
    channelModalOpen = false;
    activeGroup = null;
    activeEntry = null;
    channelInput = '';
    channelDescription = '';
  }


  function saveChannel() {
    if (!activeGroup || !activeEntry || !channelInput.trim()) return;
    const value = channelInput.trim();
    const description = channelDescription.trim();

    groups = groups.map((group) => {
      if (group.slug !== activeGroup?.slug) return group;
      return {
        ...group,
        connectors: group.connectors.map((entry) =>
          entry.id === activeEntry.id
            ? {
                ...entry,
                channels: [...entry.channels, { name: value, description, muted: false }]
              }
            : entry
        )
      };
    });
    closeModal();
  }

  function removeChannel(groupSlug: ChannelGroup['slug'], connectorId: string, channelName: string) {
    groups = groups.map((group) => {
      if (group.slug !== groupSlug) return group;
      return {
        ...group,
        connectors: group.connectors.map((entry) =>
          entry.id === connectorId
            ? {
                ...entry,
                channels: entry.channels.filter((channel) => channel.name !== channelName)
              }
            : entry
        )
      };
    });
  }

  function toggleMute(
    groupSlug: ChannelGroup['slug'],
    connectorId: string,
    channelName: string
  ) {
    groups = groups.map((group) => {
      if (group.slug !== groupSlug) return group;
      return {
        ...group,
        connectors: group.connectors.map((entry) =>
          entry.id === connectorId
            ? {
                ...entry,
                channels: entry.channels.map((channel) =>
                  channel.name === channelName ? { ...channel, muted: !channel.muted } : channel
                )
              }
            : entry
        )
      };
    });
  }
</script>

<section class="space-y-8 px-4 pb-12 pt-2 md:px-12 md:pt-4">
  <header class="space-y-2">
    <span class="pill">channels</span>
    <p class="text-sm text-muted">
      Управляйте привязками каналов, назначайте workflow и контролируйте доступы.
    </p>
  </header>

  <div class="space-y-6">
    {#each groups as group (group.slug)}
      <article class="glass-card space-y-6">
        <div class="flex items-center justify-between">
          <div class="space-y-1">
            <p class="text-sm uppercase tracking-wide text-muted">{group.description}</p>
            <h2 class="flex items-center gap-2 text-2xl font-semibold">
              <span class={group.color}>{group.icon}</span>
              {group.name}
            </h2>
          </div>
              </div>

        <div class="space-y-4">
          {#each group.connectors as entry (entry.id)}
            <div class="rounded-2xl border border-border bg-surfaceMuted/70 p-4 shadow-sm">
              <div class="flex items-start justify-between">
                <div>
                  <p class="text-sm font-semibold text-text">{entry.id}</p>
                  {#if entry.comment}
                    <p class="text-xs text-muted">{entry.comment}</p>
                  {/if}
                </div>
                <button
                  type="button"
                  class="btn-primary bg-surface text-text shadow-none hover:shadow-sm"
                  on:click={() => openChannelModal(group, entry)}
                >
                  Добавить канал
                </button>
              </div>

              {#if entry.channels.length > 0}
                <div class="mt-4 grid gap-3 text-sm text-muted sm:grid-cols-2 xl:grid-cols-3">
                  {#each entry.channels as channel (channel.name)}
                    <div class="flex flex-col justify-between rounded-2xl border border-border/60 bg-surface p-4 shadow-sm">
                      <div class="flex items-start justify-between gap-3">
                        <div>
                          <p class="text-sm font-semibold text-text">{channel.name}</p>
                          {#if channel.description}
                            <p class="mt-1 text-xs text-muted">{channel.description}</p>
                          {/if}
                        </div>
                        <button
                          type="button"
                          class="icon-btn"
                          aria-label="Удалить канал"
                          on:click={() => removeChannel(group.slug, entry.id, channel.name)}
                        >
                          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="h-4 w-4">
                            <path d="M6 7h12" />
                            <path d="M10 11v6" />
                            <path d="M14 11v6" />
                            <path d="M5 7l1 12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2l1-12" />
                            <path d="M9 7V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v3" />
                          </svg>
                        </button>
                      </div>
                      <div class="mt-3 flex items-center justify-between text-xs">
                        {#if channel.muted}
                          <span class="inline-flex items-center rounded-full bg-border/70 px-2.5 py-1 font-semibold uppercase tracking-wide text-text">
                            Muted
                          </span>
                        {:else}
                          <span class="text-muted">Активен</span>
                        {/if}
                        <button
                          type="button"
                          class="inline-flex items-center justify-center gap-1 rounded-lg border border-border px-3 py-1 font-semibold text-muted transition hover:border-primary hover:text-primary"
                          on:click={() => toggleMute(group.slug, entry.id, channel.name)}
                        >
                          {channel.muted ? 'Unmute' : 'Mute'}
                        </button>
                      </div>
                    </div>
                  {/each}
                </div>
              {/if}
            </div>
          {/each}
        </div>
      </article>
    {/each}
  </div>

  {#if channelModalOpen && activeGroup && activeEntry}
    <div class="modal-backdrop" role="presentation" on:click={closeModal}></div>
    <div class="modal-wrap" role="dialog" aria-modal="true">
      <div class="modal">
        <div class="modal-header">
          <h3 class="text-lg font-semibold text-text">
            Добавить канал для {activeEntry.id}
          </h3>
          <button type="button" class="modal-close" on:click={closeModal} aria-label="Закрыть">
            ✕
          </button>
        </div>

        <div class="modal-body space-y-4">
          <div class="space-y-1">
            <label class="text-sm font-medium text-text" for="channel-input">Канал</label>
            <input
              id="channel-input"
              class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:border-accent focus:outline-none"
              bind:value={channelInput}
              placeholder="Например: @status-updates или #marketing"
              autocomplete="off"
            />
          </div>
          <div class="space-y-1">
            <label class="text-sm font-medium text-text" for="channel-description">Описание</label>
            <textarea
              id="channel-description"
              class="h-20 w-full resize-none rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:border-accent focus:outline-none"
              bind:value={channelDescription}
              placeholder="Кратко объясните назначение канала"
            ></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn-secondary" on:click={closeModal}>Отменить</button>
          <button
            type="button"
            class="btn-primary"
            on:click={saveChannel}
            disabled={!channelInput.trim()}
          >
            Сохранить
          </button>
        </div>
      </div>
    </div>
  {/if}
</section>

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(15, 23, 42, 0.26);
    backdrop-filter: blur(6px);
    z-index: 40;
  }

  .modal-wrap {
    position: fixed;
    inset: 0;
    z-index: 50;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 1.5rem;
  }

  .modal {
    width: min(420px, 100%);
    border-radius: 1.25rem;
    background: var(--surface, #ffffff);
    border: 1px solid var(--border, #e2e8f0);
    box-shadow: 0 30px 60px -35px rgba(15, 23, 42, 0.45);
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    padding: 1.75rem;
    max-height: calc(100vh - 4rem);
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .modal-body {
    overflow-y: auto;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
  }

  .modal-close,
  .icon-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 30px;
    height: 30px;
    border-radius: 999px;
    border: 1px solid rgba(148, 163, 184, 0.3);
    background: #fff;
    color: #64748b;
    transition: 120ms ease;
  }

  .modal-close:hover,
  .icon-btn:hover {
    color: #2563eb;
    border-color: rgba(37, 99, 235, 0.5);
    transform: translateY(-1px);
  }

  .btn-secondary {
    @apply inline-flex items-center justify-center gap-2 rounded-xl border border-border bg-surface px-5 py-2 text-sm font-semibold text-muted transition hover:text-text;
  }
</style>


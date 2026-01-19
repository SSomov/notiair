<script lang="ts">
	import { onMount } from "svelte";
	import {
		listTelegramTokens,
		listChannels,
		createChannel,
		updateChannel,
		deleteChannel,
		type TelegramToken,
		type Channel,
	} from "$lib/api";

	type ChannelLink = Channel & {
		// Alias для совместимости
	};

	type ConnectorEntry = {
		id: string;
		name: string;
		comment: string;
		channels: ChannelLink[];
	};

	type ChannelGroup = {
		slug: "telegram" | "slack" | "smtp";
		name: string;
		icon: string;
		color: string;
		description: string;
		connectors: ConnectorEntry[];
	};

	let telegramTokens: TelegramToken[] = [];
	let loading = true;
	let error: string | null = null;
	let saving = false;

	let groups: ChannelGroup[] = [
		{
			slug: "telegram",
			name: "Telegram",
			icon: "✈️",
			color: "text-accent",
			description: "Боты и каналы Telegram.",
			connectors: [],
		},
	];

	let activeGroup: ChannelGroup | null = null;
	let activeEntry: ConnectorEntry | null = null;
	let editingChannel: ChannelLink | null = null;
	let channelInput = "";
	let channelNameInput = "";
	let channelDescription = "";
	let channelModalOpen = false;

	onMount(async () => {
		await loadData();
	});

	async function loadData() {
		try {
			loading = true;
			error = null;
			telegramTokens = await listTelegramTokens();
			await updateGroups();
		} catch (e) {
			console.error("Failed to load data:", e);
			error = e instanceof Error ? e.message : "Не удалось загрузить данные";
			telegramTokens = [];
			await updateGroups();
		} finally {
			loading = false;
		}
	}

	async function updateGroups() {
		groups = await Promise.all(
			groups.map(async (group) => {
				if (group.slug === "telegram") {
					const connectors = await Promise.all(
						telegramTokens
							.filter((token) => token.isActive)
							.map(async (token) => {
								try {
									const channels = await listChannels(token.id);
									return {
										id: token.id,
										name: token.name || "Без названия",
										comment: token.comment || "",
										channels: channels.map((ch) => ({
											id: ch.id,
											name: ch.name,
											displayName: ch.displayName,
											description: ch.description,
											muted: ch.muted,
										})),
									};
								} catch (e) {
									console.error(`Failed to load channels for ${token.id}:`, e);
									return {
										id: token.id,
										name: token.name || "Без названия",
										comment: token.comment || "",
										channels: [],
									};
								}
							}),
					);
					return {
						...group,
						connectors,
					};
				}
				return group;
			}),
		);
	}

	$: filteredGroups = groups.filter((group) => group.connectors.length > 0);

	function openChannelModal(group: ChannelGroup, entry: ConnectorEntry, channel?: ChannelLink) {
		activeGroup = group;
		activeEntry = entry;
		editingChannel = channel || null;
		channelInput = channel?.name || "";
		channelNameInput = channel?.displayName || "";
		channelDescription = channel?.description || "";
		channelModalOpen = true;
	}

	function closeModal() {
		channelModalOpen = false;
		activeGroup = null;
		activeEntry = null;
		editingChannel = null;
		channelInput = "";
		channelNameInput = "";
		channelDescription = "";
	}

	async function saveChannel() {
		if (!activeGroup || !activeEntry || !channelInput.trim() || saving) return;

		try {
			saving = true;
			error = null;

			const value = channelInput.trim();
			const displayName = channelNameInput.trim() || undefined;
			const description = channelDescription.trim();

			if (editingChannel) {
				const updated = await updateChannel(editingChannel.id, {
					name: value,
					displayName,
					description,
					muted: editingChannel.muted,
				});
				groups = groups.map((group) => {
					if (group.slug !== activeGroup?.slug) return group;
					return {
						...group,
						connectors: group.connectors.map((entry) =>
							entry.id === activeEntry.id
								? {
										...entry,
										channels: entry.channels.map((ch) =>
											ch.id === editingChannel.id
												? {
														id: updated.id,
														name: updated.name,
														displayName: updated.displayName,
														description: updated.description,
														muted: updated.muted,
													}
												: ch,
										),
									}
								: entry,
						),
					};
				});
			} else {
				const created = await createChannel(activeEntry.id, {
					name: value,
					displayName,
					description,
					muted: false,
				});
				groups = groups.map((group) => {
					if (group.slug !== activeGroup?.slug) return group;
					return {
						...group,
						connectors: group.connectors.map((entry) =>
							entry.id === activeEntry.id
								? {
										...entry,
										channels: [
											...entry.channels,
											{
												id: created.id,
												name: created.name,
												displayName: created.displayName,
												description: created.description,
												muted: created.muted,
											},
										],
									}
								: entry,
						),
					};
				});
			}

			closeModal();
		} catch (e) {
			error = e instanceof Error ? e.message : "Не удалось сохранить канал";
		} finally {
			saving = false;
		}
	}

	async function removeChannel(
		groupSlug: ChannelGroup["slug"],
		connectorId: string,
		channelId: string,
	) {
		if (!confirm("Вы уверены, что хотите удалить этот канал?")) return;

		try {
			error = null;
			await deleteChannel(channelId);
			groups = groups.map((group) => {
				if (group.slug !== groupSlug) return group;
				return {
					...group,
					connectors: group.connectors.map((entry) =>
						entry.id === connectorId
							? {
									...entry,
									channels: entry.channels.filter((channel) => channel.id !== channelId),
								}
							: entry,
					),
				};
			});
		} catch (e) {
			error = e instanceof Error ? e.message : "Не удалось удалить канал";
		}
	}

	async function toggleMute(
		groupSlug: ChannelGroup["slug"],
		connectorId: string,
		channel: ChannelLink,
	) {
		try {
			error = null;
			const updated = await updateChannel(channel.id, {
				name: channel.name,
				displayName: channel.displayName,
				description: channel.description,
				muted: !channel.muted,
			});
			groups = groups.map((group) => {
				if (group.slug !== groupSlug) return group;
				return {
					...group,
					connectors: group.connectors.map((entry) =>
						entry.id === connectorId
							? {
									...entry,
									channels: entry.channels.map((ch) =>
										ch.id === channel.id
											? {
													id: updated.id,
													name: updated.name,
													displayName: updated.displayName,
													description: updated.description,
													muted: updated.muted,
												}
											: ch,
									),
								}
							: entry,
					),
				};
			});
		} catch (e) {
			error = e instanceof Error ? e.message : "Не удалось изменить статус канала";
		}
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
    {#if error}
      <div class="glass-card p-4 bg-red-50 border-red-200">
        <p class="text-sm text-red-600">{error}</p>
      </div>
    {/if}

    {#if loading}
      <div class="glass-card p-8 text-center">
        <p class="text-sm text-muted">Загрузка коннекторов...</p>
      </div>
    {:else if filteredGroups.length === 0}
      <div class="glass-card p-8 text-center">
        <p class="text-sm text-muted">Нет доступных коннекторов</p>
      </div>
    {:else}
      {#each filteredGroups as group (group.slug)}
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
                  <p class="text-sm font-semibold text-text">{entry.name}</p>
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
                          <p class="text-sm font-semibold text-text">
                            {channel.displayName || channel.name}
                          </p>
                          {#if channel.displayName}
                            <p class="mt-0.5 text-xs text-muted font-mono">{channel.name}</p>
                          {/if}
                          {#if channel.description}
                            <p class="mt-1 text-xs text-muted">{channel.description}</p>
                          {/if}
                        </div>
                        <div class="flex items-center gap-2">
                          <button
                            type="button"
                            class="icon-btn"
                            aria-label="Редактировать канал"
                            on:click={() => openChannelModal(group, entry, channel)}
                          >
                            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="h-4 w-4">
                              <path d="M16.862 3.487 20.51 7.136a1.5 1.5 0 0 1 0 2.121l-9.193 9.193-4.593.511a1 1 0 0 1-1.1-1.1l.511-4.593 9.193-9.193a1.5 1.5 0 0 1 2.121 0Z" />
                              <path d="M19 11.5 12.5 5" />
                            </svg>
                          </button>
                          <button
                            type="button"
                            class="icon-btn"
                            aria-label="Удалить канал"
                            on:click={() => removeChannel(group.slug, entry.id, channel.id)}
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
                          on:click={() => toggleMute(group.slug, entry.id, channel)}
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
    {/if}
  </div>

  {#if channelModalOpen && activeGroup && activeEntry}
    <div class="modal-backdrop" role="presentation" on:click={closeModal}></div>
    <div class="modal-wrap" role="dialog" aria-modal="true">
      <div class="modal">
        <div class="modal-header">
          <h3 class="text-lg font-semibold text-text">
            {editingChannel ? 'Редактировать канал' : 'Добавить канал'} для {activeEntry.name}
          </h3>
          <button type="button" class="modal-close" on:click={closeModal} aria-label="Закрыть">
            ✕
          </button>
        </div>

        <div class="modal-body space-y-4">
          <div class="space-y-1">
            <label class="text-sm font-medium text-text" for="channel-name-input">
              Название <span class="text-xs text-muted">(необязательно)</span>
            </label>
            <input
              id="channel-name-input"
              class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:border-accent focus:outline-none"
              bind:value={channelNameInput}
              placeholder="Например: Статус обновлений"
              autocomplete="off"
            />
          </div>
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
            disabled={!channelInput.trim() || saving}
          >
            {saving ? 'Сохранение...' : editingChannel ? 'Сохранить' : 'Добавить'}
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


<script lang="ts">
	import { onMount } from "svelte";
	import { page } from "$app/stores";
	import { resolve } from "$app/paths";
	import { getLocaleFromPath, addLocaleToPath } from "$lib/i18n/utils";
	import { get } from "svelte/store";
	import { locale, t } from "$lib/i18n";
	import { resolveI18nError } from "$lib/i18n/resolveError";
	import {
		listTelegramTokens,
		createTelegramToken,
		updateTelegramToken,
		deleteTelegramToken,
		toggleTelegramTokenActive,
		listChannels,
		createChannel,
		updateChannel,
		deleteChannel,
		type TelegramToken,
		type Channel,
	} from "$lib/api";

	const THREAD_SLOT = 96;
	const THREAD_W = 64;

	let tokens: TelegramToken[] = [];
	let channelsByTokenId: Record<string, Channel[]> = {};
	let loading = true;
	let error: string | null = null;
	let errorDisplay: string | null = null;

	let modalOpen = false;
	let editing: TelegramToken | null = null;
	let nameInput = "";
	let secretInput = "";
	let commentInput = "";

	let channelModalOpen = false;
	let channelModalToken: TelegramToken | null = null;
	let editingChannel: Channel | null = null;
	let channelInput = "";
	let channelNameInput = "";
	let channelDescription = "";
	let savingChannel = false;

	$: loc = getLocaleFromPath($page.url.pathname);
	$: hrefConnectors = resolve(addLocaleToPath("/connectors", loc));
	let saving = false;

	function maskSecret(secret: string): string {
		if (!secret || secret.length <= 8) return "••••••••";
		const start = secret.substring(0, 2);
		const end = secret.substring(secret.length - 2);
		const masked = "•".repeat(Math.min(12, secret.length - 4));
		return `${start}${masked}${end}`;
	}

	function threadGraphic(count: number) {
		const H = count * THREAD_SLOT;
		const yStart = H / 2;
		const items: { d: string; yEnd: number }[] = [];
		for (let i = 0; i < count; i++) {
			const yEnd = i * THREAD_SLOT + THREAD_SLOT / 2;
			items.push({
				d: `M 0 ${yStart} C 22 ${yStart}, 44 ${yEnd}, ${THREAD_W} ${yEnd}`,
				yEnd,
			});
		}
		return { H, yStart, items };
	}

	onMount(async () => {
		await loadTokens();
	});

	async function refreshChannelMap() {
		if (tokens.length === 0) {
			channelsByTokenId = {};
			return;
		}
		const entries = await Promise.all(
			tokens.map(async (tok) => {
				try {
					const ch = await listChannels(tok.id);
					return [tok.id, ch] as const;
				} catch {
					return [tok.id, []] as const;
				}
			}),
		);
		channelsByTokenId = Object.fromEntries(entries);
	}

	async function loadTokens() {
		try {
			loading = true;
			error = null;
			const data = await listTelegramTokens();
			tokens = Array.isArray(data) ? data : [];
			await refreshChannelMap();
		} catch (e) {
			error = e instanceof Error ? e.message : "errors.loadTokens";
			tokens = [];
			channelsByTokenId = {};
		} finally {
			loading = false;
		}
	}

	function openModal(tok?: TelegramToken) {
		editing = tok || null;
		nameInput = tok?.name || "";
		secretInput = tok?.secret || "";
		commentInput = tok?.comment || "";
		modalOpen = true;
		error = null;
	}

	function closeModal() {
		modalOpen = false;
		editing = null;
		nameInput = "";
		secretInput = "";
		commentInput = "";
		error = null;
	}

	async function handleSave() {
		if (!secretInput.trim() || !nameInput.trim() || saving) return;

		try {
			saving = true;
			error = null;

			if (editing) {
				const id = editing.id;
				const updated = await updateTelegramToken(id, {
					name: nameInput.trim(),
					secret: secretInput.trim(),
					comment: commentInput.trim(),
				});
				tokens = tokens.map((row) => (row.id === id ? updated : row));
			} else {
				const created = await createTelegramToken({
					name: nameInput.trim(),
					secret: secretInput.trim(),
					comment: commentInput.trim(),
				});
				tokens = [...tokens, created];
				channelsByTokenId = { ...channelsByTokenId, [created.id]: [] };
			}

			closeModal();
		} catch (e) {
			error = e instanceof Error
				? e.message
				: editing
					? "errors.updateToken"
					: "errors.createToken";
		} finally {
			saving = false;
		}
	}

	async function handleDelete(id: string) {
		if (!confirm(get(t)("telegramConnectorPage.confirmDelete"))) return;

		try {
			error = null;
			await deleteTelegramToken(id);
			tokens = tokens.filter((row) => row.id !== id);
			const next = { ...channelsByTokenId };
			delete next[id];
			channelsByTokenId = next;
		} catch (e) {
			error = e instanceof Error ? e.message : "errors.deleteToken";
		}
	}

	async function handleToggleActive(tok: TelegramToken) {
		try {
			error = null;
			const updated = await toggleTelegramTokenActive(tok.id, !tok.isActive);
			tokens = tokens.map((row) => (row.id === tok.id ? updated : row));
		} catch (e) {
			error = e instanceof Error ? e.message : "errors.toggleTokenStatus";
		}
	}

	function openChannelModal(tok: TelegramToken, ch?: Channel) {
		channelModalToken = tok;
		editingChannel = ch || null;
		channelInput = ch?.name || "";
		channelNameInput = ch?.displayName || "";
		channelDescription = ch?.description || "";
		channelModalOpen = true;
		error = null;
	}

	function closeChannelModal() {
		channelModalOpen = false;
		channelModalToken = null;
		editingChannel = null;
		channelInput = "";
		channelNameInput = "";
		channelDescription = "";
	}

	async function handleSaveChannel() {
		if (!channelModalToken || !channelInput.trim() || savingChannel) return;

		try {
			savingChannel = true;
			error = null;
			const tid = channelModalToken.id;
			const value = channelInput.trim();
			const displayName = channelNameInput.trim() || undefined;
			const description = channelDescription.trim();

			if (editingChannel) {
				const ec = editingChannel;
				const updated = await updateChannel(ec.id, {
					name: value,
					displayName,
					description,
					muted: ec.muted,
				});
				channelsByTokenId = {
					...channelsByTokenId,
					[tid]: (channelsByTokenId[tid] || []).map((c) =>
						c.id === ec.id ? updated : c,
					),
				};
			} else {
				const created = await createChannel(tid, {
					name: value,
					displayName,
					description,
					muted: false,
				});
				const list = channelsByTokenId[tid] || [];
				channelsByTokenId = { ...channelsByTokenId, [tid]: [...list, created] };
			}

			closeChannelModal();
		} catch (e) {
			error = e instanceof Error ? e.message : "errors.saveChannel";
		} finally {
			savingChannel = false;
		}
	}

	async function handleDeleteChannel(tokenId: string, channelId: string) {
		if (!confirm(get(t)("channelsPage.confirmDeleteChannel"))) return;

		try {
			error = null;
			await deleteChannel(channelId);
			channelsByTokenId = {
				...channelsByTokenId,
				[tokenId]: (channelsByTokenId[tokenId] || []).filter((c) => c.id !== channelId),
			};
		} catch (e) {
			error = e instanceof Error ? e.message : "errors.deleteChannel";
		}
	}

	$: {
		$locale;
		errorDisplay = error ? resolveI18nError(error) : null;
	}
</script>

<section class="space-y-8 px-4 pb-12 pt-2 md:px-12 md:pt-4">
  <header class="space-y-2">
    <div class="flex items-center gap-3">
      <a href={hrefConnectors} class="text-muted hover:text-text transition">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="h-5 w-5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 19.5 8.25 12l7.5-7.5" />
        </svg>
      </a>
      <span class="pill">{$t('telegramConnectorPage.pill')}</span>
    </div>
    <p class="text-sm text-muted max-w-2xl">
      {$t('telegramConnectorPage.intro')}
    </p>
  </header>

  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-lg font-semibold text-text">{$t('telegramConnectorPage.title')}</h2>
      <button
        type="button"
        class="btn-primary bg-surfaceMuted text-text shadow-none hover:shadow-sm"
        on:click={() => openModal()}
      >
        {$t('telegramConnectorPage.addToken')}
      </button>
    </div>

    {#if error}
      <div class="glass-card p-4 bg-red-50 border-red-200">
        <p class="text-sm text-red-600">{errorDisplay}</p>
      </div>
    {/if}

    {#if loading}
      <div class="glass-card p-8 text-center">
        <p class="text-sm text-muted">{$t('telegramConnectorPage.loading')}</p>
      </div>
    {:else if tokens.length === 0}
      <div class="glass-card p-8 text-center">
        <p class="text-sm text-muted">{$t('telegramConnectorPage.empty')}</p>
      </div>
    {:else}
      <div class="space-y-6">
        {#each tokens as token}
          {@const chans = channelsByTokenId[token.id] || []}
          {@const tg = threadGraphic(chans.length)}
          <div
            class="flex flex-col gap-4 rounded-2xl border border-border/40 bg-surfaceMuted/30 p-3 sm:p-4 lg:flex-row lg:items-stretch"
          >
            <div class="glass-card min-w-0 flex-1 space-y-3 p-4">
              <div class="flex items-start justify-between gap-3">
                <div class="flex-1 space-y-2 min-w-0">
                  <div class="space-y-1">
                    <div class="flex items-center justify-between gap-2">
                      <p class="font-semibold text-text truncate">{token.name || $t('common.noName')}</p>
                      <label class="relative inline-flex items-center cursor-pointer shrink-0">
                        <input
                          type="checkbox"
                          checked={token.isActive}
                          on:change={() => handleToggleActive(token)}
                          class="sr-only peer"
                        />
                        <div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-accent rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-accent"></div>
                        <span class="ml-2 text-xs text-muted">{token.isActive ? $t('common.activeFemale') : $t('common.inactive')}</span>
                      </label>
                    </div>
                    <p class="text-xs font-mono text-muted break-all">
                      {$t('telegramConnectorPage.secretPrefix')} {maskSecret(token.secret)}
                    </p>
                  </div>
                  {#if token.comment}
                    <p class="text-sm text-muted">{token.comment}</p>
                  {/if}
                </div>
                <div class="flex items-center gap-2 shrink-0">
                  <button
                    type="button"
                    class="icon-btn text-accent border-accent/40 hover:border-accent"
                    title={$t('telegramConnectorPage.addChannelTitle')}
                    aria-label={$t('telegramConnectorPage.addChannelAria')}
                    on:click={() => openChannelModal(token)}
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="h-4 w-4">
                      <path stroke-linecap="round" d="M12 5v14M5 12h14" />
                    </svg>
                  </button>
                  <button
                    type="button"
                    class="icon-btn"
                    title={$t('common.edit')}
                    aria-label={$t('common.edit')}
                    on:click={() => openModal(token)}
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="h-4 w-4">
                      <path d="M16.862 3.487 20.51 7.136a1.5 1.5 0 0 1 0 2.121l-9.193 9.193-4.593.511a1 1 0 0 1-1.1-1.1l.511-4.593 9.193-9.193a1.5 1.5 0 0 1 2.121 0Z" />
                      <path d="M19 11.5 12.5 5" />
                    </svg>
                  </button>
                  <button
                    type="button"
                    class="icon-btn"
                    title={$t('common.delete')}
                    aria-label={$t('common.delete')}
                    on:click={() => handleDelete(token.id)}
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
            </div>

            {#if chans.length > 0}
              <div class="flex min-h-0 min-w-0 flex-1 flex-row items-stretch gap-0 lg:max-w-xl">
                <div
                  class="relative min-h-[4.5rem] w-10 shrink-0 sm:w-12 md:w-16"
                  aria-hidden="true"
                >
                  <svg
                    class="absolute inset-0 h-full w-full overflow-visible"
                    viewBox="0 0 {THREAD_W} {tg.H}"
                    preserveAspectRatio="none"
                  >
                    <circle
                      cx="0"
                      cy={tg.yStart}
                      r="5"
                      fill="var(--surface, #ffffff)"
                      stroke="#2563eb"
                      stroke-width="2"
                    />
                    {#each tg.items as it}
                      <path
                        d={it.d}
                        stroke="#2563eb"
                        stroke-width="2"
                        fill="none"
                      />
                      <circle
                        cx={THREAD_W}
                        cy={it.yEnd}
                        r="5"
                        fill="var(--surface, #ffffff)"
                        stroke="#2563eb"
                        stroke-width="2"
                      />
                    {/each}
                  </svg>
                </div>
                <div class="flex min-w-0 flex-1 flex-col justify-center gap-2">
                  {#each chans as ch (ch.id)}
                    <div
                      class="flex min-h-24 min-w-0 flex-col justify-center rounded-2xl border border-border/70 bg-surface px-3 py-2.5 shadow-sm"
                    >
                      <div class="flex items-start justify-between gap-2">
                        <div class="min-w-0 flex-1">
                          <p class="text-sm font-semibold text-text truncate">
                            {ch.displayName || ch.name}
                          </p>
                          {#if ch.displayName}
                            <p class="mt-0.5 truncate font-mono text-xs text-muted">{ch.name}</p>
                          {/if}
                          {#if ch.description}
                            <p class="mt-1 line-clamp-2 text-xs text-muted">{ch.description}</p>
                          {/if}
                        </div>
                        <div class="flex shrink-0 items-center gap-1">
                          <button
                            type="button"
                            class="icon-btn h-8 w-8"
                            aria-label={$t('channelsPage.editChannelAria')}
                            on:click={() => openChannelModal(token, ch)}
                          >
                            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="h-3.5 w-3.5">
                              <path d="M16.862 3.487 20.51 7.136a1.5 1.5 0 0 1 0 2.121l-9.193 9.193-4.593.511a1 1 0 0 1-1.1-1.1l.511-4.593 9.193-9.193a1.5 1.5 0 0 1 2.121 0Z" />
                              <path d="M19 11.5 12.5 5" />
                            </svg>
                          </button>
                          <button
                            type="button"
                            class="icon-btn h-8 w-8"
                            aria-label={$t('channelsPage.deleteChannelAria')}
                            on:click={() => handleDeleteChannel(token.id, ch.id)}
                          >
                            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="h-3.5 w-3.5">
                              <path d="M6 7h12" />
                              <path d="M10 11v6" />
                              <path d="M14 11v6" />
                              <path d="M5 7l1 12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2l1-12" />
                              <path d="M9 7V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v3" />
                            </svg>
                          </button>
                        </div>
                      </div>
                    </div>
                  {/each}
                </div>
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>

  {#if modalOpen}
    <div class="modal-backdrop" role="presentation" on:click={closeModal}></div>
    <div class="modal-wrap" role="dialog" aria-modal="true">
      <div class="modal modal-wide">
        <div class="modal-header">
          <h3 class="text-lg font-semibold text-text">
            {editing ? $t('telegramConnectorPage.editTitle') : $t('telegramConnectorPage.newTitle')}
          </h3>
          <button type="button" class="modal-close" on:click={closeModal} aria-label={$t('common.close')}>
            ✕
          </button>
        </div>
        <div class="modal-body space-y-4">
          <div class="space-y-1">
            <label class="text-sm font-medium text-text" for="tg-name">{$t('common.name')}</label>
            <input
              id="tg-name"
              class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:border-accent focus:outline-none"
              bind:value={nameInput}
              placeholder={$t('telegramConnectorPage.namePlaceholder')}
              autocomplete="off"
            />
          </div>
          <div class="space-y-1">
            <label class="text-sm font-medium text-text" for="tg-secret">{$t('telegramConnectorPage.secretLabel')}</label>
            <input
              id="tg-secret"
              type="password"
              class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:border-accent focus:outline-none"
              bind:value={secretInput}
              placeholder={$t('telegramConnectorPage.secretPlaceholder')}
              autocomplete="new-password"
            />
          </div>
          <div class="space-y-1">
            <label class="text-sm font-medium text-text" for="tg-comment">{$t('common.comment')}</label>
            <textarea
              id="tg-comment"
              class="h-20 w-full resize-none rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:border-accent focus:outline-none"
              bind:value={commentInput}
              placeholder={$t('common.internalNotes')}
            ></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn-secondary" on:click={closeModal}>{$t('common.cancel')}</button>
          <button
            type="button"
            class="btn-primary"
            on:click={handleSave}
            disabled={!secretInput.trim() || !nameInput.trim() || saving}
          >
            {saving ? $t('common.saving') : editing ? $t('common.save') : $t('common.add')}
          </button>
        </div>
      </div>
    </div>
  {/if}

  {#if channelModalOpen && channelModalToken}
    <div class="modal-backdrop" style="z-index: 45" role="presentation" on:click={closeChannelModal}></div>
    <div class="modal-wrap" style="z-index: 55" role="dialog" aria-modal="true">
      <div class="modal">
        <div class="modal-header">
          <h3 class="text-lg font-semibold text-text">
            {editingChannel ? $t('channelsPage.modalTitleEdit') : $t('channelsPage.modalTitleAdd')}
            {' '}
            {$t('channelsPage.modalFor')} {channelModalToken.name || $t('common.noName')}
          </h3>
          <button type="button" class="modal-close" on:click={closeChannelModal} aria-label={$t('common.close')}>
            ✕
          </button>
        </div>

        <div class="modal-body space-y-4">
          <div class="space-y-1">
            <label class="text-sm font-medium text-text" for="tg-ch-display">
              {$t('common.nameOptional')}
            </label>
            <input
              id="tg-ch-display"
              class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:border-accent focus:outline-none"
              bind:value={channelNameInput}
              placeholder={$t('channelsPage.placeholderDisplayName')}
              autocomplete="off"
            />
          </div>
          <div class="space-y-1">
            <label class="text-sm font-medium text-text" for="tg-ch-name">{$t('common.channel')}</label>
            <input
              id="tg-ch-name"
              class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:border-accent focus:outline-none"
              bind:value={channelInput}
              placeholder={$t('channelsPage.placeholderChannel')}
              autocomplete="off"
            />
          </div>
          <div class="space-y-1">
            <label class="text-sm font-medium text-text" for="tg-ch-desc">{$t('common.description')}</label>
            <textarea
              id="tg-ch-desc"
              class="h-20 w-full resize-none rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:border-accent focus:outline-none"
              bind:value={channelDescription}
              placeholder={$t('channelsPage.placeholderDescription')}
            ></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn-secondary" on:click={closeChannelModal}>{$t('common.cancel')}</button>
          <button
            type="button"
            class="btn-primary"
            on:click={handleSaveChannel}
            disabled={!channelInput.trim() || savingChannel}
          >
            {savingChannel ? $t('common.saving') : editingChannel ? $t('common.save') : $t('common.add')}
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

  .modal-wide {
    width: min(520px, 100%);
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

  .modal-close {
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

  .modal-close:hover {
    color: #2563eb;
    border-color: rgba(37, 99, 235, 0.5);
    transform: translateY(-1px);
  }

  .btn-secondary {
    @apply inline-flex items-center justify-center gap-2 rounded-xl border border-border bg-surface px-5 py-2 text-sm font-semibold text-muted transition hover:text-text;
  }

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

  .icon-btn:hover {
    color: #2563eb;
    border-color: rgba(37, 99, 235, 0.5);
    transform: translateY(-1px);
  }
</style>

<script lang="ts">
	import { onMount } from "svelte";
	import { page } from "$app/stores";
	import { resolve } from "$app/paths";
	import { getLocaleFromPath, addLocaleToPath } from "$lib/i18n/utils";
	import {
		listSmtpAccounts,
		createSmtpAccount,
		updateSmtpAccount,
		deleteSmtpAccount,
		toggleSmtpAccountActive,
		type SmtpAccount,
	} from "$lib/api";

	let accounts: SmtpAccount[] = [];
	let loading = true;
	let error: string | null = null;

	let modalOpen = false;
	let editing: SmtpAccount | null = null;
	let nameInput = "";
	let hostInput = "";
	let portInput = 587;
	let usernameInput = "";
	let passwordInput = "";
	let fromInput = "";
	let commentInput = "";
	let useTlsInput = false;
	let useStartTlsInput = true;

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

	onMount(async () => {
		await loadAccounts();
	});

	async function loadAccounts() {
		try {
			loading = true;
			error = null;
			const data = await listSmtpAccounts();
			accounts = Array.isArray(data) ? data : [];
		} catch (e) {
			error = e instanceof Error ? e.message : "Не удалось загрузить аккаунты";
			accounts = [];
		} finally {
			loading = false;
		}
	}

	function openModal(acc?: SmtpAccount) {
		editing = acc || null;
		nameInput = acc?.name || "";
		hostInput = acc?.host || "";
		portInput = acc?.port || 587;
		usernameInput = acc?.username || "";
		passwordInput = acc?.password || "";
		fromInput = acc?.from || "";
		commentInput = acc?.comment || "";
		useTlsInput = acc?.useTls ?? false;
		useStartTlsInput = acc?.useStartTls ?? true;
		modalOpen = true;
		error = null;
	}

	function closeModal() {
		modalOpen = false;
		editing = null;
		nameInput = "";
		hostInput = "";
		portInput = 587;
		usernameInput = "";
		passwordInput = "";
		fromInput = "";
		commentInput = "";
		useTlsInput = false;
		useStartTlsInput = true;
		error = null;
	}

	async function handleSave() {
		const host = hostInput.trim();
		const name = nameInput.trim();
		if (!host || !name || saving) return;
		const port = Number(portInput);
		if (!Number.isFinite(port) || port < 1 || port > 65535) {
			error = "Укажите порт от 1 до 65535";
			return;
		}

		try {
			saving = true;
			error = null;

			const payload = {
				name,
				host,
				port,
				username: usernameInput.trim(),
				password: passwordInput,
				from: fromInput.trim(),
				comment: commentInput.trim(),
				useTls: useTlsInput,
				useStartTls: useStartTlsInput,
			};

			if (editing) {
				const updated = await updateSmtpAccount(editing.id, payload);
				accounts = accounts.map((a) => (a.id === editing.id ? updated : a));
			} else {
				const created = await createSmtpAccount(payload);
				accounts = [...accounts, created];
			}

			closeModal();
		} catch (e) {
			error = e instanceof Error ? e.message : "Не удалось сохранить аккаунт";
		} finally {
			saving = false;
		}
	}

	async function handleDelete(id: string) {
		if (!confirm("Удалить этот SMTP-аккаунт?")) return;

		try {
			error = null;
			await deleteSmtpAccount(id);
			accounts = accounts.filter((a) => a.id !== id);
		} catch (e) {
			error = e instanceof Error ? e.message : "Не удалось удалить аккаунт";
		}
	}

	async function handleToggleActive(acc: SmtpAccount) {
		try {
			error = null;
			const updated = await toggleSmtpAccountActive(acc.id, !acc.isActive);
			accounts = accounts.map((a) => (a.id === acc.id ? updated : a));
		} catch (e) {
			error = e instanceof Error ? e.message : "Не удалось изменить статус";
		}
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
      <span class="pill">smtp connectors</span>
    </div>
    <p class="text-sm text-muted max-w-2xl">
      Несколько SMTP-аккаунтов для отправки писем: хост, порт, шифрование и учётные данные.
    </p>
  </header>

  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-lg font-semibold text-text">SMTP-аккаунты</h2>
      <button
        type="button"
        class="btn-primary bg-surfaceMuted text-text shadow-none hover:shadow-sm"
        on:click={() => openModal()}
      >
        Добавить аккаунт
      </button>
    </div>

    {#if error}
      <div class="glass-card p-4 bg-red-50 border-red-200">
        <p class="text-sm text-red-600">{error}</p>
      </div>
    {/if}

    {#if loading}
      <div class="glass-card p-8 text-center">
        <p class="text-sm text-muted">Загрузка...</p>
      </div>
    {:else if accounts.length === 0}
      <div class="glass-card p-8 text-center">
        <p class="text-sm text-muted">Нет добавленных SMTP-аккаунтов</p>
      </div>
    {:else}
      <div class="grid gap-4 md:grid-cols-2">
        {#each accounts as acc}
          <div class="glass-card p-4 space-y-3">
            <div class="flex items-start justify-between gap-3">
              <div class="flex-1 space-y-2 min-w-0">
                <div class="space-y-1">
                  <div class="flex items-center justify-between gap-2">
                    <p class="font-semibold text-text truncate">{acc.name || "Без названия"}</p>
                    <label class="relative inline-flex items-center cursor-pointer shrink-0">
                      <input
                        type="checkbox"
                        checked={acc.isActive}
                        on:change={() => handleToggleActive(acc)}
                        class="sr-only peer"
                      />
                      <div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-accent rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-accent"></div>
                      <span class="ml-2 text-xs text-muted">{acc.isActive ? "Активен" : "Неактивен"}</span>
                    </label>
                  </div>
                  <p class="text-xs font-mono text-muted break-all">
                    {acc.host}:{acc.port}
                    {#if acc.useTls}
                      <span class="ml-1 text-positive">TLS</span>
                    {:else if acc.useStartTls}
                      <span class="ml-1 text-accent">STARTTLS</span>
                    {/if}
                  </p>
                  {#if acc.from}
                    <p class="text-xs text-muted">From: {acc.from}</p>
                  {/if}
                  {#if acc.username}
                    <p class="text-xs text-muted">Логин: {acc.username}</p>
                  {/if}
                  {#if acc.password}
                    <p class="text-xs font-mono text-muted">Пароль: {maskSecret(acc.password)}</p>
                  {/if}
                </div>
                {#if acc.comment}
                  <p class="text-sm text-muted">{acc.comment}</p>
                {/if}
              </div>
              <div class="flex items-center gap-2 shrink-0">
                <button
                  type="button"
                  class="icon-btn"
                  title="Редактировать"
                  aria-label="Редактировать"
                  on:click={() => openModal(acc)}
                >
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="h-4 w-4">
                    <path d="M16.862 3.487 20.51 7.136a1.5 1.5 0 0 1 0 2.121l-9.193 9.193-4.593.511a1 1 0 0 1-1.1-1.1l.511-4.593 9.193-9.193a1.5 1.5 0 0 1 2.121 0Z" />
                    <path d="M19 11.5 12.5 5" />
                  </svg>
                </button>
                <button
                  type="button"
                  class="icon-btn"
                  title="Удалить"
                  aria-label="Удалить"
                  on:click={() => handleDelete(acc.id)}
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
            {editing ? 'Редактировать SMTP' : 'Новый SMTP-аккаунт'}
          </h3>
          <button type="button" class="modal-close" on:click={closeModal} aria-label="Закрыть">
            ✕
          </button>
        </div>
        <div class="modal-body space-y-4">
          <div class="space-y-1">
            <label class="text-sm font-medium text-text" for="smtp-name">Название</label>
            <input
              id="smtp-name"
              class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:border-accent focus:outline-none"
              bind:value={nameInput}
              placeholder="Например: Transactional"
              autocomplete="off"
            />
          </div>
          <div class="grid gap-4 sm:grid-cols-2">
            <div class="space-y-1">
              <label class="text-sm font-medium text-text" for="smtp-host">Хост</label>
              <input
                id="smtp-host"
                class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:border-accent focus:outline-none"
                bind:value={hostInput}
                placeholder="smtp.example.com"
                autocomplete="off"
              />
            </div>
            <div class="space-y-1">
              <label class="text-sm font-medium text-text" for="smtp-port">Порт</label>
              <input
                id="smtp-port"
                type="number"
                min="1"
                max="65535"
                class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:border-accent focus:outline-none"
                bind:value={portInput}
              />
            </div>
          </div>
          <div class="space-y-1">
            <label class="text-sm font-medium text-text" for="smtp-from">Адрес отправителя (From)</label>
            <input
              id="smtp-from"
              type="email"
              class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:border-accent focus:outline-none"
              bind:value={fromInput}
              placeholder="noreply@example.com"
              autocomplete="off"
            />
          </div>
          <div class="grid gap-4 sm:grid-cols-2">
            <div class="space-y-1">
              <label class="text-sm font-medium text-text" for="smtp-user">Логин</label>
              <input
                id="smtp-user"
                class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:border-accent focus:outline-none"
                bind:value={usernameInput}
                autocomplete="off"
              />
            </div>
            <div class="space-y-1">
              <label class="text-sm font-medium text-text" for="smtp-pass">Пароль</label>
              <input
                id="smtp-pass"
                type="password"
                class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:border-accent focus:outline-none"
                bind:value={passwordInput}
                placeholder={editing ? 'Оставьте пустым, чтобы не менять' : ''}
                autocomplete="new-password"
              />
            </div>
          </div>
          <div class="flex flex-wrap gap-4">
            <label class="inline-flex items-center gap-2 text-sm text-text cursor-pointer">
              <input type="checkbox" bind:checked={useTlsInput} class="rounded border-border" />
              Implicit TLS (обычно порт 465)
            </label>
            <label class="inline-flex items-center gap-2 text-sm text-text cursor-pointer">
              <input type="checkbox" bind:checked={useStartTlsInput} class="rounded border-border" />
              STARTTLS (обычно 587)
            </label>
          </div>
          <div class="space-y-1">
            <label class="text-sm font-medium text-text" for="smtp-comment">Комментарий</label>
            <textarea
              id="smtp-comment"
              class="h-20 w-full resize-none rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:border-accent focus:outline-none"
              bind:value={commentInput}
              placeholder="Внутренние пометки"
            ></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn-secondary" on:click={closeModal}>Отменить</button>
          <button
            type="button"
            class="btn-primary"
            on:click={handleSave}
            disabled={!hostInput.trim() || !nameInput.trim() || saving}
          >
            {saving ? 'Сохранение...' : editing ? 'Сохранить' : 'Добавить'}
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

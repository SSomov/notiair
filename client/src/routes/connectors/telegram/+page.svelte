<script lang="ts">
  type Note = {
    id: string;
    secret: string;
    comment: string;
  };

  let notes: Note[] = [];

  let modalOpen = false;
  let editingNote: Note | null = null;
  let secretInput = '';
  let commentInput = '';

  function openModal(note?: Note) {
    editingNote = note || null;
    secretInput = note?.secret || '';
    commentInput = note?.comment || '';
    modalOpen = true;
  }

  function closeModal() {
    modalOpen = false;
    editingNote = null;
    secretInput = '';
    commentInput = '';
  }

  function saveConnector() {
    if (!secretInput.trim()) return;

    if (editingNote) {
      notes = notes.map((note) =>
        note.id === editingNote.id
          ? { ...note, secret: secretInput.trim(), comment: commentInput.trim() }
          : note
      );
    } else {
      const nextIndex = notes.length + 1;
      const id = `#TG-${String(nextIndex).padStart(3, '0')}`;
      notes = [
        ...notes,
        {
          id,
          secret: secretInput.trim(),
          comment: commentInput.trim()
        }
      ];
    }

    closeModal();
  }

  function deleteNote(id: string) {
    notes = notes.filter((note) => note.id !== id);
  }
</script>

<section class="space-y-8 px-4 pb-12 pt-2 md:px-12 md:pt-4">
  <header class="space-y-2">
    <div class="flex items-center gap-3">
      <a href="/connectors" class="text-muted hover:text-text transition">
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="h-5 w-5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 19.5 8.25 12l7.5-7.5" />
        </svg>
      </a>
      <span class="pill">telegram connectors</span>
    </div>
    <p class="text-sm text-muted max-w-2xl">
      Управляйте токенами ботов и каналов Telegram для маршрутизации уведомлений.
    </p>
  </header>

  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-lg font-semibold text-text">Токены Telegram</h2>
      <button
        type="button"
        class="btn-primary bg-surfaceMuted text-text shadow-none hover:shadow-sm"
        on:click={() => openModal()}
      >
        Добавить токен
      </button>
    </div>

    {#if notes.length === 0}
      <div class="glass-card p-8 text-center">
        <p class="text-sm text-muted">Нет добавленных токенов</p>
      </div>
    {:else}
      <div class="grid gap-4 md:grid-cols-2">
        {#each notes as note}
          <div class="glass-card p-4 space-y-3">
            <div class="flex items-start justify-between gap-3">
              <div class="flex-1">
                <p class="font-semibold text-text">{note.id}</p>
                {#if note.comment}
                  <p class="mt-1 text-sm text-muted">{note.comment}</p>
                {/if}
                <p class="mt-2 text-xs text-muted font-mono truncate">{note.secret}</p>
              </div>
              <div class="flex items-center gap-2">
                <button
                  type="button"
                  class="icon-btn"
                  title="Редактировать"
                  aria-label="Редактировать"
                  on:click={() => openModal(note)}
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
                  on:click={() => deleteNote(note.id)}
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
      <div class="modal">
        <div class="modal-header">
          <h3 class="text-lg font-semibold text-text">
            {editingNote ? 'Редактировать токен' : 'Добавить токен Telegram'}
          </h3>
          <button type="button" class="modal-close" on:click={closeModal} aria-label="Закрыть">
            ✕
          </button>
        </div>
        <div class="modal-body space-y-4">
          <div class="space-y-1">
            <label class="text-sm font-medium text-text" for="secret-input">Bot API Token</label>
            <input
              id="secret-input"
              class="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:border-accent focus:outline-none"
              bind:value={secretInput}
              placeholder="Введите Bot API Token"
              autocomplete="off"
            />
          </div>
          <div class="space-y-1">
            <label class="text-sm font-medium text-text" for="comment-input">Комментарий</label>
            <textarea
              id="comment-input"
              class="h-24 w-full resize-none rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text focus:border-accent focus:outline-none"
              bind:value={commentInput}
              placeholder="Например: основной бот для статуса заказов"
            ></textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn-secondary" on:click={closeModal}>Отменить</button>
          <button
            type="button"
            class="btn-primary"
            on:click={saveConnector}
            disabled={!secretInput.trim()}
          >
            {editingNote ? 'Сохранить' : 'Добавить'}
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

<script lang="ts">
	import { draggable } from '@neodrag/svelte';
	import type { DragEventData } from '@neodrag/svelte';

	const triggerOptions = ['API', 'HTTP Webhook', 'Cron', 'Manual'];
	let triggerMenuOpen = false;
	let selectedTrigger = 'Добавить триггер';

	type NodeVariant = 'trigger' | 'action' | 'channel';

	type CanvasNode = {
		id: string;
		label: string;
		description: string;
		variant: NodeVariant;
		position: { x: number; y: number };
	};

	let nodes: CanvasNode[] = [
		{
			id: 'trigger-node',
			label: 'Trigger',
			description: 'Выберите событие запуска',
			variant: 'trigger',
			position: { x: 80, y: 140 },
		},
		{
			id: 'action-node',
			label: 'Action',
			description: 'Отправка уведомления',
			variant: 'action',
			position: { x: 360, y: 320 },
		},
		{
			id: 'channel-node',
			label: 'Channel',
			description: 'Telegram · @product-updates',
			variant: 'channel',
			position: { x: 640, y: 160 },
		},
	];

	function selectTrigger(option: string) {
		selectedTrigger = option;
		triggerMenuOpen = false;
		nodes = nodes.map((node) =>
			node.variant === 'trigger'
				? {
						...node,
						description: option === 'Добавить триггер' ? 'Выберите событие запуска' : option,
					}
				: node
		);
	}

	function handleDrag({ detail }: CustomEvent<DragEventData>, id: string) {
		nodes = nodes.map((node) =>
			node.id === id
				? {
						...node,
						position: {
							x: detail.offsetX,
							y: detail.offsetY,
						},
					}
				: node
		);
	}
</script>

<section class="space-y-8 px-4 pb-12 pt-2 md:px-12 md:pt-4">
	<header class="space-y-2">
		<span class="pill">новый workflow</span>
		<p class="max-w-2xl text-sm text-muted">
			Добавьте триггеры, действия и каналы доставки. Каждую ноду можно связать линиями и
			протестировать перед публикацией.
		</p>
	</header>

	<div class="workspace">
		<div class="workspace-toolbar">
			<div class="relative">
				<button
					type="button"
					class="btn-primary bg-surfaceMuted text-text shadow-none hover:shadow-sm"
					on:click={() => (triggerMenuOpen = !triggerMenuOpen)}
					aria-haspopup="true"
					aria-expanded={triggerMenuOpen}
				>
					{selectedTrigger}
				</button>
				{#if triggerMenuOpen}
					<ul
						class="absolute z-10 mt-2 w-48 overflow-hidden rounded-xl border border-border bg-surface shadow-lg"
					>
						{#each triggerOptions as option}
							<li>
								<button
									type="button"
									class="block w-full px-4 py-2 text-left text-sm text-text hover:bg-surfaceMuted"
									on:click={() => selectTrigger(option)}
								>
									{option}
								</button>
							</li>
						{/each}
					</ul>
				{/if}
			</div>

			<div class="flex items-center gap-3">
				<button
					type="button"
					class="btn-primary bg-surfaceMuted text-text shadow-none hover:shadow-sm"
				>
					Добавить шаблонизатор
				</button>
				<button
					type="button"
					class="btn-primary bg-surfaceMuted text-text shadow-none hover:shadow-sm"
				>
					Добавить канал
				</button>
				<button type="button" class="btn-primary">Запустить тест</button>
			</div>
		</div>

		<div class="workspace-content" id="workspace">
			{#each nodes as node (node.id)}
				<div
					class={`node ${node.variant}`}
					use:draggable={{
						position: node.position,
						bounds: '#workspace',
						grid: [12, 12],
					}}
					on:neodrag={(event) => handleDrag(event, node.id)}
				>
					<button type="button" class="edit-btn" aria-label="Редактировать ноду">
						<svg
							xmlns="http://www.w3.org/2000/svg"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="1.5"
							class="h-4 w-4"
						>
							<path
								d="M16.862 3.487 20.51 7.136a1.5 1.5 0 0 1 0 2.121l-9.193 9.193-4.593.511a1 1 0 0 1-1.1-1.1l.511-4.593 9.193-9.193a1.5 1.5 0 0 1 2.121 0Z"
							/>
							<path d="M19 11.5 12.5 5" />
						</svg>
					</button>
					<div class="connectors">
						<span class="connector right"></span>
						<span class="connector left"></span>
					</div>
					<span class="node-label">{node.label}</span>
					<p class="node-desc">{node.description}</p>
				</div>
			{/each}
		</div>
	</div>
</section>

<style>
	.workspace {
		position: relative;
		border-radius: 1.5rem;
		border: 1px solid var(--border, #e2e8f0);
		overflow: hidden;
		background-color: #fff;
		min-height: 700px;
	}

	.workspace::before {
		content: '';
		position: absolute;
		inset: 0;
		background-image:
			linear-gradient(to right, rgba(148, 163, 184, 0.12) 1px, transparent 1px),
			linear-gradient(to bottom, rgba(148, 163, 184, 0.12) 1px, transparent 1px);
		background-size: 18px 18px;
		opacity: 0.7;
	}

	.workspace-toolbar {
		position: relative;
		z-index: 2;
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		padding: 1.25rem 1.5rem;
		background: rgba(255, 255, 255, 0.85);
		border-bottom: 1px солид rgba(226, 232, 240, 0.7);
		backdrop-filter: blur(6px);
	}

	.workspace-content {
		position: relative;
		z-index: 1;
		min-height: 580px;
		padding: 3rem 2rem;
	}

	.node {
		position: absolute;
		width: 240px;
		border-radius: 1.25rem;
		border: 1px солид rgba(148, 163, 184, 0.3);
		padding: 1.5rem;
		background: rgba(248, 250, 252, 0.92);
		box-shadow: 0 18px 40px -24px rgba(37, 99, 235, 0.35);
		text-align: left;
		cursor: grab;
	}

	.node:global(.neodrag-dragging) {
		cursor: grabbing;
	}

	.edit-btn {
		position: absolute;
		top: 0.75rem;
		right: 0.75rem;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		border-radius: 999px;
		border: 1px солид rgba(148, 163, 184, 0.3);
		background: #fff;
		color: #64748b;
		box-shadow: 0 6px 16px -12px rgba(15, 23, 42, 0.4);
		transition: 120ms ease;
	}

	.edit-btn:hover {
		color: #2563eb;
		border-color: rgba(37, 99, 235, 0.5);
		transform: translateY(-1px);
	}

	.node-label {
		display: inline-flex;
		padding: 0.25rem 0.75rem;
		border-radius: 999px;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		форма: 0.16em;
		background: rgba(37, 99, 235, 0.1);
		color: #2563eb;
	}

	.node-desc {
		margin-top: 0.75rem;
		font-size: 0.875rem;
		color: #64748b;
	}

	.node.action {
		background: rgba(59, 130, 246, 0.08);
	}

	.node.channel {
		background: rgba(16, 185, 129, 0.12);
	}

	.connectors {
		position: absolute;
		inset: 0;
		pointer-events: none;
	}

	.connector {
		position: absolute;
		width: 10px;
		height: 10px;
		border-radius: 999px;
		background: #2563eb;
		border: 2px solid #fff;
		box-shadow: 0 0 0 2px rgba(37, 99, 235, 0.2);
	}

	.connector.right {
		top: 50%;
		right: -6px;
		transform: translateY(-50%);
	}

	.connector.left {
		top: 50%;
		left: -6px;
		transform: translateY(-50%);
	}
</style>

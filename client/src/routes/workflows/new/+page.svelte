<script lang="ts">
	import { draggable } from '@neodrag/svelte';
	import type { DragEventData } from '@neodrag/svelte';
	import { tick } from 'svelte';

	const triggerOptions = ['API', 'HTTP Webhook', 'Cron', 'Manual'];
	let triggerMenuOpen = false;
	let selectedTrigger = 'Добавить триггер';

	type NodeVariant = 'trigger' | 'action' | 'channel';
	type PortType = 'left' | 'right';

	type CanvasNode = {
		id: string;
		label: string;
		description: string;
		variant: NodeVariant;
		position: { x: number; y: number };
	};

	type Edge = {
		id: string;
		from: { nodeId: string; port: PortType };
		to: { nodeId: string; port: PortType };
	};

	type ConnectingState = {
		nodeId: string;
		port: PortType;
	} | null;

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

	let edges: Edge[] = [];
	let connecting: ConnectingState = null;
	let mousePosition: { x: number; y: number } | null = null;
	let workspaceElement: HTMLDivElement;

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

	async function handleDrag({ detail }: CustomEvent<DragEventData>, id: string) {
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
		// Ждем обновления DOM перед пересчетом линий
		await tick();
		// Принудительно обновляем реактивность для edgePaths
		nodes = [...nodes];
	}

	function getConnectorPosition(nodeId: string, port: PortType): { x: number; y: number } | null {
		const node = nodes.find((n) => n.id === nodeId);
		if (!node || !workspaceElement) return null;

		// Найти DOM элемент ноды и коннектора для точного расчета координат
		const nodeElement = document.querySelector(`[data-node-id="${nodeId}"]`) as HTMLElement;
		if (!nodeElement) return null;

		// Найти конкретный коннектор (left или right)
		const connectorElement = nodeElement.querySelector(`.connector.${port}`) as HTMLElement;
		if (!connectorElement) return null;

		// Получить координаты коннектора относительно viewport
		const connectorRect = connectorElement.getBoundingClientRect();
		const workspaceRect = workspaceElement.getBoundingClientRect();

		// Вычислить координаты центра коннектора (10px x 10px, центр на 5px от краев)
		const connectorCenterX = connectorRect.left + connectorRect.width / 2 - workspaceRect.left;
		const connectorCenterY = connectorRect.top + connectorRect.height / 2 - workspaceRect.top;

		return { x: connectorCenterX, y: connectorCenterY };
	}

	function handleConnectorClick(nodeId: string, port: PortType, event: MouseEvent) {
		event.stopPropagation();

		if (connecting) {
			// Завершаем соединение
			if (connecting.nodeId !== nodeId || connecting.port !== port) {
				// Не позволяем соединять точку с самой собой
				const edgeId = `${connecting.nodeId}-${connecting.port}-${nodeId}-${port}`;
				const newEdge: Edge = {
					id: edgeId,
					from: connecting,
					to: { nodeId, port },
				};

				// Проверяем, нет ли уже такого соединения
				const exists = edges.some(
					(e) =>
						(e.from.nodeId === newEdge.from.nodeId &&
							e.from.port === newEdge.from.port &&
							e.to.nodeId === newEdge.to.nodeId &&
							e.to.port === newEdge.to.port) ||
						(e.from.nodeId === newEdge.to.nodeId &&
							e.from.port === newEdge.to.port &&
							e.to.nodeId === newEdge.from.nodeId &&
							e.to.port === newEdge.from.port)
				);

				if (!exists) {
					edges = [...edges, newEdge];
				}
			}
			connecting = null;
		} else {
			// Начинаем новое соединение
			connecting = { nodeId, port };
		}
	}

	function cancelConnection() {
		connecting = null;
		mousePosition = null;
	}

	function handleWorkspaceMouseMove(event: MouseEvent) {
		if (connecting && workspaceElement) {
			const rect = workspaceElement.getBoundingClientRect();
			mousePosition = {
				x: event.clientX - rect.left,
				y: event.clientY - rect.top,
			};
		}
	}


	// Реактивное вычисление путей для линий
	// Зависит от nodes и edges, чтобы обновляться при перемещении блоков
	// Также зависит от размеров окна для корректной работы при изменении viewport
	let windowResizeTrigger = 0;
	
	if (typeof window !== 'undefined') {
		const handleResize = () => {
			windowResizeTrigger++;
		};
		window.addEventListener('resize', handleResize);
		// Cleanup будет выполнен при unmount компонента
	}

	$: edgePaths = (() => {
		// Явно используем nodes, edges и windowResizeTrigger для реактивности
		const _ = nodes.length + edges.length + windowResizeTrigger;
		return edges.map((edge) => {
			const fromPos = getConnectorPosition(edge.from.nodeId, edge.from.port);
			const toPos = getConnectorPosition(edge.to.nodeId, edge.to.port);
			if (!fromPos || !toPos) return null;

			const dx = toPos.x - fromPos.x;
			const controlX1 = fromPos.x + Math.abs(dx) * 0.5;
			const controlX2 = toPos.x - Math.abs(dx) * 0.5;

			return {
				id: edge.id,
				path: `M ${fromPos.x} ${fromPos.y} C ${controlX1} ${fromPos.y}, ${controlX2} ${toPos.y}, ${toPos.x} ${toPos.y}`,
			};
		}).filter((p): p is { id: string; path: string } => p !== null);
	})();

	// Путь для активного соединения (следует за курсором)
	// Зависит от nodes, connecting и mousePosition
	$: tempPath = (() => {
		if (!connecting) return null;

		const fromPos = getConnectorPosition(connecting.nodeId, connecting.port);
		if (!fromPos) return null;

		const targetX = mousePosition?.x ?? fromPos.x + 200;
		const targetY = mousePosition?.y ?? fromPos.y;

		const dx = targetX - fromPos.x;
		const controlX1 = fromPos.x + Math.abs(dx) * 0.5;
		const controlX2 = targetX - Math.abs(dx) * 0.5;

		return `M ${fromPos.x} ${fromPos.y} C ${controlX1} ${fromPos.y}, ${controlX2} ${targetY}, ${targetX} ${targetY}`;
	})();
</script>

<svelte:window on:keydown={(e) => e.key === 'Escape' && connecting && cancelConnection()} />

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

		<!-- svelte-ignore a11y-no-noninteractive-element-interactions -->
		<div
			class="workspace-content"
			id="workspace"
			bind:this={workspaceElement}
			on:click={cancelConnection}
			on:mousemove={handleWorkspaceMouseMove}
			role="application"
			aria-label="Рабочая область для создания workflow"
		>
			<!-- SVG слой для линий -->
			<svg class="edges-layer">
				{#each edgePaths as { id, path }}
					<path
						d={path}
						stroke="#2563eb"
						stroke-width="2"
						fill="none"
						class="edge-path"
					/>
				{/each}
				{#if tempPath}
					<path
						d={tempPath}
						stroke="#2563eb"
						stroke-width="2"
						stroke-dasharray="5,5"
						fill="none"
						opacity="0.5"
						class="temp-edge-path"
					/>
				{/if}
			</svg>

			{#each nodes as node (node.id)}
				<div
					data-node-id={node.id}
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
						<button
							type="button"
							class="connector right"
							class:active={connecting?.nodeId === node.id && connecting?.port === 'right'}
							on:click={(e) => handleConnectorClick(node.id, 'right', e)}
							aria-label="Правая точка подключения"
						></button>
						<button
							type="button"
							class="connector left"
							class:active={connecting?.nodeId === node.id && connecting?.port === 'left'}
							on:click={(e) => handleConnectorClick(node.id, 'left', e)}
							aria-label="Левая точка подключения"
						></button>
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
		letter-spacing: 0.16em;
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

	.edges-layer {
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		pointer-events: none;
		z-index: 0;
	}

	.edges-layer path {
		pointer-events: stroke;
		cursor: pointer;
	}

	.edge-path {
		transition: stroke 0.2s ease;
	}

	.edge-path:hover {
		stroke: #1d4ed8;
		stroke-width: 2.5;
	}

	.temp-edge-path {
		pointer-events: none;
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
		pointer-events: all;
		cursor: crosshair;
		transition: all 0.2s ease;
		padding: 0;
	}

	.connector:hover {
		background: #1d4ed8;
		transform: scale(1.2);
	}

	.connector.active {
		background: #dc2626;
		box-shadow: 0 0 0 3px rgba(220, 38, 38, 0.3);
	}

	.connector.right {
		top: 50%;
		right: -6px;
		transform: translateY(-50%);
	}

	.connector.right:hover {
		transform: translateY(-50%) scale(1.2);
	}

	.connector.left {
		top: 50%;
		left: -6px;
		transform: translateY(-50%);
	}

	.connector.left:hover {
		transform: translateY(-50%) scale(1.2);
	}
</style>

<script lang="ts">
	import { draggable } from "@neodrag/svelte";
	import type { DragEventData } from "@neodrag/svelte";
	import { onMount, tick } from "svelte";
	import { page } from "$app/stores";
	import { saveWorkflow, getWorkflow } from "$lib/api";
	import type { WorkflowDraft } from "$lib/types/workflow";

	const triggerOptions = ["API", "HTTP Webhook", "Cron", "Manual"];
	let triggerMenuOpen = false;
	let selectedTrigger = "Добавить триггер";

	let workflowId: string | null = null;
	let workflowName = "Новый workflow";
	let editingName = false;
	let isActive = false; // По умолчанию черновик
	let saving = false;
	let loading = false;
	let error: string | null = null;

type NodeVariant = "trigger" | "template" | "channel";
type PortType = "left" | "right";

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

	let nodes: CanvasNode[] = [];
	let edges: Edge[] = [];
let connecting: ConnectingState = null;
let mousePosition: { x: number; y: number } | null = null;
let workspaceElement: HTMLDivElement;

function selectTrigger(option: string) {
	selectedTrigger = option;
	triggerMenuOpen = false;
	nodes = nodes.map((node) =>
		node.variant === "trigger"
			? {
					...node,
					description:
						option === "Добавить триггер" ? "Выберите событие запуска" : option,
				}
			: node,
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
			: node,
	);
	// Ждем обновления DOM перед пересчетом линий
	await tick();
	// Принудительно обновляем реактивность для edgePaths
	nodes = [...nodes];
}

function getConnectorPosition(
	nodeId: string,
	port: PortType,
): { x: number; y: number } | null {
	const node = nodes.find((n) => n.id === nodeId);
	if (!node || !workspaceElement) return null;

	// Найти DOM элемент ноды и коннектора для точного расчета координат
	const nodeElement = document.querySelector(
		`[data-node-id="${nodeId}"]`,
	) as HTMLElement;
	if (!nodeElement) return null;

	// Найти конкретный коннектор (left или right)
	const connectorElement = nodeElement.querySelector(
		`.connector.${port}`,
	) as HTMLElement;
	if (!connectorElement) return null;

	// Получить координаты коннектора относительно viewport
	const connectorRect = connectorElement.getBoundingClientRect();
	const workspaceRect = workspaceElement.getBoundingClientRect();

	// Вычислить координаты центра коннектора (10px x 10px, центр на 5px от краев)
	const connectorCenterX =
		connectorRect.left + connectorRect.width / 2 - workspaceRect.left;
	const connectorCenterY =
		connectorRect.top + connectorRect.height / 2 - workspaceRect.top;

	return { x: connectorCenterX, y: connectorCenterY };
}

function handleConnectorClick(
	nodeId: string,
	port: PortType,
	event: MouseEvent,
) {
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
						e.to.port === newEdge.from.port),
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

function generateNodeId(variant: NodeVariant): string {
	const existingIds = nodes.map((n) => n.id);
	let counter = 1;
	let newId = `${variant}-node-${counter}`;
	while (existingIds.includes(newId)) {
		counter++;
		newId = `${variant}-node-${counter}`;
	}
	return newId;
}

function addChannelNode() {
	const channelCount = nodes.filter((n) => n.variant === "channel").length;
	const newChannel: CanvasNode = {
		id: generateNodeId("channel"),
		label: "Channel",
		description: "Новый канал доставки",
		variant: "channel",
		position: { x: 200 + channelCount * 280, y: 200 + channelCount * 80 },
	};
	nodes = [...nodes, newChannel];
}

function addTemplateNode() {
	const templateCount = nodes.filter((n) => n.variant === "template").length;
	const newTemplate: CanvasNode = {
		id: generateNodeId("template"),
		label: "Template",
		description: "Новый шаблон",
		variant: "template",
		position: { x: 150 + templateCount * 300, y: 180 + templateCount * 100 },
	};
	nodes = [...nodes, newTemplate];
}

function deleteNode(nodeId: string, event: MouseEvent) {
	event.stopPropagation();

	// Удалить узел
	nodes = nodes.filter((n) => n.id !== nodeId);

	// Удалить все связанные линии (edges)
	edges = edges.filter(
		(e) => e.from.nodeId !== nodeId && e.to.nodeId !== nodeId,
	);

	// Если удаляемый узел был в процессе соединения, отменить соединение
	if (connecting?.nodeId === nodeId) {
		connecting = null;
	}
}

function deleteEdge(edgeId: string, event?: MouseEvent | KeyboardEvent) {
	event?.stopPropagation();
	edges = edges.filter((e) => e.id !== edgeId);
}

function getMidpoint(pathString: string): { x: number; y: number } | null {
	// Парсим SVG path и находим точку на 50% длины
	try {
		const path = document.createElementNS("http://www.w3.org/2000/svg", "path");
		path.setAttribute("d", pathString);
		const length = path.getTotalLength();
		const midpoint = path.getPointAtLength(length / 2);
		return { x: midpoint.x, y: midpoint.y };
	} catch {
		return null;
	}
}

let hoveredEdgeId: string | null = null;

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

if (typeof window !== "undefined") {
	const handleResize = () => {
		windowResizeTrigger++;
	};
	window.addEventListener("resize", handleResize);
	// Cleanup будет выполнен при unmount компонента
}

$: edgePaths = (() => {
	// Явно используем nodes, edges и windowResizeTrigger для реактивности
	const _ = nodes.length + edges.length + windowResizeTrigger;
	return edges
		.map((edge) => {
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
		})
		.filter((p): p is { id: string; path: string } => p !== null);
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

	function toggleEditName() {
		editingName = !editingName;
	}

	function saveName() {
		if (!workflowName.trim()) {
			workflowName = "Новый workflow";
		}
		editingName = false;
	}

	onMount(async () => {
		const id = $page.url.searchParams.get("id");
		if (id) {
			workflowId = id;
			try {
				loading = true;
				error = null;
				const workflow = await getWorkflow(id);
				
				workflowName = workflow.name || "Новый workflow";
				isActive = workflow.isActive || false;
				
				// Преобразуем nodes из API формата в CanvasNode
				nodes = (workflow.nodes || []).map((node) => ({
					id: node.id,
					label: (node.config as any)?.label || node.id,
					description: (node.config as any)?.description || "",
					variant: ((node.config as any)?.variant || node.type) as NodeVariant,
					position: node.position || { x: 0, y: 0 },
				}));
				
				// Ждем, пока nodes отрендерятся, перед загрузкой edges
				await tick();
				// Дополнительная задержка для гарантии рендера
				await new Promise(resolve => setTimeout(resolve, 100));
				
				// Преобразуем edges из API формата в Edge
				edges = (workflow.edges || []).map((edge, index) => ({
					id: `${edge.from}-${edge.to}-${index}`,
					from: { nodeId: edge.from, port: "right" as PortType },
					to: { nodeId: edge.to, port: "left" as PortType },
				}));
				
				// Принудительно обновляем реактивность
				edges = [...edges];
				nodes = [...nodes];
			} catch (e) {
				error = e instanceof Error ? e.message : "Не удалось загрузить workflow";
			} finally {
				loading = false;
			}
		}
	});

	async function saveWorkflowToAPI() {
		if (saving) return;

		try {
			saving = true;
			error = null;

			// Преобразуем edges, убеждаясь что они валидны
			const edgesData = edges
				.filter((edge) => edge.from?.nodeId && edge.to?.nodeId)
				.map((edge) => ({
					from: edge.from.nodeId,
					to: edge.to.nodeId,
				}));

			const workflowData = {
				id: workflowId || crypto.randomUUID(),
				name: workflowName.trim() || "Новый workflow",
				description: "",
				nodes: nodes.map((node) => ({
					id: node.id,
					type: node.variant === "trigger" ? "trigger" : "action",
					position: {
						x: node.position.x,
						y: node.position.y,
					},
					config: {
						label: node.label,
						description: node.description,
						variant: node.variant,
					},
				})),
				edges: edgesData,
				filters: {},
				isActive: isActive,
			};

			await saveWorkflow(workflowData);
			// Обновляем workflowId после сохранения
			if (!workflowId) {
				const saved = await getWorkflow(workflowData.id);
				workflowId = saved.id;
			}
		} catch (e) {
			error = e instanceof Error ? e.message : "Не удалось сохранить workflow";
		} finally {
			saving = false;
		}
	}
</script>

<svelte:window on:keydown={(e) => e.key === 'Escape' && connecting && cancelConnection()} />

<section class="space-y-8 px-4 pb-12 pt-2 md:px-12 md:pt-4">
	<header class="space-y-2">
		<div class="flex items-center gap-3">
			<span class="pill">{workflowId ? "редактирование workflow" : "новый workflow"}</span>
			<div class="flex items-center gap-2">
				{#if editingName}
					<input
						type="text"
						class="rounded-lg border border-border bg-surface px-3 py-1.5 text-lg font-semibold text-text focus:border-accent focus:outline-none"
						bind:value={workflowName}
						on:blur={saveName}
						on:keydown={(e) => e.key === 'Enter' && saveName()}
						autofocus
					/>
				{:else}
					<h1 class="text-lg font-semibold text-text">{workflowName}</h1>
					<button
						type="button"
						class="icon-btn"
						title="Редактировать название"
						aria-label="Редактировать название"
						on:click={toggleEditName}
					>
						<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="h-4 w-4">
							<path d="M16.862 3.487 20.51 7.136a1.5 1.5 0 0 1 0 2.121l-9.193 9.193-4.593.511a1 1 0 0 1-1.1-1.1l.511-4.593 9.193-9.193a1.5 1.5 0 0 1 2.121 0Z" />
							<path d="M19 11.5 12.5 5" />
						</svg>
					</button>
				{/if}
			</div>
			<div class="flex items-center gap-2 ml-auto">
				<label class="relative inline-flex items-center cursor-pointer">
					<input
						type="checkbox"
						bind:checked={isActive}
						class="sr-only peer"
					/>
					<div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-accent rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-accent"></div>
					<span class="ml-2 text-sm text-muted">{isActive ? "Активен" : "Черновик"}</span>
				</label>
			</div>
		</div>
		<p class="max-w-2xl text-sm text-muted">
			Добавьте триггеры, действия и каналы доставки. Каждую ноду можно связать линиями и
			протестировать перед публикацией.
		</p>
		{#if error}
			<div class="rounded-lg border border-red-200 bg-red-50 p-3">
				<p class="text-sm text-red-600">{error}</p>
			</div>
		{/if}
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
					on:click={addTemplateNode}
				>
					Добавить шаблонизатор
				</button>
				<button
					type="button"
					class="btn-primary bg-surfaceMuted text-text shadow-none hover:shadow-sm"
					on:click={addChannelNode}
				>
					Добавить канал
				</button>
				<button type="button" class="btn-primary">Запустить тест</button>
				<button
					type="button"
					class="btn-primary bg-accent text-white shadow-sm hover:shadow-md"
					on:click={saveWorkflowToAPI}
					disabled={saving}
				>
					{saving ? 'Сохранение...' : 'Сохранить'}
				</button>
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
					<!-- svelte-ignore a11y-no-static-element-interactions -->
					<g
						class="edge-group"
						role="group"
						on:mouseenter={() => (hoveredEdgeId = id)}
						on:mouseleave={() => (hoveredEdgeId = null)}
					>
						<path
							d={path}
							stroke="#2563eb"
							stroke-width="2"
							fill="none"
							class="edge-path"
						/>
						{#if hoveredEdgeId === id}
							{@const midpoint = getMidpoint(path)}
							{#if midpoint}
								<!-- svelte-ignore a11y-click-events-have-key-events -->
								<!-- svelte-ignore a11y-no-static-element-interactions -->
								<g
									class="edge-delete-button"
									role="button"
									tabindex="0"
									transform="translate({midpoint.x}, {midpoint.y})"
									on:click={(e) => deleteEdge(id, e)}
									on:keydown={(e) => e.key === 'Enter' && deleteEdge(id)}
								>
									<circle
										cx="0"
										cy="0"
										r="10"
										fill="#fff"
										stroke="#dc2626"
										stroke-width="2"
										class="edge-delete-circle"
									/>
									<line
										x1="-4"
										y1="-4"
										x2="4"
										y2="4"
										stroke="#dc2626"
										stroke-width="1.5"
										stroke-linecap="round"
									/>
									<line
										x1="4"
										y1="-4"
										x2="-4"
										y2="4"
										stroke="#dc2626"
										stroke-width="1.5"
										stroke-linecap="round"
									/>
								</g>
							{/if}
						{/if}
					</g>
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
					<button
						type="button"
						class="delete-btn"
						aria-label="Удалить блок"
						on:click={(e) => deleteNode(node.id, e)}
					>
						<svg
							xmlns="http://www.w3.org/2000/svg"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							stroke-width="1.5"
							class="h-4 w-4"
						>
							<path
								d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
							/>
							<line x1="10" y1="11" x2="10" y2="17" />
							<line x1="14" y1="11" x2="14" y2="17" />
						</svg>
					</button>
					<div class="connectors">
						{#if node.variant !== 'channel'}
							<button
								type="button"
								class="connector right"
								class:active={connecting?.nodeId === node.id && connecting?.port === 'right'}
								on:click={(e) => handleConnectorClick(node.id, 'right', e)}
								aria-label="Правая точка подключения"
							></button>
						{/if}
						{#if node.variant !== 'trigger'}
							<button
								type="button"
								class="connector left"
								class:active={connecting?.nodeId === node.id && connecting?.port === 'left'}
								on:click={(e) => handleConnectorClick(node.id, 'left', e)}
								aria-label="Левая точка подключения"
							></button>
						{/if}
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

	.delete-btn {
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

	.delete-btn:hover {
		color: #dc2626;
		border-color: rgba(220, 38, 38, 0.5);
		background: rgba(220, 38, 38, 0.1);
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

	.node.template {
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

	.edge-group {
		cursor: pointer;
	}

	.edge-path {
		transition: stroke 0.2s ease;
	}

	.edge-group:hover .edge-path {
		stroke: #1d4ed8;
		stroke-width: 2.5;
	}

	.edge-delete-button {
		cursor: pointer;
		pointer-events: all;
	}

	.edge-delete-circle {
		transition: all 0.2s ease;
	}

	.edge-delete-button:hover .edge-delete-circle {
		fill: #fee2e2;
		stroke: #b91c1c;
		transform: scale(1.1);
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

	.icon-btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
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

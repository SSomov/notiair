# Telegram Workflow Client

SvelteKit-приложение под Bun для визуального конструирования workflow, управления шаблонами и мониторинга очередей уведомлений.

## Основные разделы

- **Template Studio** — редактор Telegram-шаблонов с превью и подсветкой переменных.
- **Workflow Builder** — визуальные диаграммы (узлы/стрелки) с фильтрами, шагами и условиями.
- **Queue Monitor** — статус задач, ретраи, ручной перезапуск.
- **Audit Trail** — история изменений шаблонов и workflow.

## Стек

- SvelteKit + Bun
- dnd-kit / svelte-dnd-action для drag & drop в визуальном редакторе
- Tailwind CSS + DaisyUI для UI-компонентов
- Zustand-подобный store (svelte/store) для клиентского состояния

## Запуск (dev)

```bash
bun install
bun dev
```

## Интеграция с API

- `GET /v1/templates` / `POST /v1/templates`
- `GET /v1/workflows` / `POST /v1/workflows`
- `POST /v1/notifications/preview` — локальный прогон workflow без отправки.
- `GET /v1/queues/pending`

UI оперирует теми же контрактами, что и backend, обеспечивая последовательность данных. Визуальный редактор сохраняет JSON workflow в формате, который напрямую понимает API.

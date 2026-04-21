<div align="center">
  <img src="docs/img/logo.svg" alt="NotiAir" width="88" height="88"/>
  <p class="text-xs font-semibold uppercase tracking-[0.45em] text-accent" style="margin:0.25rem 0 0;color:#2563eb">NotiAir</p>
  <p class="text-xs uppercase tracking-[0.3em] text-muted" style="margin:0.35rem 0 0;color:#64748b;font-size:0.75rem;letter-spacing:0.3em;text-transform:uppercase">Notification Service</p>
  <p style="margin:0.75rem 0 0;max-width:36rem;line-height:1.45;color:#475569;">Notification orchestration with visual workflows, message templates, and reliable async delivery.</p>
  <p>
    <a href="./README_ru.md">Русский</a>
  </p>
</div>

---

## Overview

NotiAir is a platform for designing and running notification pipelines: message templates with variables, a visual workflow builder (conditions, filters, routing), and asynchronous delivery through channels such as Telegram—with an outbox-backed path for dependable dispatch.

## Features

- **Templates** — Variables, Markdown-friendly content, live preview with sample payloads
- **Workflow builder** — Drag-and-drop graph: triggers, filters, actions, conditional routing
- **Queues** — Asynq on Redis, retries, task visibility
- **Delivery** — Outbox pattern for at-least-once style reliability; Telegram Bot API included; architecture ready for more channels
- **Connectors** — Manage channel connectors (e.g. Telegram) from the UI

## Architecture

| Layer | Role |
|--------|------|
| **Backend (Go)** | REST API (Fiber), PostgreSQL, Asynq/Redis workers, outbox, Telegram transport |
| **Frontend (SvelteKit)** | Workflow editor, template editor, queue monitoring, connector management |

## Tech stack

**Backend:** Go 1.24 · Fiber v2 · GORM · PostgreSQL · Asynq · Telegram Bot API  

**Frontend:** SvelteKit · TypeScript · Tailwind CSS · Bun · `@neodrag/svelte`

## Screenshots

| Main | Workflows |
|------|-----------|
| ![Main](docs/img/main.png) | ![Workflows](docs/img/workflows.png) |

| Workflow editor | Connectors |
|-----------------|------------|
| ![Workflow editor](docs/img/workflow.png) | ![Connectors](docs/img/connectors.png) |

| Connector setup |
|-----------------|
| ![Connector](docs/img/connector.png) |

## Quick start

### Prerequisites

- Go 1.24+
- PostgreSQL 14+
- Redis 6+
- [Bun](https://bun.sh/) (frontend)
- Docker & Docker Compose (optional, for local infra)

### 1. Infrastructure

```bash
cd .ops
docker compose up -d
```

### 2. Backend

```bash
cd api
cp .env.example .env
```

Edit `.env` (example):

```env
HTTP_ADDR=:8080
QUEUE_URL=redis://localhost:6379/0
QUEUE_NAMESPACE=notiair
TELEGRAM_BOT_TOKEN=your_bot_token
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=notiair
```

```bash
go run ./main.go
```

### 3. Frontend

```bash
cd client
bun install
bun dev
```

The app is served by Vite (typically `http://localhost:5173`).

## API (v1)

| Area | Method & path | Description |
|------|----------------|-------------|
| Notifications | `POST /v1/notifications/dispatch` | Dispatch through a workflow |
| Templates | `GET/POST /v1/templates` | List and upsert templates |
| Workflows | `GET/POST /v1/workflows` | List and upsert workflows |
| Queues | `GET /v1/queues/pending` | Pending queue tasks |

## Development

Make targets are defined under `.ops/Makefile`. From the repo root use `make -C .ops <target>` (recipes assume the working directory is `.ops` when needed).

| Target | Description |
|--------|----------------|
| `build-api` | Docker image `notiair:api-local` (`Dockerfile.api`, repo root as context) |
| `build-client` | Docker image `notiair:client-local` (`Dockerfile.client`) |
| `build-all` | `build-api` then `build-client` |
| `dev-api` | `go run ./main.go` in `api/` |
| `dev-client` | `bun dev` in `client/` |
| `dev` | `compose.dev.yml` dependencies, then API + client; stops compose on exit |

```bash
make -C .ops build-all
make -C .ops dev-api
make -C .ops dev
```

### Repository layout

```
notiair/
├── api/                 # Backend
│   ├── handlers/        # HTTP handlers
│   ├── internal/        # config, persistence, queue, routing, templates, transport, workflow
│   ├── routes/
│   ├── services/
│   └── main.go
└── client/
    └── src/
        ├── lib/         # api client, components, stores, types
        └── routes/      # templates, workflows, queues, connectors (e.g. connectors/telegram)
```

**Backend modules (short):** `internal/config` · `internal/persistence/*` · `internal/routing` · `internal/queue` · `services/` · `handlers/` · `routes/`  

**Frontend:** `lib/api` · `lib/components` · `lib/stores` · `lib/types` · `routes/*`

# NotiAir

Система управления уведомлениями с поддержкой визуального конструирования workflow, шаблонов сообщений и асинхронной доставки через различные каналы связи.

## Описание

NotiAir — это платформа для создания и управления сложными сценариями отправки уведомлений. Система позволяет:

- Создавать шаблоны сообщений с переменными
- Конструировать workflow с условиями, фильтрами и маршрутизацией
- Отправлять уведомления через Telegram и другие каналы
- Мониторить очереди задач и управлять доставкой
- Использовать паттерн Outbox для гарантированной доставки

## Архитектура

Проект состоит из двух основных компонентов:

### Backend API (Go)
- REST API на базе Fiber
- PostgreSQL для хранения данных
- Asynq (Redis) для асинхронной обработки задач
- Outbox pattern для надежной доставки
- Поддержка Telegram Bot API

### Frontend Client (SvelteKit)
- Визуальный редактор workflow с drag & drop
- Редактор шаблонов с превью
- Мониторинг очередей и задач
- Управление каналами и коннекторами

## Технологический стек

### Backend
- **Go 1.24** — основной язык
- **Fiber v2** — HTTP фреймворк
- **GORM** — ORM для работы с БД
- **PostgreSQL** — основная БД
- **Asynq** — очередь задач на базе Redis
- **Telegram Bot API** — интеграция с Telegram

### Frontend
- **SvelteKit** — фреймворк
- **TypeScript** — типизация
- **Tailwind CSS** — стилизация
- **Bun** — runtime и пакетный менеджер
- **@neodrag/svelte** — drag & drop для workflow редактора

## Структура проекта

```
notiair/
├── api/                    # Backend API
│   ├── handlers/          # HTTP обработчики
│   ├── internal/          # Внутренние модули
│   │   ├── config/        # Конфигурация
│   │   ├── persistence/   # Репозитории (database, outbox, serviceconfig)
│   │   ├── queue/         # Клиент очереди (Asynq)
│   │   ├── routing/       # Сервис маршрутизации workflow
│   │   ├── templates/     # Доменная модель шаблонов
│   │   ├── transport/     # Транспорты (HTTP, Telegram)
│   │   └── workflow/      # Доменная модель workflow
│   ├── routes/            # Регистрация маршрутов
│   ├── services/          # Бизнес-логика
│   └── main.go           # Точка входа
│
└── client/                # Frontend приложение
    └── src/
        ├── lib/
        │   ├── api/       # API клиент
        │   ├── components/# UI компоненты
        │   ├── stores/    # Svelte stores
        │   └── types/     # TypeScript типы
        └── routes/        # SvelteKit страницы
            ├── templates/ # Управление шаблонами
            ├── workflows/ # Управление workflow
            ├── queues/    # Мониторинг очередей
            ├── channels/  # Управление каналами
            └── connectors/# Управление коннекторами
```

## Быстрый старт

### Требования
- Go 1.24+
- PostgreSQL 14+
- Redis 6+
- Bun (для frontend)
- Docker и Docker Compose (опционально, для инфраструктуры)

### Запуск инфраструктуры

```bash
cd ops
docker compose up -d
```

### Настройка Backend

1. Перейдите в директорию API:
```bash
cd api
```

2. Скопируйте файл с переменными окружения:
```bash
cp .env.example .env
```

3. Настройте переменные окружения в `.env`:
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

4. Запустите API:
```bash
go run ./main.go
```

### Настройка Frontend

1. Перейдите в директорию клиента:
```bash
cd client
```

2. Установите зависимости:
```bash
bun install
```

3. Запустите dev сервер:
```bash
bun dev
```

Frontend будет доступен по адресу `http://localhost:5173` (или другому порту, указанному Vite).

## API Endpoints

### Уведомления
- `POST /v1/notifications/dispatch` — отправка уведомления через workflow

### Шаблоны
- `GET /v1/templates` — список всех шаблонов
- `POST /v1/templates` — создание/обновление шаблона

### Workflow
- `GET /v1/workflows` — список всех workflow
- `POST /v1/workflows` — создание/обновление workflow

### Очереди
- `GET /v1/queues/pending` — список задач в очереди

## Основные возможности

### Шаблоны сообщений
- Создание шаблонов с переменными
- Поддержка Markdown форматирования
- Превью шаблонов с подстановкой переменных

### Workflow Builder
- Визуальное конструирование сценариев
- Узлы типов: trigger, filter, action
- Условная маршрутизация
- Фильтрация по параметрам payload

### Асинхронная доставка
- Очередь задач на базе Asynq/Redis
- Retry механизм для неудачных отправок
- Outbox pattern для гарантированной доставки
- Мониторинг статусов задач

### Интеграции
- Telegram Bot API
- Расширяемая архитектура для добавления новых каналов

## Разработка

### Структура модулей Backend

- `internal/config` — загрузка конфигурации из переменных окружения
- `internal/persistence/database` — подключение к PostgreSQL
- `internal/persistence/outbox` — таблица исходящих сообщений (Outbox pattern)
- `internal/persistence/serviceconfig` — конфигурации сервисов (тип, дефолтные настройки, активность)
- `internal/templates` — доменная модель шаблонов (в памяти)
- `internal/workflow` — доменная модель workflow (в памяти)
- `internal/routing` — сервис разрешения целей по workflow
- `internal/queue` — клиент очереди Asynq
- `services/` — бизнес-логика (NotificationService)
- `handlers/` — HTTP обработчики
- `routes/` — регистрация маршрутов

### Структура Frontend

- `lib/api/` — клиент для взаимодействия с API
- `lib/components/` — переиспользуемые компоненты (WorkflowCanvas и др.)
- `lib/stores/` — Svelte stores для управления состоянием
- `lib/types/` — TypeScript типы и интерфейсы
- `routes/` — страницы приложения (SvelteKit file-based routing)

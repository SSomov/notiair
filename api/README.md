# Telegram Notification API

Сервис управляет шаблонами, workflow и отправкой уведомлений в каналы/группы Telegram.

## Запуск инфраструктуры
```bash
cd ops
docker compose up -d
```

## Переменные окружения
Скопируйте `api/.env.example` в `api/.env` и задайте значения токенов и БД.

| Переменная | Описание |
|------------|----------|
| `HTTP_ADDR` | адрес HTTP сервера |
| `QUEUE_URL` | строка подключения к Redis/Asynq |
| `QUEUE_NAMESPACE` | имя очереди |
| `TELEGRAM_BOT_TOKEN` | токен Telegram-бота |
| `DB_*` | параметры подключения к Postgres |

## Структура модулей
- `internal/config` — загрузка конфигурации
- `internal/persistence/database` — подключение к БД
- `internal/persistence/outbox` — таблица исходящих сообщений
- `internal/persistence/serviceconfig` — конфигурации сервисов (type, default, isActive)
- `internal/templates`, `internal/workflow` — доменные сущности
- `services/` — бизнес-логика
- `handlers/` — HTTP-обработчики
- `routes/` — регистрация маршрутов

## Запуск API (dev)
```bash
cd api
bun install # для клиента, если нужен UI
cp .env.example .env
go run ./main.go
```


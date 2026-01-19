.PHONY: build-api build-client build-all dev dev-api dev-client

# Локальная сборка API (из корня проекта)
build-api:
	cd .. && docker build --pull -t clash:api-local -f .ops/Dockerfile.api .

# Локальная сборка Client (из корня проекта)
build-client:
	cd .. && docker build --pull -t clash:client-local -f .ops/Dockerfile.client .

# Сборка обоих образов
build-all: build-api build-client

# Запуск API в dev режиме
dev-api:
	cd ../api && go run ./main.go

# Запуск Client в dev режиме
dev-client:
	cd ../client && bun dev

# Запуск API и Client вместе в dev режиме
dev:
	@echo "Запуск docker-compose, API и Client в dev режиме..."
	@trap 'cd .. && docker compose -f .ops/compose.dev.yml down; kill 0' EXIT; \
	cd .. && docker compose -f .ops/compose.dev.yml up -d && \
	cd api && go run ./main.go & \
	cd ../client && bun dev & \
	wait

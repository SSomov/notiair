.PHONY: build-api build-client build-all dev dev-api dev-client

_OPS_MK := $(abspath $(lastword $(MAKEFILE_LIST)))
OPS_DIR := $(dir $(_OPS_MK))
ROOT := $(abspath $(OPS_DIR)/..)

# Локальная сборка API (контекст — корень репозитория; cwd не важен)
build-api:
	docker build --pull -t notiair:api-local -f $(OPS_DIR)Dockerfile.api $(ROOT)

# Локальная сборка Client
build-client:
	docker build --pull -t notiair:client-local -f $(OPS_DIR)Dockerfile.client $(ROOT)

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

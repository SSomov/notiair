.PHONY: build-api build-client build-all

# Локальная сборка API (из корня проекта)
build-api:
	cd .. && docker build --pull -t clash:api-local -f .ops/Dockerfile.api .

# Локальная сборка Client (из корня проекта)
build-client:
	cd .. && docker build --pull -t clash:client-local -f .ops/Dockerfile.client .

# Сборка обоих образов
build-all: build-api build-client


.PHONY: help db.up db.migrate local.run docker.full.run clean test

# Настройки по умолчанию
DB_NAME ?= posts_comments_db
DB_USER ?= postgres
DB_PASS ?= 1234567890qwe
DB_PORT ?= 5432
APP_PORT ?= 8080
STORE_TYPE ?= postgres

help:
	@echo "Posts & Comments Service Management"
	@echo ""
	@echo "Usage:"
	@echo "  make db.up           - Запустить только PostgreSQL в Docker"
	@echo "  make db.migrate      - Применить миграции (требует запущенную БД)"
	@echo "  make local.run       - Запустить сервер локально (STORE_TYPE=postgres|memory)"
	@echo "  make docker.full.run - Полный запуск в Docker (с БД и миграциями)"
	@echo "  make clean           - Остановить и удалить все контейнеры"
	@echo "  make test            - Запустить тесты"
	@echo ""
	@echo "Примеры:"
	@echo "  make local.run STORE_TYPE=memory"
	@echo "  make docker.full.run STORE_TYPE=postgres"

# Запуск только PostgreSQL в Docker
db.up:
	docker-compose up -d db
	@echo "PostgreSQL запущен на localhost:${DB_PORT}"
	@echo "Данные: ${DB_USER}:${DB_PASS}@localhost:${DB_PORT}/${DB_NAME}"

# Применение миграций через Docker
db.migrate:
	@if [ "${STORE_TYPE}" = "postgres" ]; then \
		docker-compose run --rm migrate; \
	else \
		echo "Миграции не требуются для STORE_TYPE=memory"; \
	fi

# Локальный запуск сервера (без Docker)
local.run:
	@echo "Запуск сервера с хранилищем: ${STORE_TYPE}"
	go run cmd/server/main.go \
		-store ${STORE_TYPE} \
		$(if $(filter $(STORE_TYPE),postgres),-dsn "postgres://${DB_USER}:${DB_PASS}@localhost:${DB_PORT}/${DB_NAME}?sslmode=disable") \
		-port ${APP_PORT}

# Полный запуск в Docker (app + db + migrations)
docker.full.run:
	@echo "Запуск полного стека с хранилищем: ${STORE_TYPE}"
	STORE_TYPE=${STORE_TYPE} docker-compose up -d --build

	@if [ "${STORE_TYPE}" = "postgres" ]; then \
		echo "Ожидание готовности PostgreSQL..."; \
		while ! docker-compose exec db pg_isready -U ${DB_USER} -d ${DB_NAME}; do sleep 1; done; \
		echo "Применение миграций..."; \
		docker-compose run --rm migrate; \
	fi

	@echo ""
	@echo "Сервис запущен:"
	@echo "  - API: http://localhost:${APP_PORT}"
	@echo "  - Хранилище: ${STORE_TYPE}"
	@if [ "${STORE_TYPE}" = "postgres" ]; then \
		echo "  - PostgreSQL: postgres://${DB_USER}:*****@localhost:${DB_PORT}/${DB_NAME}"; \
	fi

# Остановка и очистка
clean:
	docker-compose down -v
	rm -f server
	@echo "Все контейнеры и volumes удалены"

# Запуск тестов
test:
	go test ./... -v -cover
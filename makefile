.PHONY: help
help:
	@echo "Musical-Zoe Service - Available Commands:"
	@echo ""
	@echo "Development:"
	@echo "  run/api             - run the API application"
	@echo "  dev/setup           - setup development environment (Docker + DB)"
	@echo "  dev/start           - start development services"
	@echo "  dev/stop            - stop development services"
	@echo "  dev/reset           - reset development environment (removes data)"
	@echo ""
	@echo "Testing:"
	@echo "  test                - run all tests"
	@echo "  test/unit           - run unit tests only"
	@echo "  test/coverage       - run tests with coverage report"
	@echo ""
	@echo "Database:"
	@echo "  db/migrate/up       - run database migrations"
	@echo "  db/migrate/down     - rollback database migrations"
	@echo "  db/migrate/status   - show migration status"
	@echo "  db/connect          - connect to development database"
	@echo ""
	@echo "Docker & Deployment:"
	@echo "  docker/dev          - start development environment"
	@echo "  docker/logs         - view container logs"

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@echo 'Running cmd/api...'
	go run ./cmd/api

## dev/setup: setup complete development environment
.PHONY: dev/setup
dev/setup:
	@echo 'Setting up development environment...'
	./scripts/dev.sh

## dev/start: start development services
.PHONY: dev/start
dev/start:
	@echo 'Starting development services...'
	docker-compose up -d

## dev/stop: stop development services
.PHONY: dev/stop
dev/stop:
	@echo 'Stopping development services...'
	docker-compose down

## dev/reset: reset development environment (removes all data)
.PHONY: dev/reset
dev/reset:
	@echo 'Resetting development environment...'
	docker-compose down -v
	./scripts/dev.sh

## db/migrate/up: run database migrations
.PHONY: db/migrate/up
db/migrate/up:
	@echo 'Running database migrations...'
	@if [ -f cmd/api/.env ]; then \
		. cmd/api/.env && \
		goose -dir internal/sql/schema postgres "$$MUSICALZOE_DB_DSN" up; \
	else \
		echo "Error: cmd/api/.env file not found"; \
		exit 1; \
	fi

## db/migrate/down: rollback database migrations
.PHONY: db/migrate/down
db/migrate/down:
	@echo 'Rolling back database migrations...'
	@if [ -f cmd/api/.env ]; then \
		. cmd/api/.env && \
		goose -dir internal/sql/schema postgres "$$MUSICALZOE_DB_DSN" down; \
	else \
		echo "Error: cmd/api/.env file not found"; \
		exit 1; \
	fi

## db/migrate/status: show migration status
.PHONY: db/migrate/status
db/migrate/status:
	@echo 'Checking migration status...'
	@if [ -f cmd/api/.env ]; then \
		. cmd/api/.env && \
		goose -dir internal/sql/schema postgres "$$MUSICALZOE_DB_DSN" status; \
	else \
		echo "Error: cmd/api/.env file not found"; \
		exit 1; \
	fi

## db/connect: connect to development database
.PHONY: db/connect
db/connect:
	@echo 'Connecting to development database...'
	@if [ -f cmd/api/.env ]; then \
		. cmd/api/.env && \
		PGPASSWORD="$$POSTGRES_PASSWORD" psql -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" -h localhost; \
	else \
		echo "Error: cmd/api/.env file not found"; \
		exit 1; \
	fi

## docker/dev: start development environment (alias for dev/setup)
.PHONY: docker/dev
docker/dev: dev/setup

## docker/logs: view container logs
.PHONY: docker/logs
docker/logs:
	@echo 'Viewing container logs...'
	docker-compose logs -f

## test: run all tests
.PHONY: test
test:
	@echo 'Running tests...'
	go test -v ./cmd/api/...

## test/unit: run unit tests only
.PHONY: test/unit
test/unit:
	@echo 'Running unit tests...'
	go test -v -short ./cmd/api/...

## test/coverage: run tests with coverage report
.PHONY: test/coverage
test/coverage:
	@echo 'Running tests with coverage...'
	go test -v -race -buildvcs -coverprofile=coverage.out ./cmd/api/...
	go tool cover -html=coverage.out -o coverage.html
	@echo 'Coverage report generated: coverage.html'

## build: build the cmd/api application
.PHONY: build
build:
	@echo 'Building cmd/api...'
	go build -ldflags='-s' -o=./bin/api ./cmd/api

## clean: remove build artifacts
.PHONY: clean
clean:
	@echo 'Cleaning build artifacts...'
	rm -rf ./bin
	rm -f coverage.out coverage.html

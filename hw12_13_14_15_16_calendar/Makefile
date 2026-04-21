BIN_CALENDAR := "./bin/calendar"
BIN_SCHEDULER := "./bin/calendar_scheduler"
BIN_STORER := "./bin/calendar_storer"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN_CALENDAR) -ldflags "$(LDFLAGS)" ./cmd/calendar
	go build -v -o $(BIN_SCHEDULER) -ldflags "$(LDFLAGS)" ./cmd/scheduler
	go build -v -o $(BIN_STORER) -ldflags "$(LDFLAGS)" ./cmd/storer

run: build
	$(BIN_CALENDAR) -config ./configs/config.toml

run-scheduler: build
	$(BIN_SCHEDULER) -config ./configs/scheduler_config.toml

run-storer: build
	$(BIN_STORER) -config ./configs/storer_config.toml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN_CALENDAR) version

test:
	go test -v -count=1 -race -timeout=1m ./internal/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.54.2

lint: install-lint-deps
	golangci-lint run ./...

# Генерация API кода из OpenAPI спецификации
gen-api: install-gen-deps
	@echo "Generating API code from OpenAPI specification..."
	@which oapi-codegen > /dev/null || (echo "oapi-codegen not installed. Run: go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.16.3" && exit 1)
	cd api && oapi-codegen -package api -generate types,models openapi.yaml > generated.go
	@echo "API code generated successfully"

install-gen-deps:
	go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.16.3

migrate:
	@echo "Applying migrations..."
	@which goose > /dev/null || (echo "goose not installed. Run: go install github.com/pressly/goose/v3/cmd/goose@latest" && exit 1)
	cd migrations && goose postgres "postgres://calendar:calendar@localhost:5432/calendar?sslmode=disable" up

# Docker команды
up:
	@echo "Starting all services with docker-compose..."
	docker-compose -f deployments/docker-compose.yml up -d
	@echo "Services started. Calendar API available at http://localhost:8080"

down:
	@echo "Stopping all services..."
	docker-compose -f deployments/docker-compose.yml down
	@echo "Services stopped"

logs:
	docker-compose -f deployments/docker-compose.yml logs -f

# Интеграционные тесты
integration-tests:
	@echo "Running integration tests..."
	@docker-compose -f deployments/docker-compose.test.yml up --no-build -d
	@echo "Waiting for integration tests to complete..."
	@docker wait calendar-integration-tests || EXIT_CODE=$$?; \
	echo "Collecting test logs..."; \
	docker logs calendar-integration-tests; \
	echo "Cleaning up test environment..."; \
	docker-compose -f deployments/docker-compose.test.yml down -v; \
	if [ "$$EXIT_CODE" = "0" ] || [ -z "$$EXIT_CODE" ]; then \
		echo "Integration tests passed!"; \
		exit 0; \
	else \
		echo "Integration tests failed with exit code $$EXIT_CODE"; \
		exit $$EXIT_CODE; \
	fi

.PHONY: build run run-scheduler run-storer build-img run-img version test install-lint-deps lint migrate gen-api install-gen-deps up down logs integration-tests

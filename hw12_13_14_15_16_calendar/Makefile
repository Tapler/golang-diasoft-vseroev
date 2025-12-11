BIN := "./bin/calendar"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

run: build
	$(BIN) -config ./configs/config.toml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version

test:
	go test -v -count=1 -race -timeout=1m ./internal/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.50.1

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

.PHONY: build run build-img run-img version test install-lint-deps lint migrate gen-api install-gen-deps

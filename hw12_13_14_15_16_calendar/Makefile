# ───── настройки ─────
BIN       ?= ./bin
SERVICES  := calendar scheduler sender   # ← новый сервис? просто допиши сюда
TAG       ?= develop

GIT_HASH   := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%S)
LDFLAGS    := -X main.release=$(TAG) \
              -X main.buildDate=$(BUILD_DATE) \
              -X main.gitHash=$(GIT_HASH)

# ───── Go ─────
generate:          ## go generate ./...
	go generate ./...

build-loc:         ## собрать все бинарники
	@for s in $(SERVICES); do \
	    echo "→ build $$s"; \
	    go build -v -o $(BIN)/$$s -ldflags '$(LDFLAGS)' ./cmd/$$s ; \
	done

run-loc: build-loc ## запустить на хосте
	@for s in $(SERVICES); do \
	    $(BIN)/$$s -config ./configs/$$s\_config.toml & \
	done ; wait

version: build-loc ## показать версии
	@for s in $(SERVICES); do $(BIN)/$$s version ; done

# ───── Docker ─────
build-img:         ## собрать образы сервисов + БД + Rabbit
	@for s in $(SERVICES); do \
	    docker build -t $$s:$(TAG) \
	      --build-arg LDFLAGS='$(LDFLAGS)' \
	      --build-arg SERVICE_NAME=$$s \
	      -f build/Dockerfile . ; \
	done

up: build-img     ## поднять docker-compose стек
	docker compose -f deployments/docker-compose.yml up -d

down:
	docker compose -f deployments/docker-compose.yml down

# Запуск PostgreSQL в Docker для тестов
test-db-run:
	docker run -d --rm --name calendar-test-db \
		-e POSTGRES_USER=otus_user \
		-e POSTGRES_PASSWORD=otus_password \
		-e POSTGRES_DB=otus \
		-p 5432:5432 \
		postgres:latest

# Ожидание пока БД будет готова
wait-db:
	@echo "Waiting for PostgreSQL to be ready..."
	@until pg_isready -h localhost -p 5432 -U otus_user > /dev/null 2>&1; do \
		sleep 1; \
	done

# Остановка контейнера с БД
test-db-stop:
	docker stop calendar-test-db || true

test: test-db-run wait-db
	TEST_DSN="host=localhost port=5432 user=otus_user password=otus_password dbname=otus sslmode=disable" go test -race ./internal/...
	$(MAKE) test-db-stop

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.64.8

lint: install-lint-deps
	go mod download
	golangci-lint run ./...

.PHONY: generate build-loc run-loc build-img up down version test lint test-db-run wait-db test-db-stop

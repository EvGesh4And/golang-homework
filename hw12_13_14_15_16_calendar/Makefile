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

test:
	go test -tags=integration -race ./internal/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.64.8

lint: install-lint-deps
	go mod download
	golangci-lint run ./...

# Запуск PostgreSQL в Docker для тестов
test-dc-run:
	docker compose -f ./deployments/docker-compose.yaml up -d
	sleep 2

# Остановка контейнера с БД
test-dc-stop:
	docker compose -f ./deployments/docker-compose.yaml down -v || true

integration-tests: test-dc-run
	CGO_ENABLED=1 go test -tags=integration -race ./test/...
	$(MAKE) test-dc-stop

.PHONY: generate build-loc run-loc build-img up down version test lint integration-tests

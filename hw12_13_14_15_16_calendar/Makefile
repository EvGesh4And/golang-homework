BIN := "./bin/calendar"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

generate:
	go generate ./...

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

# Запуск PostgreSQL в Docker для тестов
test-db-run:
	docker run -d --rm --name calendar-test-db \
		-e POSTGRES_USER=otus_user \
		-e POSTGRES_PASSWORD=otus_password \
		-e POSTGRES_DB=otus \
		-p 54321:5432 \
		postgres:15

# Ожидание пока БД будет готова (не самый быстрый способ, но простой)
wait-db:
	@echo "Waiting for PostgreSQL to be ready..."
	@until pg_isready -h localhost -p 54321 -U otus_user > /dev/null 2>&1; do \
		sleep 1; \
	done

# Остановка контейнера с БД (если нужно)
test-db-stop:
	docker stop calendar-test-db || true

test: test-db-run wait-db
	TEST_DSN="host=localhost port=54321 user=otus_user password=otus_password dbname=otus sslmode=disable" go test -race ./internal/...
	$(MAKE) test-db-stop

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.64.8

lint: install-lint-deps
	go mod download
	golangci-lint run ./...

.PHONY: generate build run build-img run-img version test lint test-db-run wait-db test-db-stop

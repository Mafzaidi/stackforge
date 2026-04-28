ifneq (,$(wildcard .env))
	include .env
	export
endif

CONTAINER_NAME=stackforge
APP_NAME ?= stackforge
IMAGE_NAME ?= $(REGISTRY)/$(APP_NAME)
VERSION := $(shell git describe --tags --abbrev=0 || echo "0.0.0")
COMMIT := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

MIGRATE := $(HOME)/go/bin/migrate
DB_URL = postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DATABASE)?sslmode=disable
MIGRATION_PATH := migrations

migrate-create:
	$(MIGRATE) create -ext sql -dir $(MIGRATION_PATH) -seq $(MIGRATION_NAME)

migrate-up:
	$(MIGRATE) -path $(MIGRATION_PATH) -database "$(DB_URL)" up

migrate-down:
	$(MIGRATE) -path $(MIGRATION_PATH) -database "$(DB_URL)" down

migrate-force:
	$(MIGRATE) -path $(MIGRATION_PATH) -database "$(DB_URL)" force $(ver)

build:
	go mod tidy
	go build -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)" -o ./bin/$(APP_NAME) ./cmd/api

run: build
	./bin/$(APP_NAME)

clean:
	rm -f ./bin/$(APP_NAME)
	
docker-build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		-t $(IMAGE_NAME):$(VERSION) .

docker-run: docker-build
	docker run --rm --name $(CONTAINER_NAME) $(IMAGE_NAME)

docker-push: docker-build
	docker tag $(IMAGE_NAME):$(VERSION) $(IMAGE_NAME):latest
	docker push $(IMAGE_NAME):$(VERSION)
	docker push $(IMAGE_NAME):latest

.PHONY: build test run docker-build docker-run docker-push
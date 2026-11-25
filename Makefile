.PHONY: fmt lint test run build image

IMAGE ?= cicd-demo
VERSION ?= dev
COMMIT ?= unknown

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

test:
	go test -race ./...

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

image:
	docker build --build-arg VERSION="$(VERSION)" --build-arg COMMIT="$(COMMIT)" -t "$(IMAGE)" .

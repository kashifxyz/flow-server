.PHONY: build run test lint fmt migrate generate

build:
	go build -o bin/flow-server ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./...

lint:
	go vet ./...

fmt:
	gofmt -w .

migrate:
	@echo "migrations are not configured yet"

generate:
	@echo "code generation is not configured yet"

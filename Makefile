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
	@echo "Apply migrations locally with:"
	@echo "  goose -dir migrations postgres \"host=localhost port=5432 dbname=flow sslmode=disable\" up"

generate:
	@echo "code generation is not configured yet"

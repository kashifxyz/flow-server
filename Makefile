.PHONY: build run test lint fmt migrate migrate-revert generate

# Source FLOW_DATABASE_URL from .env in the same shell as goose.
define goose-with-env
	if [ ! -f .env ]; then echo "missing .env; copy .env.example to .env" >&2; exit 1; fi; \
	set -a && . ./.env && set +a && \
	if [ -z "$$FLOW_DATABASE_URL" ]; then echo "FLOW_DATABASE_URL is not set in .env" >&2; exit 1; fi && \
	goose -dir migrations postgres "$$FLOW_DATABASE_URL" $(1)
endef

build:
	go build -o bin/flow-server ./cmd/server

run:
	go run ./cmd/server -mode=dev -url=http://localhost:5173

test:
	go test ./...

lint:
	go vet ./...

fmt:
	gofmt -w .

migrate:
	@$(call goose-with-env,up)

migrate-revert:
	@$(call goose-with-env,down)

generate:
	@echo "code generation is not configured yet"

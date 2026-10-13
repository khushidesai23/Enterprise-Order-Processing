.PHONY: run build test test-db swagger fmt tidy docker-up docker-down logs clean

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./...

# Integration and e2e tests against the postgres-test container (port 5433).
test-db:
	ENABLE_DB_TESTS=1 go test -p 1 -tags=integration ./internal/repository ./internal/modules/order
	ENABLE_DB_TESTS=1 go test -tags=e2e ./internal/test/e2e/...

swagger:
	swag init -g ./cmd/server/main.go --parseInternal --parseDependency

fmt:
	go fmt ./...

tidy:
	go mod tidy

docker-up:
	docker compose up -d

docker-down:
	docker compose down

logs:
	docker compose logs -f

clean:
	rm -rf bin

.PHONY: run build test test-db swagger fmt tidy docker-up docker-down logs clean promote-admin

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

# Grant the admin role to an already registered user. Registration never
# grants it, so the first admin is created this way; later ones can be
# promoted through PATCH /api/v1/users/{id}/role. The user must log in again
# to get a token carrying the new role.
promote-admin:
	@test -n "$(EMAIL)" || (echo "usage: make promote-admin EMAIL=user@example.com" && exit 1)
	echo "UPDATE users SET role = 'admin' WHERE email = lower(:'email');" | \
		docker compose exec -T postgres sh -c 'psql -v ON_ERROR_STOP=1 -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" -v email="$$0"' "$(EMAIL)"

clean:
	rm -rf bin

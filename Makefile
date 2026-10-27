.PHONY: run build test test-db swagger fmt tidy docker-up docker-down logs clean promote-admin \n	frontend-install frontend-dev frontend-build frontend-test frontend-e2e seed

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

# Frontend (frontend/): storefront at /shop, admin console at /admin.
# The dev server proxies /api to the API on :8080 (API_PROXY_TARGET).
frontend-install:
	cd frontend && npm ci

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

frontend-test:
	cd frontend && npm run typecheck && npm test

# Browser end-to-end tests against the running API (make run) and a
# promoted admin (ADMIN_EMAIL / ADMIN_PASSWORD, default admin@meridian.test).
frontend-e2e:
	cd frontend && npx playwright test

# Large demo store: ~840 products with images in 24 departments, 1,500
# customers, 25,000 orders with payments over a year, ~11,000 reviews.
# Safe to re-run. Scale with ARGS, e.g. make seed ARGS="-orders 100000".
seed:
	go run ./cmd/seed $(ARGS)

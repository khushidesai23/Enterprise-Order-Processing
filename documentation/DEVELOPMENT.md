# Development Guide

## Prerequisites

- Go (version in `go.mod`), Docker with Compose, Git
- Node 22 for the frontend
- Optional: Python 3.13 for Locust, ngrok for Razorpay webhooks

## Local Startup

1. Copy `.env.example` to `.env`. Set `JWT_SECRET` and the Razorpay test
   keys; the API refuses to start without them.
2. Start the infrastructure:

   ```bash
   docker compose up -d
   ```

   This starts PostgreSQL, the test database, Kafka, Debezium, the OTel
   Collector, Jaeger, Prometheus and Grafana.

3. Run the API (it migrates the schema on start):

   ```bash
   go run ./cmd/server
   ```

4. One-time CDC setup, after the first start:

   ```powershell
   Get-Content config\debezium\setup.sql | docker exec -i order-postgres psql -U postgres -d order_processing
   curl.exe -X POST -H "Content-Type: application/json" --data "@config/debezium/postgres-connector.json" http://localhost:8083/connectors
   ```

5. Run the frontend (storefront at http://localhost:5173/shop, admin console
   at `/admin`):

   ```bash
   cd frontend && npm install && npm run dev
   ```

6. Optionally load the demo store and promote an admin:

   ```bash
   make seed
   make promote-admin EMAIL=you@example.com
   ```

On Windows, if a native PostgreSQL service already listens on 5432, set
`DB_PORT=5434` (see the README configuration table).

## Common Commands

| Command | Does |
|---|---|
| `make run` / `make build` | Run or build the API (`bin/server`) |
| `make test` | Go unit tests |
| `make test-db` | Integration and e2e tests against `postgres-test` |
| `make fmt` / `make tidy` | `go fmt`, `go mod tidy` |
| `make swagger` | Regenerate `docs/` from annotations |
| `make docker-up` / `make docker-down` / `make logs` | Compose stack |
| `make promote-admin EMAIL=...` | Grant the admin role to a registered user |
| `make seed ARGS=...` | Demo-store seeder (`cmd/seed`) |
| `make frontend-install` / `frontend-dev` / `frontend-build` | Frontend |
| `make frontend-test` / `frontend-e2e` | Typecheck + Vitest / Playwright |
| `make clean` | Remove `bin/` |

## API Documentation

Swagger UI: http://localhost:8080/swagger/index.html

Regenerate Swagger after changing API annotations:

```bash
swag init -g ./cmd/server/main.go --parseInternal --parseDependency
```

## Payments Locally

The frontend completes payments through `POST /payments/verify`, so no
public webhook URL is needed.

To test webhooks:

1. Expose the API with `ngrok http 8080`.
2. Set the Razorpay webhook URL to
   `https://<ngrok-domain>/api/v1/payments/webhook`.
3. Subscribe to `payment.captured`, `payment.failed` and `refund.created`.

`payment-demo/` is a standalone test-mode checkout page.

## Development Expectations

- Run `gofmt` before committing; CI fails on unformatted code.
- Keep handlers focused on HTTP, business rules in services, database
  access in repositories.
- Use one transaction for multi-step workflows, and lock rows in a
  consistent order (order row first, products by ID).
- Enforce roles in routes (`RequireRole`) and ownership in services
  (`internal/authz`). Return `404` for another user's resources.
- Add tests with new behaviour; DB-backed tests need `ENABLE_DB_TESTS=1`.
- Update Swagger when public API contracts change.
- Never commit `.env` or real secrets.

## Further Reading

- [ARCHITECTURE.md](ARCHITECTURE.md), [OBSERVABILITY.md](OBSERVABILITY.md),
  [TESTING.md](TESTING.md), [ROADMAP.md](ROADMAP.md)
- [DEPLOYMENT.md](../DEPLOYMENT.md), [LOAD_TESTING.md](../LOAD_TESTING.md),
  [frontend/README.md](../frontend/README.md)
- [PROJECT_REPORT.pdf](PROJECT_REPORT.pdf),
  [FAILURE_RECOVERY_DRILLS.pdf](FAILURE_RECOVERY_DRILLS.pdf)

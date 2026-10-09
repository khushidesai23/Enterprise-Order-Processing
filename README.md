# Enterprise Order Processing Platform

A Go modular monolith for transactional order processing: inventory
reservation, Razorpay payments with idempotent webhooks, JWT authentication,
PostgreSQL change data capture (Debezium + Kafka), and a full local
observability stack (Prometheus, Grafana, OpenTelemetry, Jaeger) with a
reproducible Locust load-test suite.

Documentation:

- **This README**: architecture, setup, API, observability, testing, roadmap.
- **[LOAD_TESTING.md](LOAD_TESTING.md)**: load-test scenarios, what to watch,
  measured results and troubleshooting.
- API contract: Swagger UI at http://localhost:8080/swagger/index.html
  (`docs/swagger.yaml`).

## Status

| Area | State |
|---|---|
| Auth, users, categories, products, inventory, orders, payments | Implemented and tested |
| Transactional order creation with row-locked, deadlock-free inventory reservation | Implemented; covered by concurrency integration tests |
| Razorpay checkout + webhook signature verification and idempotency | Implemented and tested |
| Prometheus metrics, provisioned Grafana dashboard, OTel tracing to Jaeger | Implemented and verified under load |
| CDC: PostgreSQL -> Debezium -> Kafka -> API consumer | Running and measured; the consumer only counts and logs events |
| Locust suite with seeded data and load shapes | Implemented; results in LOAD_TESTING.md |
| Authorization (roles, ownership) | **Not implemented**, see [Known limitations](#known-limitations) |

## Architecture

```text
Client -> Gin router -> middleware -> handler -> service -> repository -> PostgreSQL
            |                                                             |
            |  middleware: recovery, OTel, Prometheus, 10s request         | WAL (logical)
            |  deadline, access log, CORS, JWT (per route)                 v
            |                                                   Debezium -> Kafka
            +-- /metrics -> Prometheus -> Grafana                          |
            +-- OTLP -> OTel Collector -> Jaeger          API CDC consumer <-+
```

| Layer | Responsibility |
|---|---|
| `internal/app`, `internal/api/routes` | Dependency wiring, middleware, route registration (each module has `RegisterRoutes`) |
| `internal/api/middleware` | JWT auth, CORS, Prometheus, request deadline, access log |
| `internal/modules/<module>` | Handler (HTTP), service (business rules), DTOs, mappers, errors |
| `internal/repository` | All database access (GORM) |
| `internal/messaging/kafka` | Debezium CDC consumer |
| `pkg/metrics`, `pkg/telemetry`, `pkg/logger` | Prometheus collectors, OTel tracer, Zap logger |

Modules talk to each other through service interfaces, not HTTP, so any
module can later be extracted without changing callers.

### Consistency rules

- **Order creation** runs in one transaction: insert order (with total),
  reserve stock per item, insert items, commit. Product prices and the user
  are read before the transaction. Each reservation is a single guarded
  `UPDATE inventories ... WHERE available_quantity >= ?`, so stock can never
  go negative, and items are processed in product-ID order so concurrent
  multi-item orders cannot deadlock.
- **Status changes** (cancel, payment, shipping, refund) lock the order row
  (`SELECT ... FOR UPDATE OF orders`), so concurrent transitions are
  serialized and reserved stock is released or confirmed exactly once.
- **Webhooks** are deduplicated on the Razorpay event ID (unique
  `payment_webhooks.payload_id`), and the payment row is locked while its
  status is applied. A failed application rolls back the webhook record so
  the gateway's retry is processed again.
- **Request deadline**: every request context has a 10 s deadline (below the
  15 s server write timeout). Stalled queries are cancelled and the client
  receives `504` instead of a silently dropped connection.

### Order lifecycle

```text
CREATED -> PAYMENT_PENDING -> PAID -> PACKED -> SHIPPED -> DELIVERED
   |             |             |
   +-------------+-------------+--> CANCELLED (releases reserved stock)
```

Shipping confirms (consumes) the reserved stock. `payment.failed` cancels the
order; refunds cancel non-delivered orders.

## Technology

Go 1.26, Gin, GORM, PostgreSQL 16, JWT, Razorpay, Viper, Zap, Swaggo,
OpenTelemetry, OTel Collector 0.158.0, Jaeger 1.76.0, Prometheus v3.11.3,
Grafana 13.0.1, Kafka 3.2 (KRaft) and Debezium 3.2, Locust 2.46, Docker
Compose, GitHub Actions.

## Project structure

```text
cmd/server/            entry point (config, telemetry, DB, CDC consumer, HTTP server, graceful shutdown)
config/                config.go, prometheus.yml, collector-config.yaml,
                       debezium/ (connector + setup.sql), grafana/ (provisioning + dashboard)
docs/                  generated Swagger
internal/api/          handlers (health), middleware, response, routes
internal/app/          router and dependency wiring
internal/database/     connection pool, migrations, health
internal/messaging/    Kafka CDC consumer
internal/models/       GORM models
internal/modules/      auth, user, category, product, inventory, order, payment
internal/repository/   data access
internal/test/         e2e/, integration/ (DB fixtures and harness), mocks/
locust/                seed.py, scenarios/, requirements.txt
payment-demo/          local Razorpay checkout page
pkg/                   logger, metrics, telemetry, utils, validator
```

## Getting started

Prerequisites: Go (version in `go.mod`), Docker with Compose, Git. Optional:
Python 3.13 for Locust, ngrok for webhook testing.

```powershell
Copy-Item .env.example .env      # then set JWT_SECRET and Razorpay test keys
docker compose up -d             # PostgreSQL, test DB, Kafka, Debezium, OTel, Jaeger, Prometheus, Grafana
go run ./cmd/server              # or: go build -o bin/server.exe ./cmd/server; .\bin\server.exe
curl.exe http://localhost:8080/api/v1/ready
```

One-time CDC setup (after the API has started once and migrated the schema):

```powershell
Get-Content config\debezium\setup.sql | docker exec -i order-postgres psql -U postgres -d order_processing
curl.exe -X POST -H "Content-Type: application/json" --data "@config/debezium/postgres-connector.json" http://localhost:8083/connectors
```

On Linux/macOS the same steps work with `cp`, `cat ... |` and `curl`; a
`Makefile` wraps the common commands (`make run`, `make test`, `make test-db`,
`make swagger`, ...).

### Configuration

All settings come from environment variables, optionally loaded from `.env`
(see `.env.example`). Notable ones:

| Variable | Default | Notes |
|---|---|---|
| `DB_PORT` | 5432 | **Windows:** if a native PostgreSQL service listens on 5432, `localhost:5432` reaches it instead of the container. Use `5434` and `docker compose up -d postgres`. Check with `Get-NetTCPConnection -LocalPort 5432 -State Listen`. |
| `DB_MAX_OPEN_CONNS` / `DB_MAX_IDLE_CONNS` / `DB_CONN_MAX_IDLE_TIME` | 80 / 80 / 5m | Keep open connections below PostgreSQL `max_connections` (100) minus Debezium/admin connections. |
| `KAFKA_CDC_TOPIC` | categories, orders, inventories topics | Comma-separated `order-processing.public.<table>` topics. |
| `JWT_SECRET`, `RAZORPAY_*` | none | Required; the app refuses to start without them. |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `localhost:4317` | OTel Collector. |

Never commit `.env`.

### Local URLs

| Interface | URL |
|---|---|
| API | http://localhost:8080/api/v1 |
| Swagger UI | http://localhost:8080/swagger/index.html |
| API metrics | http://localhost:8080/metrics |
| Grafana dashboard | http://localhost:3000/d/order-processing-load-testing (admin / admin, local only) |
| Prometheus | http://localhost:9090 (targets: /targets) |
| Jaeger | http://localhost:16686 (service `enterprise-order-processing`) |
| Kafka Connect REST | http://localhost:8083/connectors |
| Locust UI / exporter | http://localhost:8089 / http://localhost:9646/metrics (while Locust runs) |

## API

All business routes are under `/api/v1` and, except login, user sign-up,
category/product reads, health and the payment webhook, require
`Authorization: Bearer <token>` from `POST /api/v1/auth/login`.

| Group | Routes |
|---|---|
| Health | `GET /health`, `/ready` (checks DB), `/ping`, `/version` |
| Auth | `POST /auth/login`, `GET /auth/me` |
| Users | `POST /users` (sign-up), `GET/PUT/DELETE /users/:id`, `GET /users` |
| Catalog | `/categories`, `/products` (`GET /products/category/:categoryId`) |
| Inventory | `/inventory`, `/inventory/:productId` + `add-stock`, `remove-stock`, `reserve`, `release`, `confirm` |
| Orders | `POST /orders`, `GET /orders?page=&limit=` (limit <= 100), `GET /orders/me`, `GET /orders/:id`, `PATCH /orders/:id/status`, `PATCH /orders/:id/cancel` |
| Payments | `POST /payments`, `GET /payments`, `/payments/summary`, `/payments/:id`, `/payments/order/:orderId`, `POST /payments/:id/refund`, `POST /payments/webhook` |

Regenerate Swagger after changing annotations:
`swag init -g ./cmd/server/main.go --parseInternal --parseDependency`.

### Payments and webhooks

`payment-demo/` is a local checkout page for Razorpay test mode: create an
order, `POST /payments`, open the page and pay. Razorpay cannot reach
`localhost`, so expose the API (`ngrok http 8080`) and configure the webhook
URL `https://<ngrok-domain>/api/v1/payments/webhook`. Webhooks must carry
`X-Razorpay-Signature` (verified) and `X-Razorpay-Event-Id` (dedupe key;
missing -> 400).

## Observability

| Signal | Path |
|---|---|
| Metrics | API `/metrics` and Locust `:9646/metrics`, scraped every 5 s -> Prometheus -> Grafana |
| Traces | otelgin HTTP spans + otelgorm SQL spans (no bind values) -> OTLP -> Collector -> Jaeger; also a Jaeger datasource in Grafana |
| Logs | Zap JSON on stdout with `trace_id`/`span_id` |

`/metrics` itself is excluded from metrics, traces and access logs. Grafana's
provisioned dashboard **Performance / Order Processing - Load Testing** covers
Locust vs server traffic, latency percentiles (overall and per route), errors,
Go runtime, DB pool, CDC and a slow-trace table; LOAD_TESTING.md explains
which panels to watch for each test.

| Metric | Labels | Meaning |
|---|---|---|
| `order_processing_http_requests_total` | method, route, status | Requests handled (route = Gin template) |
| `order_processing_http_request_duration_seconds` | method, route, status | Handler latency histogram |
| `order_processing_http_requests_in_flight` | method, route | Requests being handled |
| `go_sql_{open,in_use,idle,max_open}_connections` | db_name | DB pool state at scrape time |
| `go_sql_wait_count_total`, `go_sql_wait_duration_seconds_total` | db_name | Waits for a pool connection |
| `go_sql_max_{idle,idle_time,lifetime}_closed_total` | db_name | Connections closed by pool limits |
| `order_processing_cdc_events_total` | table, operation | Debezium events consumed (c/u/d/r) |
| `order_processing_cdc_errors_total` | stage | Consumer read/decode/process errors |
| `order_processing_cdc_end_to_end_lag_seconds` | | PostgreSQL commit to consumption |
| `order_processing_cdc_consumer_lag_messages` | | Messages behind the latest fetched partition head |
| `locust_users`, `locust_requests_total`, `locust_request_duration_seconds` | name, method, result | Load generator (while running) |
| `go_*`, `process_*` | | Runtime/process; filter by `job="order-processing-api"` (Prometheus exports the same names) |

Not exported: Debezium/Kafka Connect JMX metrics and PostgreSQL server
metrics. Check the connector with
`curl.exe http://localhost:8083/connectors/order-processing-postgres-connector/status`
and consumer lag with
`docker exec order-kafka /kafka/bin/kafka-consumer-groups.sh --bootstrap-server kafka:9092 --describe --group order-processing-cdc-consumer`.

## Testing

| Suite | Command | Needs |
|---|---|---|
| Unit | `go test ./...` | nothing |
| Repository + order service integration | `go test -p 1 -tags=integration ./internal/repository ./internal/modules/order` | `ENABLE_DB_TESTS=1`, `postgres-test` container (5433) |
| HTTP end-to-end | `go test -tags=e2e ./internal/test/e2e/...` | same |

DB-backed tests are skipped unless `ENABLE_DB_TESTS=1` (PowerShell:
`$env:ENABLE_DB_TESTS = "1"`). They truncate the shared test database, so
run packages sequentially (`-p 1`). The integration suite includes
concurrency regressions: 50 concurrent orders against 10 units, concurrent
cancels of one order, opposite-order multi-item orders (deadlock) and
concurrent stock removals.

CI (`.github/workflows/ci.yml`) runs gofmt, build, unit, integration and e2e
tests against a PostgreSQL service container.

Load tests: see [LOAD_TESTING.md](LOAD_TESTING.md). Quick start:

```powershell
cd locust
python -m venv .venv-locust; .\.venv-locust\Scripts\pip install -r requirements.txt
.\.venv-locust\Scripts\python seed.py --host http://localhost:8080
.\.venv-locust\Scripts\locust -f scenarios/smoke.py --host http://localhost:8080 -u 2 -r 2 -t 1m --headless
```

## Known limitations

- **No authorization model.** Any authenticated user can list all orders,
  read or cancel another user's order, set order status (including `PAID`
  without a payment), adjust inventory, and update or delete users. Roles and
  ownership checks are needed before use beyond local development.
- A failed payment (`payment.failed`) cancels the order, so a later
  successful retry on the same Razorpay order cannot mark it paid.
- `BaseModel.BeforeCreate` always assigns a new UUID, so creating a record
  with a populated association re-IDs that association.
- The CDC consumer has no business logic yet.

## Roadmap

1. Authorization: roles and resource ownership checks.
2. CDC consumers with business logic (notifications, analytics, audit),
   consumer idempotency, retry and dead-letter handling.
3. Alerting and SLOs; PostgreSQL and Debezium JMX exporters.
4. TimescaleDB for operational time-series, retention and archival.

## Author

**Khushi Desai**, Associate Software Engineer | Cloud & Backend Development

GitHub: https://github.com/khushidesai23 ·
LinkedIn: https://www.linkedin.com/in/khushi-desai-ab5154225/

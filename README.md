# Enterprise Order Processing Platform

A backend-focused, enterprise-style Order Processing Platform built to demonstrate production-oriented backend engineering practices.

The project is implemented as a **modular monolith** in Go. It currently provides transactional order processing, inventory reservation, Razorpay payment integration, authentication, testing, OpenTelemetry tracing, Prometheus metrics, Grafana visualization, structured logging, Docker-based local infrastructure, and GitHub Actions CI.

> The project is intentionally evolving in phases. A PostgreSQL -> Debezium -> Kafka CDC pipeline runs locally and is consumed and measured by the API; business consumers (notifications, analytics, audit) and TimescaleDB are planned for later phases.

---

## Current Implementation Status

### Implemented

- Go backend using Gin
- PostgreSQL with GORM
- Modular monolith architecture
- Repository and service layers
- JWT authentication and protected APIs
- User management
- Category management
- Product management
- Inventory management
- Transaction-safe order creation
- Inventory reservation lifecycle
- Razorpay payment integration
- Checkout signature verification
- Razorpay webhook signature verification
- Idempotent webhook handling
- Payment and order state updates
- Swagger/OpenAPI documentation
- Structured logging with Zap
- OpenTelemetry distributed tracing
- Jaeger trace visualization
- Prometheus metrics
- Grafana provisioning with Prometheus datasource
- Health and readiness endpoints
- Graceful shutdown
- Docker Compose local infrastructure
- Unit, repository integration, service integration, and end-to-end tests
- GitHub Actions CI
- Locust load-test suite with seeded fixtures, load shapes and a Prometheus exporter
- Provisioned Grafana load-testing dashboard (HTTP, latency, errors, runtime, DB pool, CDC, Jaeger slow traces)
- PostgreSQL CDC with Debezium and Kafka (connector config, setup SQL, instrumented consumer)

### Partially implemented

- CDC consumer: events are decoded, counted and logged; no business processing yet
- Authorization: every business route requires a JWT, but there are no roles or ownership checks (see Known limitations)

### Planned

- Event consumers with business logic
- Notification service
- Analytics service
- Audit/event service
- TimescaleDB
- Business dashboards
- Alerting
- Data retention and archival

See [documentation/ROADMAP.md](documentation/ROADMAP.md) for the implementation roadmap.

---

# Architecture

```text
                                Clients
                                   |
                                   v
                           Gin HTTP Router
                                   |
                    +--------------+--------------+
                    |                             |
                    v                             v
             Public Endpoints              JWT Middleware
                                                  |
                                                  v
                                              Handlers
                                                  |
                                                  v
                                               Services
                                                  |
                                                  v
                                             Repositories
                                                  |
                                                  v
                                             PostgreSQL
                                                  |
                    +-----------------------------+-----------------------------+
                    |                             |                             |
                    v                             v                             v
              Prometheus Metrics            OpenTelemetry                 Zap Logging
                    |                             |
                    v                             v
               Prometheus                  OTEL Collector
                    |                             |
                    v                             v
                Grafana                        Jaeger
```

The application is currently a modular monolith. Business modules are separated internally so that selected domains can later be extracted into independent services if required.

For more detail, see [documentation/ARCHITECTURE.md](documentation/ARCHITECTURE.md).

---

# Technology Stack

| Area | Technology |
|---|---|
| Language | Go |
| HTTP Framework | Gin |
| ORM | GORM |
| Database | PostgreSQL 16 |
| Authentication | JWT |
| Payment Gateway | Razorpay |
| Configuration | Viper |
| Logging | Zap |
| API Documentation | Swagger / Swaggo |
| Tracing | OpenTelemetry |
| Trace Backend | Jaeger |
| Metrics | Prometheus |
| Visualization | Grafana |
| Load Testing | Locust |
| Containers | Docker Compose |
| CI | GitHub Actions |

---

# Core Business Modules

## Authentication

Provides:

- Login
- JWT generation
- Authenticated user lookup
- Protected API access

## Users

Provides user creation, retrieval, update, and management through the application module structure.

## Categories and Products

Provides catalog organization through category and product modules.

## Inventory

Inventory tracks stock quantities and supports reservation as part of the order workflow.

The current business flow is:

```text
Available Stock
      |
      v
Order Created
      |
      v
Inventory Reserved
      |
      +----------------------------+
      |                            |
      v                            v
Payment / Order Continues       Order Cancelled
      |                            |
      v                            v
Reservation Confirmed          Reserved Stock Released
```

## Orders

Order creation is handled transactionally together with the required inventory reservation work so the core business state remains consistent.

## Payments

Razorpay is integrated for payment processing.

The payment flow includes:

```text
Create Order
      |
      v
Reserve Inventory
      |
      v
Create Payment / Razorpay Order
      |
      v
Customer Completes Checkout
      |
      v
Verify Checkout Signature
      |
      v
Razorpay Webhook Received
      |
      v
Verify Webhook Signature
      |
      v
Persist Webhook Idempotently
      |
      v
Update Payment State
      |
      v
Update Order State
```

---

# Order Lifecycle

The project models a commerce-style lifecycle:

```text
Created
   |
   v
Payment Pending
   |
   +--> Payment Failed
   |
   v
Paid
   |
   v
Packed
   |
   v
Shipped
   |
   v
Delivered
```

Cancellation and refund-related behavior are handled according to the current business rules and payment/inventory lifecycle.

---

# Project Structure

```text
.
├── .github/
│   └── workflows/
│       └── ci.yml
│
├── cmd/
│   └── server/
│       └── main.go
│
├── config/
│   ├── config.go
│   ├── collector-config.yaml
│   ├── prometheus.yml
│   ├── debezium/
│   │   ├── postgres-connector.json
│   │   └── setup.sql
│   └── grafana/
│       ├── dashboards/
│       └── provisioning/
│
├── docs/
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
│
├── documentation/
│   ├── ARCHITECTURE.md
│   ├── OBSERVABILITY.md
│   ├── ROADMAP.md
│   └── TESTING.md
│
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   ├── middleware/
│   │   ├── response/
│   │   └── routes/
│   │
│   ├── app/
│   │   └── router.go
│   │
│   ├── database/
│   │   ├── health.go
│   │   ├── migrate.go
│   │   └── postgres.go
│   │
│   ├── messaging/
│   │   └── kafka/
│   │
│   ├── models/
│   │
│   ├── modules/
│   │   ├── auth/
│   │   ├── user/
│   │   ├── category/
│   │   ├── product/
│   │   ├── inventory/
│   │   ├── order/
│   │   └── payment/
│   │
│   ├── repository/
│   │
│   └── test/
│       ├── e2e/
│       ├── integration/
│       ├── helpers/
│       └── mocks/
│
├── locust/
│   ├── seed.py
│   ├── requirements.txt
│   ├── locustfile.py
│   └── scenarios/
│
├── payment-demo/
│   ├── index.html
│   └── app.js
│
├── pkg/
│   ├── logger/
│   ├── metrics/
│   └── telemetry/
│
├── docker-compose.yml
├── LOAD_TESTING.md
├── Makefile
├── go.mod
└── README.md
```

---

# Local Prerequisites

Install:

- Go version defined by `go.mod`
- Docker and Docker Compose
- Git

Optional:

- ngrok for Razorpay webhook testing
- Python 3.13 for Locust (`locust/requirements.txt`)

---

# Configuration

Copy the example environment file:

```bash
cp .env.example .env
```

Configure the application values in `.env`.

The current application expects values for:

```env
APP_NAME=Enterprise Order Processing
APP_ENV=development
APP_PORT=8080

DB_HOST=localhost
DB_PORT=5432   # use e.g. 5434 if a native PostgreSQL already listens on 5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=order_processing
DB_SSLMODE=disable

KAFKA_BROKERS=localhost:9092
KAFKA_CDC_TOPIC=order-processing.public.categories,order-processing.public.orders,order-processing.public.inventories
KAFKA_CDC_GROUP_ID=order-processing-cdc-consumer

JWT_SECRET=replace-with-a-secure-secret
JWT_EXPIRATION=24h

LOG_LEVEL=debug

RAZORPAY_KEY_ID=your_razorpay_test_key
RAZORPAY_KEY_SECRET=your_razorpay_test_secret
RAZORPAY_WEBHOOK_SECRET=your_razorpay_webhook_secret

OTEL_SERVICE_NAME=enterprise-order-processing
OTEL_SERVICE_VERSION=1.0.0
OTEL_ENVIRONMENT=development
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317

TEST_DB_HOST=localhost
TEST_DB_PORT=5433
TEST_DB_USER=postgres
TEST_DB_PASSWORD=postgres
TEST_DB_NAME=order_processing_test
TEST_DB_SSLMODE=disable

TEST_RAZORPAY_KEY_ID=rzp_test_key
TEST_RAZORPAY_KEY_SECRET=rzp_test_secret
TEST_RAZORPAY_WEBHOOK_SECRET=rzp_test_webhook_secret
TEST_JWT_SECRET=test-jwt-secret
```

Do not commit real secrets.

> **Port 5432 conflict (Windows):** if a native PostgreSQL service also
> listens on 5432, `localhost:5432` reaches it instead of the Docker
> container, so the API writes to a database Debezium never sees. Set
> `DB_PORT=5434` in `.env`, then `docker compose up -d postgres`.
> Check with `Get-NetTCPConnection -LocalPort 5432 -State Listen`.

---

# Start Local Infrastructure

Start all Docker services:

```bash
docker compose up -d
```

Current Compose services:

- PostgreSQL 16 (`wal_level=logical`)
- PostgreSQL test database (port 5433)
- Kafka 3.2 (KRaft, host port 9092)
- Debezium / Kafka Connect 3.2 (REST on 8083)
- Jaeger 1.76.0
- OpenTelemetry Collector 0.158.0
- Prometheus v3.11.3
- Grafana 13.0.1

Check service status:

```bash
docker compose ps
```

View logs:

```bash
docker compose logs -f
```

Stop infrastructure:

```bash
docker compose down
```

---

# Run the Application

Install dependencies:

```bash
go mod tidy
```

Run the API:

```bash
go run ./cmd/server
```

Or use:

```bash
make run
```

The server starts on:

```text
http://localhost:8080
```

## Local URLs

| Interface | URL | Notes |
|---|---|---|
| API | http://localhost:8080/api/v1 | `GET /api/v1/ready` checks the DB |
| Swagger UI | http://localhost:8080/swagger/index.html | |
| API metrics | http://localhost:8080/metrics | |
| Grafana | http://localhost:3000/d/order-processing-load-testing | admin / admin (local only) |
| Prometheus | http://localhost:9090 | targets: http://localhost:9090/targets |
| Jaeger | http://localhost:16686 | service `enterprise-order-processing` |
| Kafka Connect (Debezium) | http://localhost:8083/connectors | REST API |
| Locust web UI | http://localhost:8089 | when started without `--headless` |
| Locust metrics | http://localhost:9646/metrics | while Locust runs |

---

# Available Endpoints

## Application

| Endpoint | Purpose |
|---|---|
| `GET /` | Root health response |
| `GET /api/v1/health` | Application health |
| `GET /api/v1/ready` | Readiness check |
| `GET /api/v1/ping` | Basic connectivity check |
| `GET /api/v1/version` | Application version information |
| `GET /metrics` | Prometheus metrics |
| `GET /swagger/index.html` | Swagger UI |

Business APIs are exposed under:

```text
/api/v1
```

Main modules:

```text
/auth
/users
/categories
/products
/inventory
/orders
/payments
```

For the complete request and response contract, use Swagger.

---

# Swagger API Documentation

Swagger UI:

```text
http://localhost:8080/swagger/index.html
```

Regenerate Swagger files after API annotation changes:

```bash
swag init -g ./cmd/server/main.go --parseInternal --parseDependency
```

For protected endpoints:

1. Login using the authentication API.
2. Copy the JWT token.
3. Click **Authorize** in Swagger.
4. Enter:

```text
Bearer <your-jwt-token>
```

---

# Payment Testing

The repository includes a simple local demo frontend:

```text
payment-demo/
```

Basic flow:

1. Start PostgreSQL and observability infrastructure.
2. Start the backend.
3. Register or use an existing user.
4. Login and obtain a JWT.
5. Create category/product/inventory data as required.
6. Create an order.
7. Create the payment flow.
8. Open the local payment demo.
9. Complete Razorpay checkout in Test Mode.
10. Verify the payment and webhook processing.

The demo frontend is intended for local development and learning only.

---

# Razorpay Webhook Testing

Razorpay cannot deliver webhooks directly to `localhost`.

Expose the local application:

```bash
ngrok http 8080
```

Use the generated public URL in the Razorpay dashboard with the payment webhook endpoint:

```text
https://<your-ngrok-domain>/api/v1/payments/webhook
```

The application verifies the `X-Razorpay-Signature` header before processing webhook data and deduplicates on `X-Razorpay-Event-Id` (requests without an event ID are rejected with 400).

---

# Observability

The current observability pipeline is:

```text
Go Application
    |
    +--> Prometheus Metrics --> Prometheus --> Grafana
    |
    +--> OpenTelemetry Traces --> OTEL Collector --> Jaeger
    |
    +--> Structured Logs --> stdout
```

## Prometheus

Prometheus:

```text
http://localhost:9090
```

The application exposes:

```text
http://localhost:8080/metrics
```

Prometheus is configured to scrape the API through:

```text
host.docker.internal:8080
```

This assumes the API is running on the host machine while Prometheus runs in Docker.

## Grafana

Grafana:

```text
http://localhost:3000
```

Current local credentials configured in `docker-compose.yml`:

```text
username: admin
password: admin
```

Change these credentials before any non-local deployment.

## Jaeger

Jaeger UI:

```text
http://localhost:16686
```

Grafana also has a provisioned Jaeger datasource, and the load-testing
dashboard lists recent slow traces.

See [documentation/OBSERVABILITY.md](documentation/OBSERVABILITY.md) for the metric catalogue.

---

# Testing

The project contains:

- Unit tests
- Repository integration tests
- Order service integration tests
- HTTP end-to-end tests
- Locust load-test scripts

Run standard tests:

```bash
go test ./...
```

Database-backed tests need the `postgres-test` container (port 5433) and
`ENABLE_DB_TESTS=1`; without it they are skipped. They share one test
database, so run packages sequentially (`-p 1`):

```bash
ENABLE_DB_TESTS=1 go test -p 1 -tags=integration ./internal/repository ./internal/modules/order
ENABLE_DB_TESTS=1 go test -tags=e2e ./internal/test/e2e/...
```

PowerShell: `$env:ENABLE_DB_TESTS = "1"` first.

The integration suite includes concurrency regressions: 50 concurrent
orders against 10 units of stock, concurrent cancels of one order,
opposite-order multi-item orders (deadlock), and concurrent stock removals.

See [documentation/TESTING.md](documentation/TESTING.md).

---

# CI Pipeline

GitHub Actions runs on pushes to branches and pull requests.

Current CI checks:

```text
Checkout
    |
    v
Set up Go
    |
    v
Verify formatting
    |
    v
Build
    |
    v
Unit tests
    |
    v
Repository integration tests
    |
    v
Order integration tests
    |
    v
HTTP end-to-end tests
```

A PostgreSQL service container is provided to CI for database-dependent tests.

---

# Makefile Commands

```bash
make run
make build
make test
make fmt
make tidy
make docker-up
make docker-down
make logs
make clean
```

---

# Load Testing

See [LOAD_TESTING.md](LOAD_TESTING.md) for every scenario, the dashboard
guide, measured results and troubleshooting. Quick start (PowerShell, API
running):

```powershell
cd locust
python -m venv .venv-locust; .\.venv-locust\Scripts\pip install -r requirements.txt
.\.venv-locust\Scripts\python seed.py --host http://localhost:8080
.\.venv-locust\Scripts\locust -f scenarios/smoke.py --host http://localhost:8080 -u 2 -r 2 -t 1m --headless
```

`seed.py` creates the load-test user, products and stock and writes
`locust/loadtest.env` (gitignored), which every scenario reads.

---

# Known limitations

- **No authorization model.** Any authenticated user can list all orders,
  read or cancel another user's order, set order status (including `PAID`
  without a payment), adjust inventory, and update or delete users. Roles and
  ownership checks are needed before this is exposed beyond local use.
- A failed payment (`payment.failed`) cancels the order, so a later
  successful retry on the same Razorpay order cannot mark it paid.
- `BaseModel.BeforeCreate` always assigns a new UUID, so creating a record
  with a populated association re-IDs that association.

---

# Documentation

- [Architecture](documentation/ARCHITECTURE.md)
- [Load Testing](LOAD_TESTING.md)
- [Observability](documentation/OBSERVABILITY.md)
- [Testing](documentation/TESTING.md)
- [Roadmap](documentation/ROADMAP.md)
- [Swagger API](docs/swagger.yaml)

---

# Author

**Khushi Desai**

Associate Software Engineer | Cloud & Backend Development

GitHub: https://github.com/khushidesai23

LinkedIn: https://www.linkedin.com/in/khushi-desai-ab5154225/

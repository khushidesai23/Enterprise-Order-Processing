# Enterprise Order Processing Platform

A backend-focused, enterprise-style Order Processing Platform built to demonstrate production-oriented backend engineering practices.

The project is implemented as a **modular monolith** in Go. It currently provides transactional order processing, inventory reservation, Razorpay payment integration, authentication, testing, OpenTelemetry tracing, Prometheus metrics, Grafana visualization, structured logging, Docker-based local infrastructure, and GitHub Actions CI.

> The project is intentionally evolving in phases. Advanced event-driven capabilities such as Debezium, Kafka, TimescaleDB, notification consumers, analytics consumers, and audit consumers are planned for later phases.

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
- Locust load-testing scripts

### Planned

- PostgreSQL CDC
- Debezium
- Kafka/event streaming
- Event consumers
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
│   └── grafana/
│       └── provisioning/
│
├── docs/
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
│
├── documentation/
│   ├── ARCHITECTURE.md
│   ├── DEVELOPMENT.md
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
│   ├── config.py
│   └── locustfile.py
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
- Python environment for Locust

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
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=order_processing
DB_SSLMODE=disable

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

---

# Start Local Infrastructure

Start all Docker services:

```bash
docker compose up -d
```

Current Compose services:

- PostgreSQL
- PostgreSQL test database
- Jaeger
- OpenTelemetry Collector
- Prometheus
- Grafana

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

Use the generated public URL in the Razorpay dashboard with the project's configured payment webhook endpoint.

Example:

```text
https://<your-ngrok-domain>/<payment-webhook-path>
```

The application verifies Razorpay webhook signatures before processing webhook data.

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

See [documentation/OBSERVABILITY.md](documentation/OBSERVABILITY.md) for details.

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

Run repository integration tests:

```bash
go test -tags=integration ./internal/repository
```

Run order service integration tests:

```bash
go test -tags=integration ./internal/modules/order
```

Run HTTP end-to-end tests:

```bash
go test -tags=e2e ./internal/test/e2e/...
```

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

Locust files are available in:

```text
locust/
```

They are intended to exercise the application under concurrent HTTP traffic and support performance experimentation during the observability phase.

---

# Documentation

- [Architecture](documentation/ARCHITECTURE.md)
- [Development Guide](documentation/DEVELOPMENT.md)
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

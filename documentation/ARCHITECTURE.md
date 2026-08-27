# Architecture

## 1. Architecture Style

The project is implemented as a modular monolith.

This means the application runs as one deployable backend process, while business capabilities are organized into separate modules with their own handlers, services, DTOs, mappers, errors, and routing.

This is appropriate for the current learning and implementation phase because it keeps local development manageable while preserving clear boundaries for future extraction.

## 2. Request Flow

```text
Client
  |
  v
Gin Router
  |
  v
Middleware
  |
  v
Handler
  |
  v
Service
  |
  v
Repository
  |
  v
PostgreSQL
```

### Responsibilities

| Layer | Responsibility |
|---|---|
| Router | Registers endpoints and middleware |
| Middleware | Authentication, CORS, request/metrics concerns |
| Handler | HTTP request parsing and response handling |
| Service | Business rules and workflow orchestration |
| Repository | Database access |
| Model | Persistent data representation |

## 3. Application Composition

The application startup is responsible for:

1. Loading configuration.
2. Registering Prometheus metrics.
3. Initializing OpenTelemetry tracing.
4. Initializing structured logging.
5. Connecting to PostgreSQL.
6. Running database migrations.
7. Building application dependencies and routes.
8. Starting the HTTP server.
9. Performing graceful shutdown.

## 4. Business Modules

```text
auth
user
category
product
inventory
order
payment
```

The modules interact through application-level service logic rather than network calls because the application is currently a single deployable unit.

## 5. Order and Inventory Consistency

Order creation and inventory reservation are critical business operations.

The intended consistency model is:

```text
Begin Transaction
      |
      v
Validate Order Request
      |
      v
Validate Stock
      |
      v
Reserve Inventory
      |
      v
Persist Order and Items
      |
      v
Commit
```

If a critical step fails, the transaction should fail rather than leaving partially applied business state.

## 6. Payment Architecture

```text
Application
    |
    v
Payment Service
    |
    v
Gateway Abstraction
    |
    v
Razorpay Gateway
```

This abstraction keeps payment-provider-specific logic separate from the main order workflow.

Webhook processing adds:

```text
Webhook Received
    |
    v
Verify Signature
    |
    v
Check Idempotency
    |
    v
Persist Event
    |
    v
Update Payment
    |
    v
Update Order
```

## 7. Observability Architecture

```text
HTTP Requests
    |
    +--> Prometheus middleware / metrics
    |
    +--> OpenTelemetry spans
    |
    +--> Structured logs
```

Tracing:

```text
Application -> OTEL Collector -> Jaeger
```

Metrics:

```text
Application /metrics -> Prometheus -> Grafana
```

Logs currently go to application stdout and are not implemented as a Loki pipeline.

## 8. Future Event-Driven Evolution

The planned future architecture is:

```text
PostgreSQL
    |
    v
WAL
    |
    v
Debezium
    |
    v
Kafka / Event Stream
    |
    +--> Notification Consumer
    +--> Analytics Consumer
    +--> Audit Consumer
```

This should be introduced after the core business model is stable.

## 9. Key Trade-off

A modular monolith has lower operational complexity today.

CDC and event streaming will add stronger decoupling later, but also introduce infrastructure, delivery semantics, schema evolution, consumer idempotency, observability, and operational overhead.

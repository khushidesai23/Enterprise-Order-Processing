# Architecture

## 1. Architecture Style

The backend is a modular monolith: one deployable Go process whose business
capabilities are separate modules, each with its own handler, service, DTOs,
mappers, errors and `RegisterRoutes`.

Modules call each other through service interfaces, not HTTP, so a module
can later be extracted without changing its callers.

```text
                    React storefront (/shop) + admin console (/admin)
                                         |
                                         v  /api/v1
Gin router -> middleware -> handler -> service -> repository -> PostgreSQL 16
    |                                                                |
    |  recovery, OTel, Prometheus, 10s request deadline,             | WAL (logical)
    |  access log, CORS, JWT + RequireRole (per route)               v
    |                                                     Debezium -> Kafka
    +-- /metrics -> Prometheus -> Grafana                            |
    +-- OTLP -> OTel Collector -> Jaeger          API CDC consumer <-+
```

## 2. Layers

| Layer | Package | Responsibility |
|---|---|---|
| Composition | `cmd/server`, `internal/app`, `internal/api/routes` | Config, telemetry, DB, migrations, CDC consumer, dependency wiring, routes, graceful shutdown |
| Middleware | `internal/api/middleware` | JWT auth, `RequireRole`, CORS, Prometheus, request deadline, access log |
| Authorization | `internal/authz` | The caller (`Principal`) and ownership checks |
| Handler | `internal/modules/<module>` | HTTP parsing, validation, response mapping |
| Service | `internal/modules/<module>` | Business rules and transactions |
| Repository | `internal/repository` | All database access (GORM) |
| Model | `internal/models` | Persistent data |
| Messaging | `internal/messaging/kafka` | Debezium CDC consumer |
| Shared | `pkg/logger`, `pkg/metrics`, `pkg/telemetry`, `pkg/utils`, `pkg/validator` | Zap logger, Prometheus collectors, OTel tracer, helpers |

## 3. Startup

1. Load configuration (environment variables, optional `.env`).
2. Register Prometheus metrics.
3. Initialize OpenTelemetry tracing (skipped when `OTEL_ENABLED=false`).
4. Initialize the Zap JSON logger.
5. Connect to PostgreSQL and run migrations.
6. Start the CDC consumer (skipped when `KAFKA_CDC_ENABLED=false`).
7. Build dependencies and routes, start the HTTP server.
8. On SIGINT/SIGTERM, stop the CDC consumer, drain the HTTP server, then
   flush the tracer and close the DB pool.

## 4. Business Modules

| Module | Responsibility |
|---|---|
| `auth` | Login (bcrypt, constant-time for unknown emails), JWT issue, `/auth/me` |
| `user` | Sign-up (always `customer`), profiles, admin user management, role changes |
| `category` | Category CRUD; a category with products cannot be deleted |
| `product` | Product CRUD, catalogue details, search with filters, sorting, paging and facets |
| `review` | One review per customer per product, verified-purchase flag, rating histogram |
| `inventory` | Stock levels and add/remove/reserve/release/confirm |
| `order` | Order creation with stock reservation, lifecycle, cancel, admin status summary |
| `payment` | Razorpay checkout, signature verification, webhooks, refunds, payment summary |

## 5. Order and Inventory Consistency

```text
Read user and product prices (outside the transaction)
      |
      v
Begin transaction
      |
      v
Insert order with total
      |
      v
For each item, in product-ID order:
  UPDATE inventories SET available -= q, reserved += q
  WHERE product_id = ? AND available_quantity >= q
      |
      v
Insert order items
      |
      v
Commit (any failure rolls everything back)
```

- The guarded `UPDATE` means stock can never go negative.
- Locking items in product-ID order prevents deadlocks between multi-item
  orders.
- Status changes lock the order row (`SELECT ... FOR UPDATE OF orders`), so
  reserved stock is released (cancel) or confirmed (ship) exactly once.

Lifecycle:

```text
CREATED -> PAYMENT_PENDING -> PAID -> PACKED -> SHIPPED -> DELIVERED
   |             |             |
   +-------------+-------------+--> CANCELLED (releases reserved stock)
```

## 6. Payment Architecture

```text
Payment Service -> PaymentGateway interface -> Razorpay gateway (razorpay-go)
                                            -> Stub gateway (tests)
```

Two paths complete a payment, and they are idempotent with each other:

- `POST /payments/verify`: the browser posts the Checkout result; the API
  checks the HMAC-SHA256 signature and ownership.
- `POST /payments/webhook`: Razorpay's server-to-server event.

```text
Webhook received
    |
    v
Verify X-Razorpay-Signature (bad -> 401)
    |
    v
Insert payment_webhooks row; duplicate X-Razorpay-Event-Id -> acknowledge, stop
    |
    v
Lock order, then re-read payment
    |
    v
Update payment (never backwards: a late failure cannot undo a capture)
    |
    v
Update order, commit (failure rolls back the webhook row so the retry is applied)
```

A failed attempt leaves the order `PAYMENT_PENDING` with stock reserved, and
checkout returns the same Razorpay order for a retry. Refunds go through
`PaymentGateway.RefundPayment` as the last step before commit; the
`refund.created` webhook covers a refund that succeeded when the commit
did not.

## 7. Change Data Capture

```text
PostgreSQL (wal_level=logical, publication + replication slot)
    |
    v
Debezium PostgreSQL connector (Kafka Connect)
    |
    v
Kafka topics order-processing.public.{categories,orders,inventories}
    |
    v
API CDC consumer (group order-processing-cdc-consumer)
    -> order_processing_cdc_* metrics and logs
```

The consumer currently counts and logs events; it has no business logic.
Delivery is at-least-once from Debezium, and the consumer commits offsets
before processing (at-most-once). Notification, analytics or audit consumers
need deduplication on the event's source position. See
[FAILURE_RECOVERY_DRILLS.pdf](FAILURE_RECOVERY_DRILLS.pdf).

## 8. Observability

| Signal | Path |
|---|---|
| Traces | otelgin + otelgorm -> OTLP -> OTel Collector -> Jaeger |
| Metrics | `/metrics` -> Prometheus -> provisioned Grafana dashboard |
| Logs | Zap JSON on stdout with `trace_id` and `span_id` |

Details are in [OBSERVABILITY.md](OBSERVABILITY.md).

## 9. Frontend

`frontend/` is a React 18 + TypeScript + Vite single-page app with two areas:

- the **meridian/** storefront at `/shop`;
- the admin console at `/admin`.

It uses TanStack Query against `/api/v1`. In development Vite proxies `/api`
to `:8080`; in production Netlify does. See
[frontend/README.md](../frontend/README.md).

## 10. Deployment

| Part | Local | Hosted |
|---|---|---|
| Frontend | Vite dev server (5173) | Netlify, proxies `/api/*` to Render |
| API | `go run ./cmd/server` on the host | Render (Docker, `Dockerfile`) |
| Database | `postgres` container | Neon |
| Kafka, Debezium, OTel, Prometheus, Grafana | Docker Compose | Not deployed (`OTEL_ENABLED=false`, `KAFKA_CDC_ENABLED=false`) |

See [DEPLOYMENT.md](../DEPLOYMENT.md).

## 11. Key Trade-offs

- A modular monolith keeps transactions local and operations simple. The
  cost is that modules scale together.
- CDC decouples future consumers from the write path. It adds
  infrastructure, at-least-once delivery, schema evolution and connector
  operations; the drills showed the connector needs a heartbeat and
  restart monitoring.
- Payment correctness relies on database locks and idempotency keys rather
  than distributed transactions.

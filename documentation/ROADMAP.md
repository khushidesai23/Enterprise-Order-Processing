# Project Roadmap

This roadmap reflects the implementation state as of October 2026.

## Phase 1 - Core Order Processing

### Completed

- Project structure, configuration (Viper, `.env`) and Docker Compose stack
- GORM models, migrations and repository layer
- Authentication (JWT HS256, bcrypt) and `customer`/`admin` roles with
  ownership checks
- Users, categories, products (catalogue details, search with facets),
  product reviews, inventory
- Transactional order creation with guarded, deadlock-free stock
  reservation
- Order lifecycle with row-locked status changes and cancellation
- Razorpay checkout, browser signature verification (`/payments/verify`),
  idempotent webhooks, refunds
- 10 s request deadline returning `504`
- Swagger documentation
- Unit, integration (including concurrency) and HTTP end-to-end tests in CI

## Phase 2 - Frontend and Deployment

### Completed

- React storefront (meridian/) and admin console
- Vitest unit tests and Playwright browser tests; frontend checks in CI
- Demo-store seeder (`cmd/seed`)
- Hosted deployment: Netlify (frontend), Render (API) and Neon (PostgreSQL)

## Phase 3 - Observability and Load Testing

### Completed

- OpenTelemetry tracing (HTTP and SQL spans) to Jaeger through the
  collector
- Prometheus metrics: HTTP, Go runtime, DB pool, CDC
- Provisioned Grafana dashboard
- Structured JSON logs with `trace_id`/`span_id`
- Health and readiness endpoints
- Locust suite of 11 scenarios with seeded data; results in
  [LOAD_TESTING.md](../LOAD_TESTING.md)

### Next

- Alert rules and SLOs
- PostgreSQL exporter and Debezium/Kafka Connect JMX metrics
- Log aggregation (e.g. Loki)

## Phase 4 - Event-Driven Architecture

### Completed

- PostgreSQL logical replication (publication, replication slot)
- Debezium connector publishing `categories`, `orders` and `inventories`
  changes to Kafka
- API CDC consumer with event, error, lag metrics

### Next

- Consumers with business logic (notifications, analytics, audit)
- Consumer idempotency keyed on the source position (Debezium redelivers
  after restarts)
- Commit offsets after processing; retry and dead-letter topics
- Event schema and versioning strategy
- Debezium heartbeat and connector health monitoring

## Phase 5 - Resilience

### Completed

- Failure and recovery drills (API crash, database outage, gateway timeout,
  Kafka outage, consumer crash, CDC replay, overload, backup and restore);
  see [FAILURE_RECOVERY_DRILLS.pdf](FAILURE_RECOVERY_DRILLS.pdf)

### Next (from the drill findings)

1. Pass a deadline to the Razorpay call and keep it out of the order lock;
   map gateway errors to `502`/`504` without leaking the upstream URL.
2. `Idempotency-Key` support on `POST /orders`, so a retry after a lost
   response does not create a duplicate order.
3. Debezium heartbeat (`heartbeat.interval.ms` + `heartbeat.action.query`)
   and automatic restart of a failed connector.
4. Return `503` instead of `500` when the database is unreachable.
5. Load shedding (concurrency limit) so overload fails fast with `503`.
6. Scheduled backups and WAL archiving for point-in-time recovery.

## Phase 6 - Time-Series and Data Lifecycle

### Planned

- TimescaleDB hypertables for operational time-series
- Continuous aggregates, retention and compression policies
- Hot/warm/cold lifecycle strategy

## Other

- Token revocation, or short-lived tokens with refresh, so role changes and
  deactivation take effect immediately
- Partial refunds and `refund.failed` handling

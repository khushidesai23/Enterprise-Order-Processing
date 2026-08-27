# Project Roadmap

This roadmap reflects the current implementation state.

## Phase 1 - Core Order Processing

### Completed

- Project structure and configuration
- Docker-based PostgreSQL
- Structured logging
- GORM models and migrations
- Repository layer
- Authentication and JWT
- User management
- Category management
- Product management
- Inventory management
- Order processing
- Transaction-safe inventory reservation
- Razorpay integration
- Checkout signature verification
- Webhook signature verification
- Idempotent payment webhook handling
- Swagger documentation
- Automated tests

## Phase 2 - Event-Driven Architecture

### Planned

- PostgreSQL logical replication preparation
- Change Data Capture
- Debezium
- Kafka or equivalent event transport
- Event schema/versioning strategy
- Notification consumer
- Analytics consumer
- Audit/event consumer
- Consumer idempotency
- Retry and dead-letter strategy

## Phase 3 - Observability

### Implemented

- OpenTelemetry initialization
- OTLP collector
- Jaeger tracing
- Prometheus metrics endpoint
- Prometheus scraping
- Grafana provisioning
- Health and readiness checks
- Structured logging
- Locust load-testing setup

### Next Observability Work

- Finalize Grafana dashboards
- Add focused business metrics
- Add database and dependency metrics where appropriate
- Define alert conditions
- Add SLO-oriented monitoring

## Phase 4 - Time-Series and Data Lifecycle

### Planned

- TimescaleDB
- Hypertables
- Time-series operational analytics
- Continuous aggregates where useful
- Retention policies
- Compression
- Hot/warm/cold lifecycle strategy

## Recommended Next Step

Complete and validate the Prometheus/Grafana dashboard layer before introducing CDC infrastructure.

The next major architectural phase should begin only after the current business workflows, tests, tracing, metrics, and dashboards are stable.

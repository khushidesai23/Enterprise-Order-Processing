# Testing

## Test Categories

### Unit Tests

Run:

```bash
go test ./...
```

These cover module-level behavior such as services and supporting components.

> Database-backed tests (integration, e2e) are skipped unless
> `ENABLE_DB_TESTS=1` is set. They use the `postgres-test` container on port
> 5433 and truncate its tables, so run packages one at a time (`-p 1`), e.g.
> `ENABLE_DB_TESTS=1 go test -p 1 -tags=integration ./internal/repository ./internal/modules/order`.

### Repository Integration Tests

Run:

```bash
go test -tags=integration ./internal/repository
```

These validate repository behavior against PostgreSQL.

### Order Service Integration Tests

Run:

```bash
go test -tags=integration ./internal/modules/order
```

These validate order workflows involving real persistence behavior.

### End-to-End Tests

Run:

```bash
go test -tags=e2e ./internal/test/e2e/...
```

These exercise HTTP-level flows.

## Test Database

Docker Compose provides a dedicated PostgreSQL test database on a separate configurable port.

Keep test and development data isolated.

## CI

GitHub Actions provides PostgreSQL and runs:

1. Formatting verification
2. Build
3. Unit tests
4. Repository integration tests
5. Order service integration tests
6. HTTP end-to-end tests

## Load Testing

Locust scripts are available under:

```text
locust/
```

Use them to generate concurrent traffic while observing metrics, traces, and application behavior. See [LOAD_TESTING.md](../LOAD_TESTING.md) for scenarios, commands and recorded results.

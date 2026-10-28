# Testing

## Summary

| Suite | Command | Needs |
|---|---|---|
| Go unit | `go test ./...` | Nothing |
| Repository integration | `go test -tags=integration ./internal/repository` | `ENABLE_DB_TESTS=1`, `postgres-test` container |
| Order service integration | `go test -tags=integration ./internal/modules/order` | Same |
| HTTP end-to-end | `go test -tags=e2e ./internal/test/e2e/...` | Same |
| Frontend typecheck + unit | `cd frontend; npm run typecheck; npm test` | Node 22 |
| Frontend browser end-to-end | `cd frontend; npx playwright test` | API running, an admin account, Razorpay test keys |
| Load | Locust, see [LOAD_TESTING.md](../LOAD_TESTING.md) | API and stack running, seeded data |

`make test` runs the unit tests; `make test-db` runs the integration and e2e
suites; `make frontend-test` and `make frontend-e2e` run the frontend
suites.

## Go Unit Tests

The unit tests sit next to the code they cover. They use mocks from
`internal/test/mocks` and a stub payment gateway. They cover:

- services: auth, user, category, product, inventory, order, payment;
- middleware: auth and roles, CORS, Prometheus, request deadline;
- route registration;
- utilities.

## DB-backed Tests

The integration and e2e suites run against the `postgres-test` container
(host port 5433), which is separate from the development database.

They are skipped unless `ENABLE_DB_TESTS=1` is set. In PowerShell:

```powershell
$env:ENABLE_DB_TESTS = "1"
go test -p 1 -tags=integration ./internal/repository ./internal/modules/order
go test -tags=e2e ./internal/test/e2e/...
```

The suites truncate the shared test database, so run packages one at a time
(`-p 1`).

- **Repository tests** check:
  - CRUD and constraints;
  - that stock decrements never go negative;
  - order and payment transactions;
  - product search, filters and facets;
  - reviews;
  - reuse of soft-deleted names, SKUs and emails.
- **Order service integration tests** include concurrency regressions:
  - 50 concurrent orders against 10 units of stock;
  - concurrent cancels of one order;
  - multi-item orders that lock products in opposite order (deadlock check);
  - concurrent stock removals.
- **HTTP e2e tests** (22 tests) drive the full router with the stub
  gateway. They cover:
  - catalogue and role checks;
  - checkout and verification;
  - every order lifecycle path (cancel, fulfilment, failed and retried
    payments, refunds before and after shipping, a capture after the
    customer cancelled);
  - webhook redelivery, concurrent deliveries, retries after failed
    processing and edge cases;
  - the admin order summary.

## Frontend Tests

- **Vitest**: 12 unit tests in `frontend/src/lib/lib.test.ts` cover:
  - the API client (envelope unwrapping, errors, sign-out on a rejected
    token);
  - error messages;
  - cart rules;
  - order status transitions;
  - formatting.
- **Playwright** specs in `frontend/e2e/` (`storefront.spec.ts`,
  `admin.spec.ts`) drive a real browser against the running API. They
  create real Razorpay test-mode orders, so they need test keys, internet
  access and a promoted admin (`ADMIN_EMAIL` / `ADMIN_PASSWORD`). See
  [frontend/README.md](../frontend/README.md#tests).

## CI

`.github/workflows/ci.yml` runs on pushes to any branch and on pull
requests.

**Backend job**, with a PostgreSQL service container:

1. gofmt check
2. Build
3. Unit tests
4. Repository integration tests
5. Order service integration tests
6. HTTP end-to-end tests

**Frontend job**:

1. `npm ci`
2. Typecheck
3. Vitest
4. Production build

Playwright is not run in CI because it needs Razorpay and an admin account.

## Load Testing

`locust/` holds the seeder (`seed.py`) and these scenarios:

- smoke, baseline, read-heavy, mixed, order-heavy;
- inventory contention, auth;
- spike, stress, recovery.

[LOAD_TESTING.md](../LOAD_TESTING.md) has the procedure, the dashboards to
watch and the measured results.

## Failure and Recovery Drills

The resilience drills, with their timelines, recovery times, data-loss
checks and follow-up fixes, are in
[FAILURE_RECOVERY_DRILLS.pdf](FAILURE_RECOVERY_DRILLS.pdf). They covered:

- API crash, database outage, payment gateway timeout;
- Kafka outage, consumer crash, CDC replay;
- duplicates, overload, backup and restore.

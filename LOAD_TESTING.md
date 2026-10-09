# Load Testing and Observability

This guide describes repeatable load testing for the current local development stack. User counts and durations below are **example test targets**, not service-level objectives or production capacity claims. The results section contains only values captured in executed local runs.

| Area | Existing implementation | Load-test relevance | Missing before this work |
|---|---|---|---|
| API and business flow | Gin routes for auth, products, categories, inventory, orders and payments; GORM services/repositories and transactional order creation | Scenarios exercise login, reads, order writes and stock contention | No repeatable scenario suite or recorded load results |
| Database | PostgreSQL, GORM and a configured SQL pool capped at 100 open / 10 idle connections | Transactions, inventory locking and pool pressure determine order correctness and latency | No exported SQL pool gauges/counters; order inventory read lacked a row lock |
| Metrics and dashboard | Prometheus HTTP/Go process metrics, Prometheus and Grafana services | Correlate request/error/latency behavior and host process state | No DB pool metric collection or dedicated load dashboard |
| Tracing and logs | OpenTelemetry Gin/GORM instrumentation, OTEL Collector, Jaeger and structured Zap logs | Inspect slow API and SQL spans alongside each run | Runtime trace delivery and span behavior needed validation |
| Load generation | Locust and a compatibility locustfile | Reuse existing setup for separate workload profiles | No focused scenarios, controlled fixture guidance or result capture |
| CDC/event pipeline | Kafka, Connect/Debezium and a Kafka reader are configured | CDC behavior could affect write load if functional | No registered connector, meaningful event processing, or CDC throughput/lag metrics |

## 1. Current architecture

The API is a Go 1.26.2 modular monolith using Gin, JWT middleware, GORM, and PostgreSQL. Product, category, inventory, order, payment, and auth modules use service and repository layers. Order creation creates the order, checks inventory, reserves stock, and writes order items in one GORM transaction. The SQL pool is configured for at most 100 open and 10 idle connections. `GET /api/v1/orders` and `/orders/me` accept `page` and `limit` (default 1 and 50; limit capped at 100) and return one page in the existing list response shape.

Routes used by the scenarios are `POST /api/v1/auth/login`, `GET /api/v1/products`, `GET /api/v1/categories`, `GET /api/v1/orders`, `POST /api/v1/orders`, and `GET /api/v1/health`. Business routes require a bearer token. Use Swagger at `/swagger/index.html` to confirm the API contract when the application changes.

## 2. Observability architecture and known gaps

- Gin middleware exports `order_processing_http_requests_total`, `order_processing_http_request_duration_seconds` and `order_processing_http_requests_in_flight`, labelled by method, route and status. Route labels use Gin's registered route template.
- The Prometheus Go client also exposes Go runtime and process metrics. Prometheus scrapes `host.docker.internal:8080` every 15 seconds.
- The application periodically copies `database/sql` pool stats into `order_processing_db_*` metrics. Open, in-use and idle are gauges; wait count and wait duration are cumulative counters. Collection runs every 15 seconds and once at startup/shutdown, not per request.
- Gin OTEL middleware creates server spans; `otelgorm` instruments GORM database activity without recording query variables. OTLP gRPC goes through the Collector to Jaeger.
- Zap logs are structured and written to stdout.
- Compose contains Kafka, Kafka Connect/Debezium, PostgreSQL logical WAL settings, and an application Kafka reader. The repository does **not** register a Debezium connector, and the consumer currently logs decoded events rather than performing business processing. Kafka throughput, consumer lag, and Debezium metrics are not exported by the application; CDC is not an end-to-end validated pipeline.
- Order creation reads the product inventory row with `FOR UPDATE` inside its transaction before checking and reserving stock. This prevents concurrent orders from passing the stock check against the same stale quantity.
- Order-list endpoints use bounded pagination after sustained-load traces showed that reading all order history made query and response cost grow with the entire database.

The provisioned Grafana dashboard is **Order Processing - Load Testing** in the Performance folder. It has route, method, and status selectors backed by labels that exist in the HTTP metrics. It includes traffic, latency, errors, Go runtime, in-flight requests and DB pool panels. The CDC row documents the missing metrics rather than showing invented data.

## 3. Prerequisites and startup

Install Docker Compose, Go 1.26.2 or compatible, Python 3, and Locust (`python -m pip install locust`). Configure `.env` from `.env.example`; use local-only values and do not put real secrets in source control. The application config currently requires Razorpay settings and a JWT secret even for these local load runs.

From the repository root in PowerShell:

```powershell
docker compose up -d
docker compose ps
go run ./cmd/server
```

In another terminal, set test credentials without putting them in the Locust source:

```powershell
$env:LOCUST_EMAIL = 'load-test-user@example.com'
$env:LOCUST_PASSWORD = '<local test account password>'
$env:LOCUST_PRODUCT_IDS = '<comma-separated UUIDs of dedicated high-stock products>'
Set-Location locust
```

Create a dedicated test account and product set through the existing API/Swagger or a local seed process. For mixed, spike, and sustained workloads, `LOCUST_PRODUCT_IDS` can spread orders across stocked products; provision enough aggregate stock for expected orders and replenish between runs. Order-heavy runs can use one high-stock product or a product pool. The load suite will not mutate business fixtures automatically. For contention, set only `LOCUST_PRODUCT_ID` to one dedicated product with known starting stock (for example 100 units), record initial available and reserved quantities, and do not run other writers against it. Do not point these scenarios at production data.

Endpoints: API `http://localhost:8080`, Prometheus `http://localhost:9090`, Grafana `http://localhost:3000`, Jaeger `http://localhost:16686`, Kafka Connect REST `http://localhost:8083`. The dashboard datasource uses the repository's provisioned Prometheus datasource. Grafana's current Compose credentials are development defaults; change them before exposing this stack beyond a local machine.

## 4. Locust scenarios

Run the following from the `locust` directory; `--host` selects the API. `-r` is users spawned per second and `-t` is run duration. User counts, waits and durations are workload inputs, not pass/fail thresholds.

```powershell
locust -f scenarios/smoke.py --host http://localhost:8080 --users 1 --spawn-rate 1 --run-time 2m --headless --csv results/smoke
```

| Scenario | File | Purpose and example invocation |
|---|---|---|
| Smoke | `scenarios/smoke.py` | One authenticated user; checks health and products. Use the command above, then check `/metrics`, Prometheus target/query, Grafana dashboard and Jaeger traces manually. |
| Baseline | `scenarios/baseline.py` | Weighted product/category/order reads plus order creation. Example: `locust -f scenarios/baseline.py --host http://localhost:8080 --users 5 --spawn-rate 1 --run-time 5m --headless --csv results/baseline`. |
| Read-heavy | `scenarios/read_heavy.py` | Task weights approximate products 60%, categories 30%, orders 10%. Repeat separately at 5, 10, 25, 50, then 100 users, only proceeding when the prior run remains healthy. Capture a separate CSV prefix per level. |
| Mixed workload | `scenarios/mixed_workload.py` | Authenticates at user startup; task weights are products 5, categories 3, orders 2, creates 1. Requires an existing high-stock load-test product and replenishment. |
| Order-heavy | `scenarios/order_heavy.py` | Repeated one-unit `POST /orders`; track stock, DB pool waits and SQL spans. Requires product ID and sufficient stock. |
| Inventory contention | `scenarios/inventory_contention.py` | Concurrent one-unit orders against the same product. Run at 10, 25, 50, then 100 users with a fresh/known stock state each time. The expected `400 insufficient stock available` response is counted as an expected rejection rather than a Locust failure; other errors remain failures and the 400 status remains visible in Prometheus. Confirm final available + reserved quantities and successful order quantities from the database. |
| Auth load | `scenarios/auth_load.py` | Repeats login to isolate password hashing/authentication. Login is the only request task; compare CPU and latency with baseline. Do not weaken hashing. |
| Spike | `scenarios/spike_test.py` with `scenarios/mixed_workload.py` | Illustrative shape: 10, 100, then 300 users. Run: `locust -f scenarios/mixed_workload.py,scenarios/spike_test.py --host http://localhost:8080 --headless --csv results/spike`. Tune values for the local host. |
| Sustained | `scenarios/sustained_load.py` | Example: `locust -f scenarios/sustained_load.py --host http://localhost:8080 --users 100 --spawn-rate 10 --run-time 30m --headless --csv results/sustained`. Observe during the run and several minutes after. |
| Stress progression | `scenarios/mixed_workload.py` | Run separate measured stages such as 50, 100, 200, 300, 500 users. Continue only while the host/database remain healthy. These are illustrative test steps, not capacity assertions. |

For read-heavy steps, use this PowerShell pattern and change `$users` and `$name` for each level:

```powershell
$users = 5
$name = "read-$users"
locust -f scenarios/read_heavy.py --host http://localhost:8080 --users $users --spawn-rate 1 --run-time 5m --headless --csv "results/$name"
```

Locust starts a login request when each authenticated user starts. For read-only runs this means login load is concentrated at ramp-up; `auth_load.py` is the dedicated repeated-login scenario. Keep Locust's user count, spawn rate, wait-time behavior and duration in each result record.

## 5. Prometheus queries

Use the dashboard selectors to filter actual method/route/status labels. Equivalent useful queries:

```promql
sum(rate(order_processing_http_requests_total[1m]))
sum by (route) (rate(order_processing_http_requests_total[1m]))
sum by (status) (rate(order_processing_http_requests_total[1m]))
sum(rate(order_processing_http_requests_total{status=~"5.."}[1m]))
100 * sum(rate(order_processing_http_requests_total{status=~"5.."}[1m])) / clamp_min(sum(rate(order_processing_http_requests_total[1m])), 1e-9)
histogram_quantile(0.50, sum by (le) (rate(order_processing_http_request_duration_seconds_bucket[5m])))
histogram_quantile(0.95, sum by (le) (rate(order_processing_http_request_duration_seconds_bucket[5m])))
histogram_quantile(0.99, sum by (le) (rate(order_processing_http_request_duration_seconds_bucket[5m])))
sum(order_processing_http_requests_in_flight)
go_goroutines
go_memstats_alloc_bytes
go_memstats_heap_alloc_bytes
process_resident_memory_bytes
rate(process_cpu_seconds_total[1m])
order_processing_db_open_connections
order_processing_db_in_use_connections
order_processing_db_idle_connections
rate(order_processing_db_wait_count_total[1m])
rate(order_processing_db_wait_duration_seconds_total[1m])
```

Wait count and wait duration are counters. Their `rate` indicates newly accumulated waiting over time; use `increase(...[test window])` for total waiting during a selected interval. A zero wait rate alone does not prove that the database is healthy; interpret pool saturation alongside request latency and traces.

## 6. Grafana, Jaeger and test correlation

For every run, record the wall-clock start/end time and use the same time range in Locust CSV, Grafana and Jaeger. In Grafana, choose the dashboard and filter route/method/status as needed. Use the Locust request names and API route labels to compare throughput, percentiles, status codes, process CPU/memory, goroutines, in-flight requests and DB pool usage.

In Jaeger, select the configured OTEL service name, narrow to the run window, and inspect slow `POST /api/v1/orders` or `POST /api/v1/auth/login` traces. Expand child spans and compare total server time with GORM/PostgreSQL spans. Long server duration with short database spans points toward application work or another dependency; long SQL spans suggest investigating query execution or lock waits. Confirm these interpretations from traces and metrics before naming a bottleneck. SQL instrumentation is configured, but actual trace delivery must be verified during the smoke run.

The current application has no explicit transaction-duration or PostgreSQL lock-wait Prometheus metric. GORM spans may show SQL duration, but do not assume a span identifies lock wait specifically. For stronger proof, collect PostgreSQL `pg_stat_activity`/lock observations during the controlled contention test or add appropriately scoped instrumentation in a later measured change.

## 7. Kafka / Debezium observations

Compose includes Kafka and Connect, but no connector is automatically registered in this project and the application consumer only logs supported Debezium event types. Therefore order load does not currently prove the PostgreSQL WAL → Debezium → Kafka → processing pipeline. No event throughput, consumer lag or Debezium error metrics are claimed or shown. A future CDC validation needs an explicit connector config/registration step, a meaningful consumer, and source/connector/topic/consumer measurements.

## 8. Result capture and comparison

Do not fill these fields from estimates. Save Locust CSV output and record Grafana/Jaeger observations for every executed run. Use `n/a` only when a metric is unavailable and explain why.

| Scenario | Users | RPS | P50 | P95 | P99 | Errors | CPU | Memory | Goroutines | In-flight | DB wait |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| Smoke | 1 | 0.51 | 5 ms | 27 ms | 120 ms | 0/29 | n/a | n/a | n/a | n/a | 0 observed |
| Baseline | 5 | 2.47 | 4 ms | 24 ms | 88 ms | 0/293 | n/a | n/a | n/a | n/a | 0 observed |
| Read-heavy | 5 | 2.44 | 5 ms | 13 ms | 86 ms | 0/143 | n/a | n/a | n/a | n/a | 0 observed |
| Read-heavy | 10 | 4.83 | 3 ms | 11 ms | 79 ms | 0/282 | n/a | n/a | n/a | n/a | 0 observed |
| Read-heavy | 25 | 12.26 | 4 ms | 12 ms | 91 ms | 0/728 | n/a | n/a | n/a | n/a | 0 observed |
| Read-heavy | 50 | 24.20 | 4 ms | 12 ms | 120 ms | 0/1,431 | n/a | n/a | n/a | n/a | 0 observed |
| Read-heavy | 100 | 48.35 | 4 ms | 15 ms | 150 ms | 0/2,871 | n/a | 27.9 MB heap | 36 | n/a | 0 observed |
| Mixed | 25 | 12.27 | 4 ms | 20 ms | 80 ms | 0/1,458 | n/a | n/a | n/a | n/a | 0 observed |
| Order-heavy | 25 | 12.44 | 8 ms | 25 ms | 77 ms | 0/1,480 | n/a | n/a | n/a | n/a | 0 observed |
| Inventory contention, before fix | 100 | 48.30 | 5 ms | 23 ms | 130 ms | 0/2,861* | n/a | n/a | n/a | n/a | 0 observed |
| Inventory contention, after fix | 100 | 48.53 | 6 ms | 60 ms | 160 ms | 0/2,829* | n/a | n/a | n/a | n/a | n/a |
| Auth | 25 | 11.62 | 84 ms | 130 ms | 210 ms | 0/1,381 | n/a | n/a | n/a | n/a | n/a |
| Spike, 10→100→300 users across 12 SKUs | 300 peak | 38.37 | 260 ms | 5.7 s | 8.8 s | 0/6,938 Locust failures | 2.61 CPU cores peak | 1,003 MB RSS peak | 705 peak | n/a | +6,815 waits / +2,071 s over observed 3 min window |
| Spike 10→100→300, paginated (12 SKUs) | 300 peak | 49.57 | 13 ms | 4.0 s | 5.7 s | 0/8,996 | 1.62 cores peak | 130 MB RSS peak | 581 peak | 147 peak | 3,502 waits / 832 s; 84 open / 82 in-use peak |
| Sustained 100, before pagination (18.4 min, interrupted) | 100 | 47.34 | 8 ms | 350 ms | 2.2 s | 0/52,437 | n/a | 1,689 MB RSS peak | 268 peak | 83 peak | 0 waits; 37 open / 28 in-use peak |
| Sustained 100, paginated (20 min) | 100 | 48.58 | 6 ms | 28 ms | 170 ms | 81/58,230 (0.139%) | 0.58 cores peak | 82.8 MB RSS peak | 200 peak | 83 peak | 0 waits; 10 open / 5 in-use peak |
| Stress beyond 300 | — | — | — | — | — | Not run: the 300-user spike saturated the DB pool | — | — | — | — | — |

Times are from Locust CSVs (percentiles are milliseconds). The spike used Locust's shaped 10→100→300 user stages and 12 dedicated products. Prometheus resource values are sampled maxima over each run; CPU is the maximum 30-second rate. The pre-pagination sustained run was stopped after 18.4 minutes as order-history latency and memory climbed. The post-pagination run completed its 20-minute target. Its 81 failures were transport disconnects/incomplete response bodies clustered within about four seconds; Prometheus recorded no 5xx responses, and app logs show successful HTTP responses for sampled affected requests. Their precise transport cause remains unproven. In one matching trace, Jaeger showed a 20.3-second HTTP span while the access log recorded 8 ms for the same trace; this timing discrepancy also needs investigation. The post-run API returned healthy/ready. The 100-user contention-after-fix aggregate includes the login ramp; database assertions are the correctness check: available 0, reserved 100, 100 committed units. Before the fix, the same starting stock produced available −2, reserved 102, and 102 committed units. `*` Contention's expected insufficient-stock rejections are marked successful by Locust; the failures column is the Locust failure count, not the number of rejected business requests.

### Findings and follow-up

- **Inventory correctness (high confidence):** the 100-user same-product run committed 102 units from 100 available, leaving −2 available and 102 reserved. Adding a transaction-scoped `FOR UPDATE` inventory row lock fixed the race. The rerun and a DB-backed 50-way integration regression both confirmed stock never went below zero and only the available number of orders succeeded.
- **Capacity ceiling at spike (high confidence for pool pressure):** before pagination, the 300-user multi-SKU spike sampled all 100 SQL connections in use, accumulated about 6,815 waits and 2,071 seconds of aggregate wait duration, and reached 38.37 RPS at P95/P99 5.7/8.8 seconds. After order-list pagination, the same spike reached 49.57 RPS and P95/P99 4.0/5.7 seconds; pool maxima were 84 open / 82 in use, with about 3,502 waits and 832 seconds of aggregate wait duration. RSS peaked near 130 MB and no HTTP 5xx or Locust failures occurred. The database pool did not hit its 100-connection ceiling, but waiting and multi-second tail latency remain at 300 users. The service returned healthy/ready and wait counters returned to zero in subsequent idle samples. Further stress beyond 300 users remains untested.
- **Unbounded order-history reads (high confidence):** before pagination, Jaeger captured a 7.58-second `GET /orders` trace with a 1.24 MB response. The orders query returned 8,423 rows and its `order_items` preload also read 8,423 rows; those SQL spans took about 7.49 and 4.93 seconds. Added `page`/`limit` parameters (default 50, maximum 100) to both order-list endpoints. In the 20-minute same-workload rerun, Locust `GET /orders` P95 fell from 1.6 seconds to 31 ms; a Jaeger page trace showed 50 orders and 50 items, SQL spans under 9 ms, and a 7.5 KB response. The run still had a brief cluster of transport failures described above, which needs a separate investigation.
- **CDC remains unvalidated:** no Debezium connector is registered and the consumer does not perform order processing, as described above.

For each run additionally record environment/host resources, API revision, dataset and initial/final stock, exact users/spawn/duration, endpoint mix, HTTP status distribution, Prometheus/Grafana observations, slow trace IDs and span durations, Kafka/CDC observations or why unavailable, suspected bottleneck, supporting evidence, confidence, and conclusion. Compare the same workload and data setup before/after any optimization.

## 9. Bottleneck identification method

1. Establish smoke and baseline without optimizing first.
2. Increase one workload variable at a time; record throughput, latency percentiles, failures and resource metrics.
3. Identify the first symptom (for example, throughput flattening, error growth, pool waits, CPU saturation, memory/goroutine growth, or slow SQL spans).
4. Gather evidence from Locust, Prometheus/Grafana and representative Jaeger traces in the same time window. For locking claims, collect PostgreSQL lock/activity evidence.
5. State observation, evidence, likely cause, affected component and confidence separately. A latency increase alone does not prove a cause.
6. Make one justified change only after the bottleneck is supported by evidence, then rerun the same test and compare against its baseline.
7. After spike/sustained tests, continue watching for several minutes to assess recovery. This remains a local environment result, not production capacity.

## 10. Limitations and future improvements

- The table above records the executed local runs. Raw Locust CSVs are kept under the ignored `locust/results/` directory and are local evidence, not committed artifacts.
- Load identities and product IDs must be supplied through environment variables; test data setup/replenishment is external and controlled by the operator.
- Database pool samples are periodic at 15 seconds; short bursts can occur between samples.
- No application Kafka/CDC throughput or lag metrics, registered Debezium connector, or functional event processing are currently available.
- Validate Prometheus target health, dashboard provisioning, and trace arrival locally; configuration files alone do not establish runtime health.
- Inventory concurrency has been checked against database state before and after adding the transaction row lock. Rerun the contention test after changes to order or inventory transaction behavior.
- Add CDC metrics/connector lifecycle, PostgreSQL lock/wait visibility, explicit result archival, and repeatable local fixture setup as follow-up work if those paths are in scope.

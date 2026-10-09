# Load Testing and Observability

How to run reproducible load tests against the Order Processing API, watch
them in Grafana/Jaeger, and read the results. Everything here was executed on
the machine described in [Test environment](#test-environment); numbers are
local measurements, not capacity claims for any other environment.

## Status

| Area | Status |
|---|---|
| Locust scenarios: smoke, baseline, read-heavy, mixed, order-heavy, contention, auth, spike, stress, recovery, sustained | Implemented and run (results below) |
| Controlled test data (`locust/seed.py`) with stock reset and an exhaustion guard | Implemented and used for every run |
| Grafana dashboard (provisioned), Prometheus scrape, Locust exporter, Jaeger datasource | Implemented; every dashboard query checked against live data |
| HTTP, Go runtime, DB pool (`go_sql_*`) and CDC metrics | Implemented and verified |
| CDC: PostgreSQL -> Debezium -> Kafka -> API consumer | Running and measured; the consumer records and logs events but has no business logic |
| Debezium/Kafka Connect JMX metrics, PostgreSQL server metrics | Not implemented (inspect by command, see [CDC](#cdc-setup)) |
| Distributed Locust (master/workers) | Not tested; the exporter only runs on the master/local runner |

## Test environment

| Item | Value |
|---|---|
| Host | Windows 11 Pro, single machine; Locust, the API and Docker Desktop share the CPU |
| API | `bin/server.exe` built from this branch, run on the host (`APP_ENV=development`, Gin debug mode, LOG_LEVEL=debug) |
| PostgreSQL | `postgres:16` in Docker Desktop (WSL2), host port 5434, `max_connections=100` |
| Kafka / Debezium | `quay.io/debezium/kafka:3.2` (KRaft) / `quay.io/debezium/connect:3.2` |
| Observability | Prometheus v3.11.3 (API and Locust scraped every 5s), Grafana 13.0.1, Jaeger 1.76.0, OTel Collector 0.158.0 |
| Locust | 2.46.3, Python 3.13, `wait_time = between(1, 3)` per user |

Two properties of this environment dominate some results and are called out
where they matter:

1. **Slow, erratic fsync on the Docker Desktop volume.** `pg_test_fsync` in the
   container measured 1.6 ms per single `fdatasync` but 9.5 ms for two 8 kB
   writes and 6-60 ms for `open_sync`. Commit latency, and therefore order
   P95/P99, is bound by this.
2. **Occasional Docker VM I/O stalls of 10-20 s.** During one run Kafka logged
   `Exceptionally slow controller event ... took 12309 ms` while PostgreSQL
   queries in the same VM stalled. See [Findings](#findings-and-optimizations).
3. **The machine entered Modern Standby once (13:48-14:38).** The first
   stress run is only valid up to its 300-user step and the sustained run
   started on wake was discarded; both were re-run after restarting Docker
   Desktop, whose IPv6 port forwarding had broken on resume.

## 1. Start the stack

PowerShell, from the repository root:

```powershell
Copy-Item .env.example .env        # first time only; then edit secrets/ports
docker compose up -d
docker compose ps                  # postgres, kafka, debezium healthy
go build -o bin\server.exe .\cmd\server
.\bin\server.exe                   # or: go run ./cmd/server
```

Check:

```powershell
curl.exe http://localhost:8080/api/v1/ready          # {"success":true,...}
start http://localhost:9090/targets                  # order-processing-api UP
```

> **Port 5432 conflict.** If a native PostgreSQL (e.g. the
> `postgresql-x64-18` Windows service) listens on 5432, `localhost:5432`
> reaches it instead of the container: the API then writes to a database that
> Debezium never reads, and results describe a different server. Check with
> `Get-NetTCPConnection -LocalPort 5432 -State Listen`; if the owner is not
> Docker, set `DB_PORT=5434` in `.env` and run `docker compose up -d postgres`.
> Results recorded before this branch ran against the native server.

### Local URLs

| Interface | URL |
|---|---|
| API | http://localhost:8080/api/v1 |
| Swagger UI | http://localhost:8080/swagger/index.html |
| API metrics | http://localhost:8080/metrics |
| Grafana dashboard | http://localhost:3000/d/order-processing-load-testing (admin/admin, local only) |
| Prometheus | http://localhost:9090 (targets: /targets) |
| Jaeger | http://localhost:16686 (service `enterprise-order-processing`) |
| Kafka Connect REST | http://localhost:8083/connectors |
| Locust web UI | http://localhost:8089 (when not `--headless`) |
| Locust exporter | http://localhost:9646/metrics (while Locust runs) |

### CDC setup

Needed once per PostgreSQL volume, after the API has started once (its
AutoMigrate creates the tables):

```powershell
Get-Content config\debezium\setup.sql | docker exec -i order-postgres psql -U postgres -d order_processing
curl.exe -X POST -H "Content-Type: application/json" --data "@config/debezium/postgres-connector.json" http://localhost:8083/connectors
curl.exe http://localhost:8083/connectors/order-processing-postgres-connector/status   # RUNNING / RUNNING
```

`setup.sql` is idempotent (role, grants, publication). The connector config
uses local development credentials (`cdc_user` / `cdc_password`).
`KAFKA_CDC_TOPIC` lists the topics the API consumes; topics appear when a
table first changes, and the consumer picks them up without a restart.

## 2. Prepare Locust and test data

```powershell
cd locust
python -m venv .venv-locust
.\.venv-locust\Scripts\pip install -r requirements.txt
.\.venv-locust\Scripts\python seed.py --host http://localhost:8080
```

`seed.py` (idempotent, API only) creates or resets:

| Fixture | Default | Used by |
|---|---|---|
| User `loadtest@example.com` | password from `LOCUST_PASSWORD`, else generated | all authenticated scenarios |
| 10 products `LOADTEST-0001..0010` | 1,000,000 available, 0 reserved each (`--stock`) | baseline, mixed, order-heavy, spike, stress, recovery, sustained |
| Product `LOADTEST-CONTENTION` | 100 available (`--contention-stock`) | inventory contention |

It writes `locust/loadtest.env` (gitignored), which every scenario loads, so
no credentials or IDs are typed by hand. **Re-run `seed.py` before each
measured run** to start from a known stock level.

Inventory exhaustion cannot silently distort a run:

- at test start every scenario checks that each product it uses has at least
  `LOCUST_MIN_STOCK` (default 10,000) units and aborts otherwise
  (`LOCUST_ALLOW_LOW_STOCK=1` overrides);
- outside the contention scenario, an "insufficient stock" response is a
  failure labelled *inventory exhausted, re-run seed.py*;
- available/reserved stock of every product is logged before and after the run.

## 3. Scenarios

Run from `locust/` with `.\.venv-locust\Scripts\locust` (shown as `locust`).
Add `--csv results/<name>` to keep CSVs (`results/` is gitignored). Drop
`--headless` to use the web UI at http://localhost:8089.

| Scenario | Command | Mix and intent |
|---|---|---|
| Smoke | `locust -f scenarios/smoke.py --host http://localhost:8080 -u 2 -r 2 -t 1m --headless` | health + products; verifies wiring, metrics and traces |
| Baseline | `locust -f scenarios/baseline.py --host http://localhost:8080 -u 5 -r 1 -t 3m --headless` | products 5 : categories 3 : orders list 1 : create 1 |
| Read-heavy | `locust -f scenarios/read_heavy.py --host http://localhost:8080 -u 50 -r 5 -t 3m --headless` | products 6 : categories 3 : orders list 1 |
| Mixed | `locust -f scenarios/mixed_workload.py --host http://localhost:8080 -u 50 -r 5 -t 4m --headless` | products 5 : categories 3 : orders list 2 : create 1 |
| Order-heavy | `locust -f scenarios/order_heavy.py --host http://localhost:8080 -u 50 -r 5 -t 3m --headless` | one-unit orders spread over the 10 products |
| Inventory contention | `python seed.py; locust -f scenarios/inventory_contention.py --host http://localhost:8080 -u 100 -r 20 -t 2m --headless` | one-unit orders on one 100-unit product; rejections are expected and counted as success |
| Auth | `locust -f scenarios/auth_load.py --host http://localhost:8080 -u 25 -r 5 -t 2m --headless` | repeated logins (bcrypt cost isolation) |
| Spike | `locust -f scenarios/mixed_workload.py,scenarios/spike_test.py --host http://localhost:8080 --headless` | 20 users 2 min -> 300 users 3 min -> 20 users 3 min |
| Stress | `locust -f scenarios/mixed_workload.py,scenarios/stress_test.py --host http://localhost:8080 --headless` | 25/50/100/200/300/400 users, 3 min per step |
| Recovery | `locust -f scenarios/mixed_workload.py,scenarios/recovery_test.py --host http://localhost:8080 --headless` | 20 users 2 min -> 400 users 5 min -> 20 users 5 min |
| Sustained | `locust -f scenarios/sustained_load.py --host http://localhost:8080 -u 100 -r 10 -t 15m --headless` | mixed reads/writes at constant load |

Shapes accept `LOCUST_SHAPE_SCALE` (e.g. `0.5`) to scale user counts on a
smaller machine. Every authenticated user logs in once at start, so ramps also
create a burst of bcrypt work. Verify contention correctness in the database:

```powershell
$p = (Select-String -Path loadtest.env -Pattern 'LOCUST_CONTENTION_PRODUCT_ID=(.*)').Matches.Groups[1].Value
docker exec order-postgres psql -U postgres -d order_processing -c "select available_quantity, reserved_quantity from inventories where product_id='$p'" -c "select sum(oi.quantity) units, count(distinct oi.order_id) orders from order_items oi join orders o on o.id=oi.order_id where oi.product_id='$p' and o.status<>'CANCELLED'"
```

Expected: available 0, reserved 100, 100 units in 100 orders.

## 4. What to watch

Dashboard: **Performance / Order Processing - Load Testing**, refresh 5s.
Route/method/status selectors filter the HTTP panels; the slow-trace table has
its own threshold selector.

| Row | Panels | Read it as |
|---|---|---|
| Load generator | Locust users; client vs server request rate; client failures; client vs server P95 | Server rate below client rate, or client P95 far above server P95, means time is lost before the handler (accept queue, connection resets, write timeouts) |
| Traffic | rate, by route, by status | Mix should match the scenario weights |
| Latency | P50/P95/P99; P95 by route | Which route degrades first |
| Errors | 5xx by route; 4xx by route/status; 5xx % | Contention 400s belong here, not in 5xx |
| Go runtime | CPU cores, RSS/heap, goroutines, in-flight | CPU near core count: CPU-bound (auth). In-flight rising while CPU flat: waiting on I/O |
| Database pool | open/in-use/idle/max; waits/s; wait seconds/s | In-use at max plus waits: pool-bound. In-use high without waits: requests hold connections while PostgreSQL works (commit/fsync) |
| CDC | events by table/op; commit->consume P50/P95; consumer lag and errors | Lag that rises with load: pipeline falling behind. A lag spike right after an API restart is the consumer-group rejoin |
| Slow traces | Jaeger search above the threshold | Open a trace; the gap after the last SQL statement in a transaction is the COMMIT |

Per scenario:

| Scenario | Primary graphs |
|---|---|
| Smoke/baseline | everything non-empty, 0 failures, targets UP |
| Read-heavy | P95 by route, CPU, in-flight |
| Mixed/sustained | P95 by route over time, RSS and goroutines (leaks show as steady growth), CDC lag |
| Order-heavy | POST /orders P95/P99, pool in-use vs waits, slow traces |
| Contention | 4xx rate (expected), 5xx (must be 0), POST /orders P99, DB invariant query |
| Auth | CPU (bcrypt), login P95 |
| Spike/recovery | client vs server rate and P95, pool, then time for P95/in-flight to return to the first stage's level |
| Stress | the step where server rate stops rising or P95/failures jump |

### Prometheus queries

```promql
sum(rate(order_processing_http_requests_total{job="order-processing-api"}[1m]))
histogram_quantile(0.95, sum by (le, route) (rate(order_processing_http_request_duration_seconds_bucket{job="order-processing-api"}[1m])))
sum by (status) (rate(order_processing_http_requests_total{job="order-processing-api"}[1m]))
rate(process_cpu_seconds_total{job="order-processing-api"}[1m])
process_resident_memory_bytes{job="order-processing-api"}
go_goroutines{job="order-processing-api"}
go_sql_in_use_connections{job="order-processing-api"}
rate(go_sql_wait_duration_seconds_total{job="order-processing-api"}[1m])
increase(go_sql_max_idle_closed_total{job="order-processing-api"}[15m])
sum by (table, operation) (rate(order_processing_cdc_events_total[1m]))
histogram_quantile(0.95, sum by (le) (rate(order_processing_cdc_end_to_end_lag_seconds_bucket[1m])))
sum(locust_users)
sum by (name) (rate(locust_requests_total{result="failure"}[1m]))
```

Always filter `go_*`/`process_*` by `job`: the Prometheus server exports the
same metric names (the earlier dashboard mixed them in).

### Jaeger

Service `enterprise-order-processing`, operation e.g. `POST /api/v1/orders`,
Min Duration `500ms`, lookback matching the run. SQL spans show the statement
(no bind values). `COMMIT` has no span: the empty gap between the last
statement of a transaction and the next span is commit time.

## 5. Results

Captured on this branch with the environment above. P50/P95/P99 are
Locust-measured milliseconds over the whole run (all requests, logins
included). Resource columns are Prometheus maxima over the run window (5s
scrape; CPU = 20s rate). Raw CSVs: `locust/results/v2-*` (local only).

| Run | Users | RPS | P50 | P95 | P99 | Failures | API CPU (cores) | RSS MB | Goroutines | Pool in-use (max) | Pool waits |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| Smoke | 2 | 1.0 | 5 | 24 | 58 | 0 / 60 | - | 73 | 38 | 0 | 0 |
| Baseline | 5 | 2.5 | 5 | 23 | 71 | 0 / 445 | 0.23 | 72 | 41 | 1 | 0 |
| Read-heavy | 50 | 24.7 | 6 | 23 | 94 | 0 / 4,429 | 0.41 | 76 | 91 | 2 | 0 |
| Mixed | 50 | 24.6 | 6 | 27 | 91 | 0 / 5,907 | 0.47 | 78 | 90 | 2 | 0 |
| Order-heavy, before optimization | 50 | 22.0 | 57 | 1,100 | 3,300 | 0 / 3,981 | 0.59 | 82 | 195 | 48 | 0 |
| Order-heavy, `synchronous_commit=off` (diagnostic only) | 50 | 24.6 | 26 | 98 | 190 | 0 / 2,457 | 0.71 | 77 | 100 | 4 | 0 |
| Order-heavy, after optimization | 50 | 23.5 | 13 | 420 | 2,700 | 0 / 4,221 | 0.45 | 79 | 207 | 49 | 0 |
| Inventory contention | 100 | 41.1 | 10 | 220 | 21,000 | 91 / 4,899 (see below) | 0.95 | 85 | 346 | 100 | 0 |
| Auth | 25 | 11.8 | 100 | 170 | 320 | 0 / 1,415 | 1.42 | 75 | 66 | 0 | 0 |
| Spike 20 -> 300 -> 20 (before pool/login fix) | 300 peak | 60.9 | 13 | 200 | 1,700 | 57 / 29,276 | 2.19 | 94 | 553 | 65 | 740 (65 s) |
| Recovery 20 -> 400 -> 20 (before pool/login fix) | 400 peak | 85.4 | 9 | 210 | 1,800 | 10 / 61,496 | 2.84 | 128 | 613 | 90 | 2,058 (646 s) |
| Stress 25 -> 400 (after all fixes) | 400 peak | 86.8 | 15 | 110 | 1,000 | 12 / 93,768 (HTTP 504 during a disk stall) | 1.37 | 103 | 588 | 78 | 1,369 (57 s) |
| Sustained 15 min (after all fixes) | 100 | 48.0 | 8 | 74 | 1,400 | 65 / 43,178 (HTTP 504 during disk stalls) | 0.78 | 93 | 259 | 45 | 65 (4 s) |
| Spike 20 -> 300 -> 20, after pool/login fix | 300 peak | 61.1 | 13 | 150 | 1,300 | **0** / 29,326 | 2.07 | 102 | 355 | 15 (open 80) | 824 (112 s), all during the ramp |

`POST /orders` alone: before 56/1,100/3,300 ms (P50/P95/P99), after
13/430/2,700 ms, with `synchronous_commit=off` 26/80/160 ms.

Run-specific notes:

- **Contention:** the database invariant held: available 0, reserved 100,
  exactly 100 units in 100 orders, no negative stock. The 91 failures and the
  21 s P99 come from a Docker VM stall during the run (next section), not
  from the contention logic.
- **Recovery:** with 400 users the server sustained ~198 req/s with
  server-side P95 mostly 20-90 ms. Two P95 peaks: 3 s during the ramp (400
  logins at 50/s saturate CPU on bcrypt; login P50 2.8 s) and 4 s at +240 s,
  which coincided with integration tests I ran in the same Docker VM, so treat
  that peak as interference. After the drop to 20 users, server P95 was back
  at the baseline level (23 ms vs 22 ms median) within the first 15 s window,
  and in-flight requests and pool in-use returned to 0. RSS stayed ~95 MB.
  The 10 failures were HTTP 500 caused by PostgreSQL `too many clients`
  (26 occurrences in the log), fixed afterwards; the recovery run itself was
  not repeated (the post-fix spike covers the same failure mode).
- **Stress per step** (server-side, last 120 s of each 3-minute step;
  Prometheus):

  | Users | Server req/s | Server P50/P95/P99 ms | Client P95 ms | API CPU | Pool in-use max | Pool waits | 5xx |
  |---:|---:|---|---:|---:|---:|---:|---:|
  | 25 | 12.4 | 5 / 33 / 138 | 39 | 0.19 | 1 | 0 | 0 |
  | 50 | 24.9 | 5 / 41 / 84 | 47 | 0.31 | 2 | 0 | 0 |
  | 100 | 46.8 | 5 / 48 / 1,521 | 56 | 0.42 | 23 | 14 | 10 |
  | 200 | 98.8 | 7 / 46 / 87 | 51 | 0.54 | 5 | 0 | 0 |
  | 300 | 148.0 | 9 / 53 / 94 | 78 | 0.73 | 9 | 0 | 0 |
  | 400 | 188.0 | 17 / 100 / 243 | 217 | 0.83 | 21 | 295 | 0 |

  Throughput tracks offered load (users / 2 s mean think time) up to 300
  users. At 400 users the server handled 188 of ~200 offered req/s, client
  P95 was twice server P95 and the pool started queueing: the first signs of
  saturation on this machine (Locust shares the CPU, so part of the client
  gap is the load generator). The 100-user step's P99 and all 12 failures
  are one stall at 15:12:50-59: Kafka logged a 7.7 s stalled write at
  15:12:45 and PostgreSQL logged `canceling statement due to user request`
  for inventory updates queued behind a stalled commit, i.e. the 10 s request
  deadline cancelled them, released their locks and returned 504. The same
  stall before the deadline fix produced silent connection drops.
- **Sustained (15 min, 100 users):** no leak trend: RSS 75-92 MB and
  goroutines 135-140 flat across the run (2-minute samples). All 65
  failures are 504s in three minutes (15:29, 15:31, 15:35) that coincide
  exactly with bursts of Kafka `Exceptionally slow controller event`
  disk stalls; PostgreSQL logged 30 deadline cancellations. Outside those
  windows server P95 was 24-32 ms.
- **Spike before vs after the pool/login fix** (same workload and Locust
  code): failures 57 -> 0, overall P95/P99 200/1,700 -> 150/1,300 ms,
  `POST /orders` P95/P99 330/1,900 -> 240/890 ms, max in-flight 72 -> 27,
  max goroutines 553 -> 355. All pool waits in the fixed run fall in the
  single 30 s window where the pool grew from 4 to 80 connections during the
  300-user ramp; afterwards the 80 idle connections were reused with no
  waits.
- **CDC:** steady-state commit->consume P95 was 1-5 s at 50-74 events/s with
  consumer lag ~0. Two artefacts to ignore: a 28 s P95 right after an API
  restart (the killed consumer's 30 s group session must expire before
  partitions are reassigned), and 300 s for the first 90 s of the fixed
  spike run (draining a 2,895-message backlog after a Docker restart).

## Findings and optimizations

Each item lists evidence, change, and the measured effect. Commits are on
`feature/load-testing`.

1. **App was not using the database Debezium reads (environment).** The API's
   connections went to a native PostgreSQL 18 service on 5432
   (`Get-NetTCPConnection` owner `postgresql-x64-18`); Docker's postgres only
   received Debezium. Moved the container to 5434 via `.env`. All results
   above use the container; results recorded before this branch do not.
2. **Inventory and order races (correctness).** DB-backed tests reproduced
   double release on concurrent cancels (2 of 10 succeeded; reserved 0 instead
   of 3) and deadlocks for multi-item orders listing products in opposite
   order (39 of 40 failed, SQLSTATE 40P01). Fixed by locking the order row and
   taking inventory locks in product-ID order; inventory decrements became
   single guarded `UPDATE ... WHERE qty >= ?` statements.
3. **Order latency: commit fsync + 14 statements per order.** Order-heavy:
   P95 1.1 s with 48 connections in use and no pool waits; `pg_stat_activity`
   showed `IO:WALSync` and `idle in transaction`; slow traces had all
   statements done by ~150 ms and then a ~500-620 ms gap (the commit). An A/B
   run with `synchronous_commit=off` (reverted immediately) cut P99 from
   2.9 s to 160 ms, proving commit flush dominates the tail. Code change:
   CreateOrder went from 14 statements to 5 (lean product lookup outside the
   transaction, total inserted directly, guarded update instead of
   lock+update, response built in memory). Effect: P50 56 -> 13 ms, P95
   1,100 -> 430 ms, P99 3,300 -> 2,700 ms. The remaining tail is the volume's
   fsync; on a real disk this would differ. `synchronous_commit=off` was *not*
   kept, because it trades durability.
4. **Silent transport failures (root cause of the earlier unexplained
   cluster).** Contention run: 91 Locust `HTTP 0` failures within 3 s; the
   server logged exactly 91 requests with 17-19 s handler latency and normal
   400 responses; Kafka logged a 12.3 s stalled write at the same moment.
   Handlers outlived `http.Server.WriteTimeout` (15 s), so Go closed the
   connections after the handler "succeeded". Added a 10 s request deadline:
   stalled queries are cancelled (verified: query returns at the deadline, row
   lock released) and the client gets a 504 that also shows in 5xx metrics.
5. **Pool larger than the server allows, and idle churn.** Spike: two logins
   failed with 401 (then those users' every request), and recovery: ten 500s.
   Jaeger showed `FATAL: sorry, too many clients already (SQLSTATE 53300)`:
   `MaxOpenConns=100` equalled `max_connections=100` while Debezium holds 3
   more. `MaxIdleConns=10` closed 716 connections during the spike, each later
   reopened. Pool is now configurable, default 80 open / 80 idle / 5 min idle
   timeout.
6. **Login error handling.** Any lookup error became 401, hiding the pool
   failure above; an unknown email panicked and returned 500. Unknown email is
   now 401 (with a dummy bcrypt compare for constant timing) and
   infrastructure errors are logged and returned as 500/504.
7. **Measurement fixes.** Dashboard runtime panels mixed the Prometheus
   server's own process metrics in (no `job` filter); `/metrics` scrapes were
   counted as API traffic and were the only slow samples in a smoke run
   (server P95 0.49 s vs client 21 ms); DB pool stats were sampled every 15 s.
   All fixed.

Not changed, with reasons:

- **bcrypt cost.** Auth is CPU-bound (~0.12 core-seconds per login), so ramps
  with hundreds of first logins queue on CPU. That is the price of password
  hashing; reduce login frequency (token reuse) rather than the cost.
- **Pool size beyond the server limit.** Raising `max_connections` would not
  help here: the bottleneck under order load is commit flush, not connections.

## 6. Troubleshooting

| Symptom | Cause / fix |
|---|---|
| `Stock below LOCUST_MIN_STOCK` and the run quits | Run `python seed.py` (resets stock). |
| `Set LOCUST_EMAIL and LOCUST_PASSWORD` | Run `seed.py` from `locust/`; it writes `loadtest.env`. |
| `seed.py`: login failed, user exists with another password | Set `$env:LOCUST_PASSWORD` to that user's password, or use `--email` for a new user. |
| Prometheus target `locust` DOWN | Normal when Locust is not running; with Locust running check port 9646 and `LOCUST_METRICS_PORT`. |
| Target `order-processing-api` DOWN | API not running on the host, or Docker cannot reach `host.docker.internal:8080`. |
| Grafana dashboard missing or empty | `docker compose restart grafana`; datasource UIDs `PBFA97CFB590B2093` (Prometheus) and `jaeger` must exist (Connections -> Data sources). |
| CDC panels empty | Connector status (`curl.exe http://localhost:8083/connectors/order-processing-postgres-connector/status`), topics in `KAFKA_CDC_TOPIC`, API using the container DB (port conflict above). |
| CDC lag spike after restarting the API | Consumer-group rejoin (about 30 s); disappears in steady state. |
| Locust `HTTP 0` / connection reset clusters | Check the API log for handler latency near 10 s and 504s, and Kafka logs for `Exceptionally slow controller event`: Docker VM stall. |
| `sorry, too many clients already` | `DB_MAX_OPEN_CONNS` plus other clients exceeds PostgreSQL `max_connections`. |
| A shaped run lasts far longer than its stages, max response time is minutes | The machine slept (Modern Standby) mid-run: check `Get-WinEvent -FilterHashtable @{LogName='System'; ProviderName='Microsoft-Windows-Kernel-Power'}` for events 506/507. Discard the run; keep the machine awake (plugged in, sleep disabled) during long tests. |
| After waking: `localhost:<port>` for Docker services times out but `127.0.0.1:<port>` works | Docker Desktop's IPv6 port forwarding broke on resume. `docker desktop restart`, then `docker compose up -d`. The API's open DB connections may keep working while new ones hang, so restart the API too. |
| Jaeger memory keeps growing | In-memory store; capped with `MEMORY_MAX_TRACES` in `docker-compose.yml` (it reached 3 GiB uncapped). |
| Gevent `RecursionError` when importing scenarios outside Locust | Import `locust` before `requests` (the `locust` CLI already does). |
| Git Bash: `docker exec ... /kafka/bin/...` resolves to `C:/Program Files/Git/...` | Prefix the command with `MSYS_NO_PATHCONV=1`. |

## 7. Limitations and pending work

- Results are from one Windows machine with Locust, the API and Docker on
  shared CPU and a slow virtual disk; repeat on the target environment before
  drawing capacity conclusions.
- Stage boundaries in the shapes are time-based; compare server-side windows
  (Prometheus) rather than Locust's cumulative percentiles for recovery.
- No authorization model (see README *Known limitations*); load tests use one
  user for all virtual users.
- No Debezium JMX or PostgreSQL exporter metrics; CDC consumer has no
  business processing.

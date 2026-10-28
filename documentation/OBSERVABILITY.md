# Observability

## Current Stack

| Signal | Technology |
|---|---|
| Traces | OpenTelemetry SDK (otelgin HTTP spans, otelgorm SQL spans without bind values) |
| Trace transport | OTLP gRPC |
| Trace collector | OpenTelemetry Collector 0.158.0 |
| Trace backend | Jaeger 1.76.0 (also a Grafana datasource) |
| Metrics | Prometheus client, scraped by Prometheus v3.11.3 every 5 s |
| Dashboards | Grafana 13.0.1 with a provisioned dashboard |
| Logs | Zap JSON on stdout with `trace_id` and `span_id` |

## Trace Flow

```text
Go application -> OpenTelemetry SDK -> OTel Collector (4317 gRPC / 4318 HTTP) -> Jaeger
```

Jaeger UI: http://localhost:16686 (service `enterprise-order-processing`).

Set `OTEL_ENABLED=false` where no collector runs, such as the Render
deployment.

## Metrics Flow

```text
API /metrics ----+
                 +--> Prometheus (http://localhost:9090) --> Grafana (http://localhost:3000)
Locust :9646 ----+
```

Prometheus runs in Docker and scrapes `host.docker.internal:8080`, so the
local setup assumes the API runs on the host. If the API is containerized,
change the target in `config/prometheus.yml` to the service name.

`/metrics` itself is excluded from metrics, traces and access logs.

### Application metrics

| Metric | Labels | Meaning |
|---|---|---|
| `order_processing_http_requests_total` | method, route, status | Requests handled (route = Gin template) |
| `order_processing_http_request_duration_seconds` | method, route, status | Handler latency histogram |
| `order_processing_http_requests_in_flight` | method, route | Requests being handled |
| `go_sql_*` | db_name | DB pool: open, in use, idle, waits, closed by limits |
| `order_processing_cdc_events_total` | table, operation | Debezium events consumed |
| `order_processing_cdc_errors_total` | stage | Consumer read/decode/process errors |
| `order_processing_cdc_end_to_end_lag_seconds` | | PostgreSQL commit to consumption |
| `order_processing_cdc_consumer_lag_messages` | | Messages behind the partition head |
| `go_*`, `process_*` | | Go runtime and process |

## Dashboard

`config/grafana/dashboards/load-testing.json` is provisioned as
**Performance / Order Processing - Load Testing**:

http://localhost:3000/d/order-processing-load-testing

It covers:

- Locust versus server traffic;
- latency percentiles, overall and per route;
- errors by status and route;
- Go runtime and DB pool;
- CDC events and lag;
- a table of slow traces from Jaeger.

[LOAD_TESTING.md](../LOAD_TESTING.md) explains which panels to watch for
each test.

## Logs

Every request writes one `http request` access-log line with the method,
path, query, status, latency, body size and client IP.

- Lines are logged at `info`, `warn` for 4xx, and `error` for 5xx or handler
  errors.
- Lines written with a request context carry `trace_id` and `span_id`, so a
  log line leads to its Jaeger trace. Business events are logged at `info` and
`warn`, for example:

- payment state changes;
- webhooks that are ignored;
- a capture on a cancelled order;
- CDC consumer errors.

Logs go to stdout only; there is no aggregation pipeline.

## Health Endpoints

| Endpoint | Purpose |
|---|---|
| `/api/v1/health` | Liveness (process is up) |
| `/api/v1/ready` | Readiness (database reachable) |
| `/api/v1/ping` | Simple round trip |
| `/api/v1/version` | Build version |

## CDC Checks

Debezium and Kafka Connect JMX metrics are not exported. Check the
connector and consumer group by command:

```powershell
curl.exe http://localhost:8083/connectors/order-processing-postgres-connector/status
docker exec order-kafka /kafka/bin/kafka-consumer-groups.sh --bootstrap-server kafka:9092 --describe --group order-processing-cdc-consumer
```

The failure drills showed that the connector can report `RUNNING` while
stuck. Also check the replication slot:

```sql
select slot_name, active, pg_size_pretty(pg_wal_lsn_diff(pg_current_wal_lsn(), confirmed_flush_lsn))
from pg_replication_slots;
```

## Not Implemented

- Log aggregation (Loki or similar)
- Alertmanager and alert rules
- SLO-based monitoring
- PostgreSQL exporter, Debezium JMX metrics

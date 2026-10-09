# Observability

## Current Stack

| Signal | Technology |
|---|---|
| Traces | OpenTelemetry (otelgin HTTP spans, otelgorm SQL spans) |
| Trace Transport | OTLP gRPC |
| Trace Collector | OpenTelemetry Collector 0.158.0 |
| Trace Backend | Jaeger 1.76.0 (all-in-one, in-memory) |
| Metrics | Prometheus client_golang, Prometheus v3.11.3 |
| Dashboards | Grafana 13.0.1 (provisioned dashboard + Prometheus and Jaeger datasources) |
| Load-generator metrics | Locust exporter in `locust/scenarios/common.py` |
| CDC | PostgreSQL logical replication -> Debezium 3.2 -> Kafka 3.2 -> API consumer |
| Logs | Structured Zap JSON to stdout |

## Trace Flow

```text
Go Application -> OpenTelemetry SDK -> OTEL Collector (4317 gRPC / 4318 HTTP) -> Jaeger
```

- Jaeger UI: http://localhost:16686 (service `enterprise-order-processing`)
- SQL spans carry the statement without bind values (`otelgorm.WithoutQueryVariables`).
- `COMMIT` is not a span. In a trace, the gap between the last statement of a
  transaction and the next span is the commit (WAL flush) time.
- `/metrics` scrapes are not traced.

## Metrics Flow

```text
Go Application /metrics --(5s)--> Prometheus --> Grafana
Locust :9646/metrics   --(5s)--> Prometheus --> Grafana
```

- API metrics: http://localhost:8080/metrics (Prometheus job `order-processing-api`)
- Prometheus: http://localhost:9090 (targets: http://localhost:9090/targets)
- Grafana: http://localhost:3000, dashboard **Performance / Order Processing - Load Testing**
  (http://localhost:3000/d/order-processing-load-testing)

### Application metrics

| Metric | Type | Labels | Meaning |
|---|---|---|---|
| `order_processing_http_requests_total` | counter | method, route, status | Requests handled (route = Gin template, e.g. `/api/v1/orders/:id`). `/metrics` itself is excluded. |
| `order_processing_http_request_duration_seconds` | histogram | method, route, status | Handler latency. |
| `order_processing_http_requests_in_flight` | gauge | method, route | Requests currently being handled. |
| `go_sql_{open,in_use,idle,max_open}_connections` | gauge | db_name | `database/sql` pool state, read at scrape time. |
| `go_sql_wait_count_total`, `go_sql_wait_duration_seconds_total` | counter | db_name | Waits for a free pool connection. |
| `go_sql_max_{idle,idle_time,lifetime}_closed_total` | counter | db_name | Connections closed by pool limits. |
| `order_processing_cdc_events_total` | counter | table, operation | Debezium change events consumed (c/u/d/r). |
| `order_processing_cdc_errors_total` | counter | stage (read/decode/process) | CDC consumer failures. |
| `order_processing_cdc_end_to_end_lag_seconds` | histogram | | PostgreSQL commit (`source.ts_ms`) to consumption. |
| `order_processing_cdc_consumer_lag_messages` | gauge | | Messages behind the head of the most recently fetched partition. |
| `go_*`, `process_*` | | | Go runtime and process (CPU, RSS, goroutines, GC). Filter with `job="order-processing-api"`; the Prometheus server exports the same names. |

### Locust metrics (only while Locust runs)

| Metric | Labels | Meaning |
|---|---|---|
| `locust_users` | | Running simulated users. |
| `locust_requests_total` | method, name, result | Requests sent; `result="failure"` includes transport errors. |
| `locust_request_duration_seconds` | method, name | Client-observed latency. |

## CDC pipeline

```text
PostgreSQL (wal_level=logical, publication order_processing_cdc)
  -> Debezium connector order-processing-postgres-connector (Kafka Connect :8083)
  -> topics order-processing.public.<table>
  -> API consumer (group order-processing-cdc-consumer, topics from KAFKA_CDC_TOPIC)
```

The consumer currently records and logs events; it performs no business
processing. Setup is described in [LOAD_TESTING.md](../LOAD_TESTING.md#cdc-setup).

Not exported to Prometheus (inspect by command instead):

- Debezium/Kafka Connect JMX metrics (needs a JMX exporter agent):
  `curl http://localhost:8083/connectors/order-processing-postgres-connector/status`
- Replication slot lag:
  `docker exec order-postgres psql -U postgres -d order_processing -c "select slot_name, active, pg_size_pretty(pg_wal_lsn_diff(pg_current_wal_lsn(), confirmed_flush_lsn)) from pg_replication_slots"`
- Exact per-partition consumer lag:
  `docker exec order-kafka /kafka/bin/kafka-consumer-groups.sh --bootstrap-server kafka:9092 --describe --group order-processing-cdc-consumer`
  (Git Bash: prefix with `MSYS_NO_PATHCONV=1`)

## Health Endpoints

```text
/api/v1/health
/api/v1/ready
/api/v1/ping
/api/v1/version
```

## Deployment detail

Prometheus runs in Docker and scrapes `host.docker.internal:8080` (API) and
`host.docker.internal:9646` (Locust), so the API and Locust are expected to run
on the host. If the API is containerized, change the target to its service name.

## Not implemented

- Loki log aggregation
- Alertmanager, alert rules, SLOs
- PostgreSQL server metrics (postgres_exporter) and Debezium JMX metrics
- Trace exemplars on latency histograms

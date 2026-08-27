# Observability

## Current Stack

| Signal | Technology |
|---|---|
| Traces | OpenTelemetry |
| Trace Transport | OTLP |
| Trace Collector | OpenTelemetry Collector |
| Trace Backend | Jaeger |
| Metrics | Prometheus |
| Dashboards | Grafana |
| Logs | Structured application logs to stdout |

## Trace Flow

```text
Go Application
    |
    v
OpenTelemetry SDK
    |
    v
OTEL Collector
    |
    v
Jaeger
```

The collector accepts OTLP traffic on:

```text
gRPC: 4317
HTTP: 4318
```

Jaeger UI:

```text
http://localhost:16686
```

## Metrics Flow

```text
Go Application
    |
    v
/metrics
    |
    v
Prometheus
    |
    v
Grafana
```

Application metrics endpoint:

```text
http://localhost:8080/metrics
```

Prometheus:

```text
http://localhost:9090
```

Grafana:

```text
http://localhost:3000
```

## Health Endpoints

```text
/api/v1/health
/api/v1/ready
/api/v1/ping
/api/v1/version
```

## Important Current Deployment Detail

Prometheus runs in Docker and is configured to scrape:

```text
host.docker.internal:8080
```

Therefore the current local setup assumes the API runs on the host machine.

If the API is later containerized, the Prometheus target should be changed to the Docker service name.

## Current Scope

Implemented:

- OpenTelemetry initialization
- OTLP export
- Jaeger trace backend
- Prometheus endpoint
- Prometheus scraping
- Grafana datasource provisioning
- Health/readiness endpoints
- Structured logging
- Load-testing scripts

Not currently implemented as part of the stack:

- Loki log aggregation
- Alertmanager
- Production alert rules
- SLO/SLA alerting
- CDC-specific metrics

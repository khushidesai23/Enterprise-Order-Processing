# Locust load-testing suite

See the repository-level [LOAD_TESTING.md](../LOAD_TESTING.md) for prerequisites, test data setup, exact commands, Prometheus queries, Grafana/Jaeger procedures, CDC limitations, and the blank result template.

Run Locust commands from this directory so the shared `scenarios` package resolves. Provide credentials with `LOCUST_EMAIL` and `LOCUST_PASSWORD`. Order scenarios accept `LOCUST_PRODUCT_ID` for one product or `LOCUST_PRODUCT_IDS` as a comma-separated product pool. Use one known-stock ID for contention; use several stocked IDs for general mixed/spike/sustained traffic to avoid testing one hot inventory row unintentionally. No credential or product ID is embedded as a default.

"""Shared Locust building blocks for the Order Processing API.

Besides the user base classes this module, when imported by any scenario:

* loads ``locust/loadtest.env`` (written by ``seed.py``) for variables that
  are not already set in the environment;
* starts a Prometheus exporter on ``LOCUST_METRICS_PORT`` (default 9646, set
  0 to disable) so Grafana can plot Locust users, client-side request rate,
  failures and latency next to the API metrics;
* checks at test start that every load-test product still has at least
  ``LOCUST_MIN_STOCK`` units available, so inventory exhaustion cannot turn
  a latency test into a stream of "insufficient stock" failures;
* prints the inventory of the products under test when the run stops.
"""
import logging
import os
import random
from pathlib import Path

# locust must be imported before requests: it applies gevent monkey-patching.
from locust import HttpUser, LoadTestShape, between, events
from locust.runners import WorkerRunner

import requests  # noqa: E402,I001

log = logging.getLogger("order-processing")

ENV_FILE = Path(__file__).resolve().parent.parent / "loadtest.env"


def _load_env_file(path):
    if not path.is_file():
        return
    for line in path.read_text(encoding="utf-8").splitlines():
        line = line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        os.environ.setdefault(key.strip(), value.strip())


_load_env_file(ENV_FILE)


def env_list(name):
    return [value.strip() for value in os.getenv(name, "").split(",") if value.strip()]


def configured_product_ids():
    return env_list("LOCUST_PRODUCT_IDS") or env_list("LOCUST_PRODUCT_ID")


def contention_product_ids():
    return env_list("LOCUST_CONTENTION_PRODUCT_ID") or env_list("LOCUST_PRODUCT_ID")


class AuthenticatedUser(HttpUser):
    abstract = True
    wait_time = between(1, 3)
    # Minimum available stock per product required before the run starts.
    # 0 disables the check (used by the contention scenario).
    min_stock = int(os.getenv("LOCUST_MIN_STOCK", "10000"))

    @classmethod
    def product_pool(cls):
        return configured_product_ids()

    def on_start(self):
        self.token = None
        self.email = os.getenv("LOCUST_EMAIL", "")
        self.password = os.getenv("LOCUST_PASSWORD", "")
        self.product_ids = self.product_pool()
        if not self.email or not self.password:
            raise RuntimeError("Set LOCUST_EMAIL and LOCUST_PASSWORD (or run seed.py) for authenticated scenarios")
        self.login()

    def login(self):
        with self.client.post(
            "/api/v1/auth/login",
            json={"email": self.email, "password": self.password},
            name="POST /auth/login",
            catch_response=True,
        ) as response:
            if response.status_code != 200:
                response.failure(f"Login failed: HTTP {response.status_code}")
                return
            try:
                self.token = response.json().get("data", {}).get("token")
            except (ValueError, AttributeError):
                response.failure("Login returned invalid JSON")
                return
            if not self.token:
                response.failure("Login response did not contain data.token")

    def headers(self):
        # Retry a failed start-up login instead of sending every later
        # request unauthenticated (one failed login otherwise turns into a
        # stream of 401s that hides the original failure).
        if not self.token:
            self.login()
        return {"Authorization": f"Bearer {self.token}"} if self.token else {}

    def create_order(self, quantity=1, allow_insufficient_stock=False):
        if not self.product_ids:
            raise RuntimeError("Set LOCUST_PRODUCT_IDS (or run seed.py) to existing load-test product(s)")
        product_id = random.choice(self.product_ids)
        with self.client.post(
            "/api/v1/orders",
            json={"items": [{"product_id": product_id, "quantity": quantity}]},
            headers=self.headers(),
            name="POST /orders",
            catch_response=True,
        ) as response:
            if response.status_code in (200, 201):
                return
            insufficient = False
            if response.status_code == 400:
                try:
                    insufficient = response.json().get("error") == "insufficient stock available"
                except (ValueError, AttributeError):
                    pass
            if insufficient and allow_insufficient_stock:
                response.success()
                return
            if insufficient:
                response.failure("Insufficient stock: inventory exhausted, re-run seed.py before measuring")
                return
            response.failure(f"Order returned HTTP {response.status_code}")


class ReadTasks:
    def get_products(self):
        self.client.get("/api/v1/products", headers=self.headers(), name="GET /products")

    def get_categories(self):
        self.client.get("/api/v1/categories", headers=self.headers(), name="GET /categories")

    def get_orders(self):
        self.client.get(
            "/api/v1/orders",
            params={"page": 1, "limit": 50},
            headers=self.headers(),
            name="GET /orders",
        )


def random_quantity():
    return random.randint(1, 2)


class StagedShape(LoadTestShape):
    """Runs ``stages`` in sequence: (duration_seconds, users, spawn_rate).

    ``LOCUST_SHAPE_SCALE`` multiplies user counts (e.g. 0.5 on a small
    machine) without editing the stage tables.
    """

    abstract = True
    stages = []

    def tick(self):
        scale = float(os.getenv("LOCUST_SHAPE_SCALE", "1"))
        elapsed = self.get_run_time()
        for duration, users, spawn_rate in self.stages:
            if elapsed < duration:
                return max(1, round(users * scale)), spawn_rate
            elapsed -= duration
        return None


# ---------------------------------------------------------------------------
# Run-level hooks. Guarded so importing this module under two names (e.g.
# "common" and "scenarios.common") registers them only once.
# ---------------------------------------------------------------------------

def _api_session(host):
    email, password = os.getenv("LOCUST_EMAIL", ""), os.getenv("LOCUST_PASSWORD", "")
    if not host or not email or not password:
        return None
    session = requests.Session()
    response = session.post(f"{host}/api/v1/auth/login", json={"email": email, "password": password}, timeout=10)
    response.raise_for_status()
    session.headers["Authorization"] = f"Bearer {response.json()['data']['token']}"
    return session


def _inventory(session, host, product_id):
    response = session.get(f"{host}/api/v1/inventory/{product_id}", timeout=10)
    response.raise_for_status()
    return response.json()["data"]


def _products_under_test(environment):
    products, min_stock = set(), 0
    for user_class in environment.user_classes:
        if not issubclass(user_class, AuthenticatedUser):
            continue
        pool = user_class.product_pool()
        products.update(pool)
        if pool:
            min_stock = max(min_stock, user_class.min_stock)
    return sorted(products), min_stock


def _on_test_start(environment, **_):
    if isinstance(environment.runner, WorkerRunner):
        return
    products, min_stock = _products_under_test(environment)
    if not products:
        return
    try:
        session = _api_session(environment.host)
        if session is None:
            return
        low = []
        for product_id in products:
            inventory = _inventory(session, environment.host, product_id)
            log.info("inventory before run %s available=%s reserved=%s", product_id,
                     inventory["available_quantity"], inventory["reserved_quantity"])
            if inventory["available_quantity"] < min_stock:
                low.append(f"{product_id} ({inventory['available_quantity']} available)")
    except requests.RequestException as exc:
        log.warning("inventory pre-check skipped: %s", exc)
        return
    if low and os.getenv("LOCUST_ALLOW_LOW_STOCK") != "1":
        log.error("Stock below LOCUST_MIN_STOCK=%s for %s. Re-run seed.py or set LOCUST_ALLOW_LOW_STOCK=1.",
                  min_stock, ", ".join(low))
        environment.process_exit_code = 2
        environment.runner.quit()


def _on_test_stop(environment, **_):
    if isinstance(environment.runner, WorkerRunner):
        return
    products, _ = _products_under_test(environment)
    try:
        session = _api_session(environment.host)
        if session is None:
            return
        for product_id in products:
            inventory = _inventory(session, environment.host, product_id)
            log.info("inventory after run %s available=%s reserved=%s", product_id,
                     inventory["available_quantity"], inventory["reserved_quantity"])
    except requests.RequestException as exc:
        log.warning("inventory post-run snapshot skipped: %s", exc)


def _start_exporter(environment, **_):
    if isinstance(environment.runner, WorkerRunner):
        return
    port = int(os.getenv("LOCUST_METRICS_PORT", "9646"))
    if port <= 0:
        return
    try:
        from prometheus_client import REGISTRY, Counter, Histogram, start_http_server
        from prometheus_client.core import GaugeMetricFamily
    except ImportError:
        log.warning("prometheus_client not installed; Locust metrics exporter disabled (pip install -r requirements.txt)")
        return

    requests_total = Counter("locust_requests", "Requests issued by Locust.", ["method", "name", "result"])
    duration = Histogram(
        "locust_request_duration_seconds", "Client-observed request duration.", ["method", "name"],
        buckets=(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30),
    )

    class UsersCollector:
        def collect(self):
            gauge = GaugeMetricFamily("locust_users", "Locust users currently running.")
            runner = environment.runner
            gauge.add_metric([], runner.user_count if runner else 0)
            yield gauge

    REGISTRY.register(UsersCollector())

    def on_request(request_type, name, response_time, exception, **__):
        requests_total.labels(request_type, name, "failure" if exception else "success").inc()
        duration.labels(request_type, name).observe((response_time or 0) / 1000)

    environment.events.request.add_listener(on_request)
    start_http_server(port)
    log.info("Locust Prometheus exporter listening on :%s/metrics", port)


if not getattr(events, "_order_processing_hooks", False):
    events._order_processing_hooks = True
    events.init.add_listener(_start_exporter)
    events.test_start.add_listener(_on_test_start)
    events.test_stop.add_listener(_on_test_stop)

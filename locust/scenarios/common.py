import os
import random

from locust import HttpUser, between


class AuthenticatedUser(HttpUser):
    abstract = True
    wait_time = between(1, 3)

    def on_start(self):
        self.token = None
        self.email = os.getenv("LOCUST_EMAIL", "")
        self.password = os.getenv("LOCUST_PASSWORD", "")
        configured_products = os.getenv("LOCUST_PRODUCT_IDS", "")
        self.product_ids = [value.strip() for value in configured_products.split(",") if value.strip()]
        if not self.product_ids:
            single_product = os.getenv("LOCUST_PRODUCT_ID", "").strip()
            if single_product:
                self.product_ids = [single_product]
        if not self.email or not self.password:
            raise RuntimeError("Set LOCUST_EMAIL and LOCUST_PASSWORD for authenticated scenarios")
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
        return {"Authorization": f"Bearer {self.token}"} if self.token else {}

    def create_order(self, quantity=1, allow_insufficient_stock=False):
        if not self.product_ids:
            raise RuntimeError("Set LOCUST_PRODUCT_ID or LOCUST_PRODUCT_IDS to existing load-test product(s)")
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
            if allow_insufficient_stock and response.status_code == 400:
                try:
                    if response.json().get("error") == "insufficient stock available":
                        response.success()
                        return
                except (ValueError, AttributeError):
                    pass
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

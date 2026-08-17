import os
import random

from locust import HttpUser, between, task


class OrderProcessingUser(HttpUser):
    wait_time = between(1, 3)

    def on_start(self):
        self.token = None

        self.email = os.getenv(
            "LOCUST_EMAIL",
            "admin@example.com",
        )

        self.password = os.getenv(
            "LOCUST_PASSWORD",
            "Admin@123",
        )

        self.product_id = os.getenv(
            "LOCUST_PRODUCT_ID",
            "3cfd7483-2081-441a-b640-884316909367",
        )

        self.login()

    def login(self):
        with self.client.post(
            "/api/v1/auth/login",
            json={
                "email": self.email,
                "password": self.password,
            },
            name="POST /auth/login",
            catch_response=True,
        ) as response:

            if response.status_code != 200:
                response.failure(
                    f"Login failed: {response.status_code}"
                )
                return

            try:
                body = response.json()
            except Exception:
                response.failure("Invalid JSON response")
                return

            self.token = (
                body.get("data", {})
                .get("token")
            )

            if not self.token:
                response.failure(
                    "JWT token not found"
                )

    def auth_headers(self):
        return {
            "Authorization": f"Bearer {self.token}"
        }

    @task(5)
    def get_products(self):
        self.client.get(
            "/api/v1/products",
            headers=self.auth_headers(),
            name="GET /products",
        )

    @task(3)
    def get_categories(self):
        self.client.get(
            "/api/v1/categories",
            headers=self.auth_headers(),
            name="GET /categories",
        )

    @task(2)
    def get_orders(self):
        self.client.get(
            "/api/v1/orders",
            headers=self.auth_headers(),
            name="GET /orders",
        )

    @task(1)
    def create_order(self):
        if not self.product_id:
            return

        payload = {
            "items": [
                {
                    "product_id": self.product_id,
                    "quantity": random.randint(1, 2),
                }
            ]
        }

        with self.client.post(
            "/api/v1/orders",
            json=payload,
            headers=self.auth_headers(),
            name="POST /orders",
            catch_response=True,
        ) as response:

            if response.status_code not in (200, 201):
                response.failure(
                    f"Create order failed: "
                    f"{response.status_code}"
                )
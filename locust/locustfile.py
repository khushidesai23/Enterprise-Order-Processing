import os

from locust import HttpUser, between, task


class OrderProcessingUser(HttpUser):
    wait_time = between(1, 2)

    def on_start(self):
        self.token = None

        email = os.getenv(
            "LOCUST_EMAIL",
            "admin@example.com",
        )

        password = os.getenv(
            "LOCUST_PASSWORD",
            "Admin@123",
        )

        self.login(email, password)

    def login(self, email, password):
        with self.client.post(
            "/api/v1/auth/login",
            json={
                "email": email,
                "password": password,
            },
            name="POST /auth/login",
            catch_response=True,
        ) as response:

            if response.status_code != 200:
                response.failure(
                    f"Login failed: HTTP {response.status_code}"
                )
                return

            try:
                body = response.json()
            except Exception:
                response.failure("Invalid JSON response")
                return

            # Adjust this according to your actual API response.
            token = (
                body.get("data", {})
                .get("token")
            )

            if not token:
                response.failure(
                    f"JWT token not found: {body}"
                )
                return

            self.token = token

    def auth_headers(self):
        return {
            "Authorization": f"Bearer {self.token}"
        }

    @task(5)
    def get_categories(self):
        self.client.get(
            "/api/v1/categories",
            headers=self.auth_headers(),
            name="GET /categories",
        )

    @task(5)
    def get_products(self):
        self.client.get(
            "/api/v1/products",
            headers=self.auth_headers(),
            name="GET /products",
        )

    @task(3)
    def get_orders(self):
        self.client.get(
            "/api/v1/orders",
            headers=self.auth_headers(),
            name="GET /orders",
        )
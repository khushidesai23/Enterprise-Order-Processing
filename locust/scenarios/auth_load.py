import os

from locust import HttpUser, between, task


class AuthLoadUser(HttpUser):
    wait_time = between(1, 3)

    def on_start(self):
        self.email = os.getenv("LOCUST_EMAIL", "")
        self.password = os.getenv("LOCUST_PASSWORD", "")
        if not self.email or not self.password:
            raise RuntimeError("Set LOCUST_EMAIL and LOCUST_PASSWORD for auth load")

    @task
    def login(self):
        with self.client.post(
            "/api/v1/auth/login",
            json={"email": self.email, "password": self.password},
            name="POST /auth/login",
            catch_response=True,
        ) as response:
            if response.status_code != 200:
                response.failure(f"Login failed: HTTP {response.status_code}")

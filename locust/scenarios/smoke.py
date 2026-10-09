from locust import task

try:
    from .common import AuthenticatedUser, ReadTasks
except ImportError:
    from common import AuthenticatedUser, ReadTasks


class SmokeUser(AuthenticatedUser, ReadTasks):
    @task
    def health(self):
        self.client.get("/api/v1/health", name="GET /health")

    @task
    def products(self):
        self.get_products()

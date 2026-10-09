from locust import task

try:
    from .common import AuthenticatedUser
except ImportError:
    from common import AuthenticatedUser


class OrderHeavyUser(AuthenticatedUser):
    @task
    def create_order(self):
        self.create_order_request()

    def create_order_request(self):
        super().create_order(1)

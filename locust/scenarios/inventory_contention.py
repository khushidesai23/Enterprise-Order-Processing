from locust import task

try:
    from .common import AuthenticatedUser
except ImportError:
    from common import AuthenticatedUser


class InventoryContentionUser(AuthenticatedUser):
    @task
    def reserve_same_product(self):
        # Use one unit per request against the same explicitly seeded product.
        # Compare successful orders, rejected orders, and remaining stock.
        super().create_order(1, allow_insufficient_stock=True)

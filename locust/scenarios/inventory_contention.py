from locust import task

try:
    from .common import AuthenticatedUser, contention_product_ids
except ImportError:
    from common import AuthenticatedUser, contention_product_ids


class InventoryContentionUser(AuthenticatedUser):
    """Concurrent one-unit orders against one product with limited stock.

    Uses LOCUST_CONTENTION_PRODUCT_ID (written by seed.py). "Insufficient
    stock" responses are the expected outcome once stock runs out and are
    counted as successes; correctness is checked from the final inventory:
    available + reserved must equal the seeded stock and available must
    never be negative.
    """

    min_stock = 0

    @classmethod
    def product_pool(cls):
        return contention_product_ids()

    @task
    def reserve_same_product(self):
        self.create_order(1, allow_insufficient_stock=True)

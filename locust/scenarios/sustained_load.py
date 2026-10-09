from locust import task

try:
    from .common import AuthenticatedUser, ReadTasks, random_quantity
except ImportError:
    from common import AuthenticatedUser, ReadTasks, random_quantity


class SustainedLoadUser(AuthenticatedUser, ReadTasks):
    @task(5)
    def products(self):
        self.get_products()

    @task(3)
    def categories(self):
        self.get_categories()

    @task(1)
    def orders(self):
        self.get_orders()

    @task(1)
    def create_order(self):
        super().create_order(random_quantity())

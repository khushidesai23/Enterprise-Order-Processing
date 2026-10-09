from locust import task

try:
    from .common import AuthenticatedUser, ReadTasks
except ImportError:
    from common import AuthenticatedUser, ReadTasks


class ReadHeavyUser(AuthenticatedUser, ReadTasks):
    @task(6)
    def products(self):
        self.get_products()

    @task(3)
    def categories(self):
        self.get_categories()

    @task(1)
    def orders(self):
        self.get_orders()

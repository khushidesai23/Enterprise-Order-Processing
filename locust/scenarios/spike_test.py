from locust import LoadTestShape

class SpikeShape(LoadTestShape):
    # Illustrative local test targets. Tune to the host and database limits.
    stages = [
        {"duration": 60, "users": 10, "spawn_rate": 2},
        {"duration": 120, "users": 100, "spawn_rate": 30},
        {"duration": 180, "users": 300, "spawn_rate": 50},
    ]

    def tick(self):
        run_time = self.get_run_time()
        for stage in self.stages:
            if run_time < stage["duration"]:
                return stage["users"], stage["spawn_rate"]
        return None

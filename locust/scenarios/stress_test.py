"""Stress: step the user count up until latency or errors break down.

Pair with a user class, e.g.
    locust -f scenarios/mixed_workload.py,scenarios/stress_test.py ...
Each step holds long enough for P95 and pool metrics to settle. Note the
step at which throughput stops rising, P95 grows sharply, or pool waits
start; that is the local saturation point. Scale with LOCUST_SHAPE_SCALE.
"""
try:
    from .common import StagedShape
except ImportError:
    from common import StagedShape


class StressShape(StagedShape):
    stages = [
        (180, 25, 5),
        (180, 50, 10),
        (180, 100, 20),
        (180, 200, 40),
        (180, 300, 50),
        (180, 400, 50),
    ]

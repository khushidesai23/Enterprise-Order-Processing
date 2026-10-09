"""Recovery: sustained overload, then baseline, to measure recovery time.

Pair with a user class, e.g.
    locust -f scenarios/mixed_workload.py,scenarios/recovery_test.py ...
Compare the final baseline stage against the first: P95, error rate,
in-flight requests, pool in-use/waits, goroutines and memory should return
to their baseline levels. Record how long that takes after the drop.
"""
try:
    from .common import StagedShape
except ImportError:
    from common import StagedShape


class RecoveryShape(StagedShape):
    stages = [
        (120, 20, 5),     # baseline
        (300, 400, 50),   # overload
        (300, 20, 100),   # recovery
    ]

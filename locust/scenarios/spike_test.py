"""Spike: steady baseline, a sudden burst, then back to baseline.

Pair with a user class, e.g.
    locust -f scenarios/mixed_workload.py,scenarios/spike_test.py ...
Watch latency/pool waits during the burst and how quickly they return to the
baseline level afterwards. Scale users with LOCUST_SHAPE_SCALE.
"""
try:
    from .common import StagedShape
except ImportError:
    from common import StagedShape


class SpikeShape(StagedShape):
    # (duration seconds, users, spawn rate per second)
    stages = [
        (120, 20, 5),     # baseline
        (180, 300, 100),  # spike
        (180, 20, 100),   # back to baseline: recovery window
    ]

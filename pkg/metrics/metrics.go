package metrics

import (
	"database/sql"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	registerOnce       sync.Once
	dbOpenConnections  = prometheus.NewGauge(prometheus.GaugeOpts{Namespace: "order_processing", Subsystem: "db", Name: "open_connections", Help: "Current number of open database connections."})
	dbInUseConnections = prometheus.NewGauge(prometheus.GaugeOpts{Namespace: "order_processing", Subsystem: "db", Name: "in_use_connections", Help: "Current number of in-use database connections."})
	dbIdleConnections  = prometheus.NewGauge(prometheus.GaugeOpts{Namespace: "order_processing", Subsystem: "db", Name: "idle_connections", Help: "Current number of idle database connections."})
	dbWaitCount        = prometheus.NewCounter(prometheus.CounterOpts{Namespace: "order_processing", Subsystem: "db", Name: "wait_count_total", Help: "Cumulative count of waits for a database connection."})
	dbWaitDuration     = prometheus.NewCounter(prometheus.CounterOpts{Namespace: "order_processing", Subsystem: "db", Name: "wait_duration_seconds_total", Help: "Cumulative time spent waiting for a database connection, in seconds."})

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "order_processing",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{
			"method",
			"route",
			"status",
		},
	)

	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "order_processing",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{
			"method",
			"route",
			"status",
		},
	)

	HTTPRequestsInFlight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "order_processing",
			Subsystem: "http",
			Name:      "requests_in_flight",
			Help:      "Current number of HTTP requests being processed.",
		},
		[]string{
			"method",
			"route",
		},
	)
)

func Register() {
	registerOnce.Do(func() {
		prometheus.MustRegister(
			HTTPRequestDuration,
			HTTPRequestsTotal,
			HTTPRequestsInFlight,
			dbOpenConnections,
			dbInUseConnections,
			dbIdleConnections,
			dbWaitCount,
			dbWaitDuration,
		)
	})
}

// StartDBPoolCollector periodically exports database/sql pool statistics.
// The scrape path and HTTP request path never query the pool directly.
func StartDBPoolCollector(db *sql.DB, interval time.Duration) func() {
	if interval <= 0 {
		interval = 15 * time.Second
	}
	stop := make(chan struct{})
	done := make(chan struct{})
	update := func() {
		stats := db.Stats()
		dbOpenConnections.Set(float64(stats.OpenConnections))
		dbInUseConnections.Set(float64(stats.InUse))
		dbIdleConnections.Set(float64(stats.Idle))
		dbWaitCount.Add(float64(stats.WaitCount) - dbWaitCountValue)
		dbWaitDuration.Add(stats.WaitDuration.Seconds() - dbWaitDurationValue)
		dbWaitCountValue = float64(stats.WaitCount)
		dbWaitDurationValue = stats.WaitDuration.Seconds()
	}
	update()
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				update()
			case <-stop:
				update()
				return
			}
		}
	}()
	return func() { close(stop); <-done }
}

var (
	dbWaitCountValue    float64
	dbWaitDurationValue float64
)

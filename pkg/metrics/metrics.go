package metrics

import (
	"database/sql"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var (
	registerOnce sync.Once

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

	// CDCEventsTotal counts Debezium change events consumed from Kafka.
	CDCEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "order_processing",
			Subsystem: "cdc",
			Name:      "events_total",
			Help:      "Debezium change events consumed, by source table and operation.",
		},
		[]string{
			"table",
			"operation",
		},
	)

	// CDCErrorsTotal counts consumer failures by stage (read, decode, process).
	CDCErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "order_processing",
			Subsystem: "cdc",
			Name:      "errors_total",
			Help:      "CDC consumer errors by stage.",
		},
		[]string{
			"stage",
		},
	)

	// CDCEndToEndLag is the delay between the source database commit
	// (Debezium source.ts_ms) and the application consuming the event.
	CDCEndToEndLag = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "order_processing",
			Subsystem: "cdc",
			Name:      "end_to_end_lag_seconds",
			Help:      "Seconds from PostgreSQL commit to CDC event consumption.",
			Buckets:   []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60, 300},
		},
	)
)

func Register() {
	registerOnce.Do(func() {
		prometheus.MustRegister(
			HTTPRequestDuration,
			HTTPRequestsTotal,
			HTTPRequestsInFlight,
			CDCEventsTotal,
			CDCErrorsTotal,
			CDCEndToEndLag,
		)
	})
}

// RegisterCDCConsumerLag exports the Kafka consumer lag (messages behind
// the head of the most recently fetched partition), read at scrape time.
func RegisterCDCConsumerLag(lag func() int64) error {
	return prometheus.Register(
		prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: "order_processing",
				Subsystem: "cdc",
				Name:      "consumer_lag_messages",
				Help:      "Kafka consumer lag in messages for the most recently fetched partition.",
			},
			func() float64 { return float64(lag()) },
		),
	)
}

// RegisterDBStats exports database/sql pool statistics as go_sql_* metrics
// labelled db_name. Values are read from db.Stats() at scrape time.
func RegisterDBStats(db *sql.DB, dbName string) error {
	return prometheus.Register(
		collectors.NewDBStatsCollector(db, dbName),
	)
}

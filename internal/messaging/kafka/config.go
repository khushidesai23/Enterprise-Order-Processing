package kafka

import "time"

type Config struct {
	Brokers []string
	GroupID string
	// Topics lists the Debezium topics to consume. A single topic uses a
	// plain group reader; several topics use consumer-group topics.
	Topics   []string
	MinBytes int
	MaxBytes int
	MaxWait  time.Duration
}

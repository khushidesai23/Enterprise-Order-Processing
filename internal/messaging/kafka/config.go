package kafka

import "time"

type Config struct {
	Brokers     []string
	GroupID     string
	Topic       string
	MinBytes    int
	MaxBytes    int
	MaxWait     time.Duration
}
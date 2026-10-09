package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/metrics"
)

type Consumer struct {
	reader *kafka.Reader
	logger *zap.Logger
}

type DebeziumEvent struct {
	Payload DebeziumPayload `json:"payload"`
}

type DebeziumPayload struct {
	Before    json.RawMessage `json:"before"`
	After     json.RawMessage `json:"after"`
	Source    DebeziumSource  `json:"source"`
	Operation string          `json:"op"`
	Timestamp int64           `json:"ts_ms"`
}

type DebeziumSource struct {
	Version     string `json:"version"`
	Connector   string `json:"connector"`
	Name        string `json:"name"`
	Database    string `json:"db"`
	Schema      string `json:"schema"`
	Table       string `json:"table"`
	Transaction *int64 `json:"txId"`
	LSN         *int64 `json:"lsn"`
	// Timestamp is the source database commit time in milliseconds.
	Timestamp int64 `json:"ts_ms"`
}

func NewConsumer(cfg Config, logger *zap.Logger) *Consumer {
	readerConfig := kafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		GroupID:     cfg.GroupID,
		MinBytes:    cfg.MinBytes,
		MaxBytes:    cfg.MaxBytes,
		MaxWait:     cfg.MaxWait,
		StartOffset: kafka.FirstOffset,
		// Debezium creates a table's topic on its first change. Without
		// watching, a group that joined before the topic existed is never
		// assigned its partitions.
		WatchPartitionChanges: true,
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
			logger.Sugar().Warnf("kafka reader: "+msg, args...)
		}),
	}
	if len(cfg.Topics) == 1 {
		readerConfig.Topic = cfg.Topics[0]
	} else {
		readerConfig.GroupTopics = cfg.Topics
	}

	reader := kafka.NewReader(readerConfig)

	return &Consumer{
		reader: reader,
		logger: logger,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	const maxRetryDelay = 30 * time.Second
	retryDelay := time.Second

	readerConfig := c.reader.Config()
	topics := readerConfig.GroupTopics
	if readerConfig.Topic != "" {
		topics = []string{readerConfig.Topic}
	}

	c.logger.Info(
		"Kafka CDC consumer started",
		zap.Strings("topics", topics),
	)

	for {
		message, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				c.logger.Info("Kafka CDC consumer stopped")
				return ctx.Err()
			}

			metrics.CDCErrorsTotal.WithLabelValues("read").Inc()

			c.logger.Error(
				"failed to read Kafka message",
				zap.Error(err),
				zap.Duration("retry_in", retryDelay),
			)

			timer := time.NewTimer(retryDelay)
			select {
			case <-ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				c.logger.Info("Kafka CDC consumer stopped")
				return ctx.Err()
			case <-timer.C:
			}

			retryDelay *= 2
			if retryDelay > maxRetryDelay {
				retryDelay = maxRetryDelay
			}
			continue
		}
		retryDelay = time.Second

		// Tombstones (nil value) follow deletes when enabled on the
		// connector; they carry no change payload.
		if len(message.Value) == 0 {
			continue
		}

		var event DebeziumEvent

		if err := json.Unmarshal(message.Value, &event); err != nil {
			metrics.CDCErrorsTotal.WithLabelValues("decode").Inc()

			c.logger.Error(
				"failed to decode Debezium event",
				zap.Error(err),
				zap.String("topic", message.Topic),
				zap.Int("partition", message.Partition),
				zap.Int64("offset", message.Offset),
			)

			continue
		}

		metrics.CDCEventsTotal.WithLabelValues(
			event.Payload.Source.Table,
			event.Payload.Operation,
		).Inc()
		if event.Payload.Source.Timestamp > 0 {
			metrics.CDCEndToEndLag.Observe(
				time.Since(time.UnixMilli(event.Payload.Source.Timestamp)).Seconds(),
			)
		}

		c.logger.Debug(
			"CDC event received",
			zap.String("topic", message.Topic),
			zap.Int("partition", message.Partition),
			zap.Int64("offset", message.Offset),
			zap.String("operation", event.Payload.Operation),
			zap.String("table", event.Payload.Source.Table),
		)

		if err := c.processEvent(ctx, event); err != nil {
			metrics.CDCErrorsTotal.WithLabelValues("process").Inc()

			c.logger.Error(
				"failed to process CDC event",
				zap.Error(err),
			)
		}
	}
}

func (c *Consumer) processEvent(
	ctx context.Context,
	event DebeziumEvent,
) error {
	switch event.Payload.Operation {
	case "c":
		return c.handleCreate(event)

	case "u":
		return c.handleUpdate(event)

	case "d":
		return c.handleDelete(event)

	case "r":
		return c.handleSnapshot(event)

	default:
		return fmt.Errorf(
			"unsupported Debezium operation: %s",
			event.Payload.Operation,
		)
	}
}

func (c *Consumer) handleCreate(event DebeziumEvent) error {
	c.logger.Info(
		"CDC CREATE event",
		zap.String("table", event.Payload.Source.Table),
	)

	return nil
}

func (c *Consumer) handleUpdate(event DebeziumEvent) error {
	c.logger.Info(
		"CDC UPDATE event",
		zap.String("table", event.Payload.Source.Table),
	)

	return nil
}

func (c *Consumer) handleDelete(event DebeziumEvent) error {
	c.logger.Info(
		"CDC DELETE event",
		zap.String("table", event.Payload.Source.Table),
	)

	return nil
}

func (c *Consumer) handleSnapshot(event DebeziumEvent) error {
	c.logger.Info(
		"CDC SNAPSHOT event",
		zap.String("table", event.Payload.Source.Table),
	)

	return nil
}

// Lag reports how many messages the reader is behind the head of the
// partition it fetched from most recently.
func (c *Consumer) Lag() int64 {
	return c.reader.Stats().Lag
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

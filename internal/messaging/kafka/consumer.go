package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
	logger *slog.Logger
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
	Version      string `json:"version"`
	Connector    string `json:"connector"`
	Name         string `json:"name"`
	Database     string `json:"db"`
	Schema       string `json:"schema"`
	Table        string `json:"table"`
	Transaction  *int64 `json:"txId"`
	LSN          *int64 `json:"lsn"`
}

func NewConsumer(
	cfg Config,
	logger *slog.Logger,
) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		GroupID:     cfg.GroupID,
		Topic:       cfg.Topic,
		MinBytes:    cfg.MinBytes,
		MaxBytes:    cfg.MaxBytes,
		MaxWait:     cfg.MaxWait,
		StartOffset: kafka.FirstOffset,
	})

	return &Consumer{
		reader: reader,
		logger: logger,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	for {
		message, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			c.logger.Error(
				"failed to read Kafka message",
				"error",
				err,
			)

			continue
		}

		var event DebeziumEvent

		if err := json.Unmarshal(message.Value, &event); err != nil {
			c.logger.Error(
				"failed to decode Debezium event",
				"error",
				err,
				"topic",
				message.Topic,
				"partition",
				message.Partition,
				"offset",
				message.Offset,
			)

			continue
		}

		c.logger.Info(
			"CDC event received",
			"topic",
			message.Topic,
			"partition",
			message.Partition,
			"offset",
			message.Offset,
			"operation",
			event.Payload.Operation,
			"table",
			event.Payload.Source.Table,
		)

		if err := c.processEvent(ctx, event); err != nil {
			c.logger.Error(
				"failed to process CDC event",
				"error",
				err,
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
		"table",
		event.Payload.Source.Table,
	)

	return nil
}

func (c *Consumer) handleUpdate(event DebeziumEvent) error {
	c.logger.Info(
		"CDC UPDATE event",
		"table",
		event.Payload.Source.Table,
	)

	return nil
}

func (c *Consumer) handleDelete(event DebeziumEvent) error {
	c.logger.Info(
		"CDC DELETE event",
		"table",
		event.Payload.Source.Table,
	)

	return nil
}

func (c *Consumer) handleSnapshot(event DebeziumEvent) error {
	c.logger.Info(
		"CDC SNAPSHOT event",
		"table",
		event.Payload.Source.Table,
	)

	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
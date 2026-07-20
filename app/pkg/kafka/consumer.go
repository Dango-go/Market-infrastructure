package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/segmentio/kafka-go"
)

// Handler processes a single decoded envelope. Returning an error prevents the offset
// from being committed, so the message is redelivered (at-least-once). Handlers must be
// idempotent on Envelope.ID.
type Handler func(ctx context.Context, env events.Envelope) error

// Consumer reads one topic within a consumer group and dispatches to a Handler.
type Consumer struct {
	reader *kafka.Reader
}

// NewConsumer builds a Consumer for a topic and group.
func NewConsumer(brokers []string, group string, topic events.Topic) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			GroupID: group,
			Topic:   string(topic),
		}),
	}
}

// Run consumes until ctx is cancelled, committing offsets only after the handler
// succeeds. A handler error leaves the message uncommitted for redelivery.
func (c *Consumer) Run(ctx context.Context, h Handler) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return fmt.Errorf("fetch message: %w", err)
		}
		var env events.Envelope
		if err := json.Unmarshal(msg.Value, &env); err != nil {
			// Poison message: commit to skip; structured logging happens in the caller's handler wrapper.
			_ = c.reader.CommitMessages(ctx, msg)
			continue
		}
		if err := h(ctx, env); err != nil {
			return fmt.Errorf("handle %s: %w", env.Type, err)
		}
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("commit offset: %w", err)
		}
	}
}

// Close closes the underlying reader.
func (c *Consumer) Close() error { return c.reader.Close() }

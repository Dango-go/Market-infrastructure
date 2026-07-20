// Package kafka wraps segmentio/kafka-go with envelope-aware producer and consumer-group
// helpers. The transactional-outbox relay (in each service) calls Producer.Publish to
// drain pending events; consumers use Consumer to process a topic idempotently.
package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/embedded-market/backend/pkg/events"
	"github.com/segmentio/kafka-go"
)

// Producer publishes event envelopes to Kafka.
type Producer struct {
	writer *kafka.Writer
}

// NewProducer builds a Producer that routes each message to the topic named in its
// envelope. Keys are the envelope subject, preserving per-aggregate ordering.
func NewProducer(brokers []string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Balancer:               &kafka.Hash{},
			AllowAutoTopicCreation: true,
			RequiredAcks:           kafka.RequireAll,
		},
	}
}

// Publish writes one or more envelopes, each to the topic identified by its Type.
func (p *Producer) Publish(ctx context.Context, envs ...events.Envelope) error {
	msgs := make([]kafka.Message, 0, len(envs))
	for _, env := range envs {
		body, err := json.Marshal(env)
		if err != nil {
			return fmt.Errorf("marshal envelope %s: %w", env.ID, err)
		}
		msgs = append(msgs, kafka.Message{
			Topic: string(env.Type),
			Key:   []byte(env.Subject),
			Value: body,
			Headers: []kafka.Header{
				{Key: "event-id", Value: []byte(env.ID.String())},
				{Key: "correlation-id", Value: []byte(env.CorrelationID)},
			},
		})
	}
	return p.writer.WriteMessages(ctx, msgs...)
}

// Close flushes and closes the underlying writer.
func (p *Producer) Close() error { return p.writer.Close() }

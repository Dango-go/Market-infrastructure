// Package events implements the transactional-outbox relay: it drains events the use
// cases enqueued (in the same transaction as their state change) and publishes them to
// Kafka, guaranteeing at-least-once delivery without dual-write data loss.
package events

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	pkgkafka "github.com/embedded-market/backend/pkg/kafka"
	"github.com/embedded-market/backend/services/auth/internal/domain"
)

// Relay polls the outbox and forwards pending events to Kafka.
type Relay struct {
	store    domain.Store
	producer *pkgkafka.Producer
	clock    domain.Clock
	log      *zap.Logger
	interval time.Duration
	batch    int32
}

// NewRelay builds an outbox relay.
func NewRelay(store domain.Store, producer *pkgkafka.Producer, clock domain.Clock, log *zap.Logger, interval time.Duration, batch int32) *Relay {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	if batch <= 0 {
		batch = 100
	}
	return &Relay{store: store, producer: producer, clock: clock, log: log, interval: interval, batch: batch}
}

// Run drains the outbox on each tick until ctx is cancelled.
func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.drain(ctx); err != nil {
				r.log.Warn("outbox relay drain failed", zap.Error(err))
			}
		}
	}
}

// drain fetches a batch with FOR UPDATE SKIP LOCKED, publishes it, and marks it published
// — all inside one transaction so the row locks prevent other replicas from grabbing the
// same events. If publish succeeds but the commit fails the events are re-published on a
// later tick; consumers are idempotent on the envelope id, so this is safe.
func (r *Relay) drain(ctx context.Context) error {
	return r.store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		envs, err := tx.Outbox().FetchUnpublished(ctx, r.batch)
		if err != nil {
			return err
		}
		if len(envs) == 0 {
			return nil
		}
		if err := r.producer.Publish(ctx, envs...); err != nil {
			return err
		}
		ids := make([]uuid.UUID, 0, len(envs))
		for _, e := range envs {
			ids = append(ids, e.ID)
		}
		return tx.Outbox().MarkPublished(ctx, ids, r.clock.Now())
	})
}

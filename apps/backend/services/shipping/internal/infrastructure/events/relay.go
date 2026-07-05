package events

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	pkgkafka "github.com/embedded-market/backend/pkg/kafka"
	"github.com/embedded-market/backend/services/shipping/internal/domain"
)

type Relay struct {
	store    domain.Store
	producer *pkgkafka.Producer
	clock    domain.Clock
	log      *zap.Logger
	interval time.Duration
	batch    int32
}

func NewRelay(store domain.Store, producer *pkgkafka.Producer, clock domain.Clock, log *zap.Logger, interval time.Duration, batch int32) *Relay {
	if interval <= 0 { interval = 2 * time.Second }
	if batch <= 0 { batch = 100 }
	return &Relay{store: store, producer: producer, clock: clock, log: log, interval: interval, batch: batch}
}

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

func (r *Relay) drain(ctx context.Context) error {
	return r.store.WithinTx(ctx, func(ctx context.Context, tx domain.Store) error {
		envs, err := tx.Outbox().FetchUnpublished(ctx, r.batch)
		if err != nil { return err }
		if len(envs) == 0 { return nil }
		if err := r.producer.Publish(ctx, envs...); err != nil { return err }
		ids := make([]uuid.UUID, 0, len(envs))
		for _, env := range envs { ids = append(ids, env.ID) }
		return tx.Outbox().MarkPublished(ctx, ids, r.clock.Now())
	})
}

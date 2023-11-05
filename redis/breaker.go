package redis

import (
	"context"
	"time"

	"magicrouter/core"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type BreakerService struct {
	client *redis.Client
	cfg    core.BreakerConfig
}

// BreakerRecord is the data structure stored in Redis.
type BreakerRecord struct {
	// Failures is the number of failures since the last reset.
	Failures int `redis:"failures"`
	// LastFailure is the timestamp of the last failure.
	LastFailure time.Time `redis:"last_failure"`
}

func (r BreakerRecord) State(cfg core.BreakerConfig) core.BreakerState {
	if r.Failures >= cfg.MaxFailures {
		if time.Since(r.LastFailure) > cfg.ResetTimeout {
			return core.BreakerStateHalfOpen
		}
		return core.BreakerStateOpen
	}
	return core.BreakerStateClosed
}

func (b *BreakerService) GetState(ctx context.Context, breakerID string) core.BreakerState {
	var record BreakerRecord
	err := b.client.HGetAll(ctx, breakerID).Scan(&record)
	if err != nil {
		log.Err(err).Str("breaker_id", breakerID).Msg("failed to get breaker record from redis")
		return core.BreakerStateClosed
	}
	return record.State(b.cfg)
}

func (b *BreakerService) ReportFailure(ctx context.Context, breakerID string) {
	b.mutateRecord(ctx, breakerID, func(record BreakerRecord) BreakerRecord {
		record.Failures++
		record.LastFailure = time.Now().UTC()
		return record
	})
}

func (b *BreakerService) ReportSuccess(ctx context.Context, breakerID string) {
	b.mutateRecord(ctx, breakerID, func(record BreakerRecord) BreakerRecord {
		record.Failures = 0
		return record
	})
}

func (b *BreakerService) mutateRecord(ctx context.Context, breakerID string, mutator func(BreakerRecord) BreakerRecord) {
	var record BreakerRecord
	err := b.client.HGetAll(ctx, breakerID).Scan(&record)
	if err != nil && err != redis.Nil {
		log.Err(err).Str("breaker_id", breakerID).Msg("failed to get breaker record from redis")
		return
	}

	record = mutator(record)

	err = b.client.HSet(ctx, breakerID, record).Err()
	if err != nil {
		log.Err(err).Str("breaker_id", breakerID).Msg("failed to update breaker record in redis")
		return
	}
}

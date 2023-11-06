package redis

import (
	"context"
	"fmt"
	"time"

	"magicrouter/core"

	"github.com/redis/go-redis/v9"
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

func (b *BreakerService) GetState(ctx context.Context, breakerID string) (core.BreakerState, error) {
	var record BreakerRecord
	err := b.client.HGetAll(ctx, breakerID).Scan(&record)
	if err != nil {
		return core.BreakerStateClosed, fmt.Errorf("failed to get breaker record from redis: %w", err)
	}
	return record.State(b.cfg), nil
}

func (b *BreakerService) ReportFailure(ctx context.Context, breakerID string) error {
	_, err := b.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HIncrBy(ctx, breakerID, "failures", 1)
		pipe.HSet(ctx, breakerID, "last_failure", time.Now().UTC().Format(time.RFC3339Nano))
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to increment failure count in redis: %w", err)
	}
	return nil
}

func (b *BreakerService) ReportSuccess(ctx context.Context, breakerID string) error {
	err := b.client.HSet(ctx, breakerID, "failures", 0).Err()
	if err != nil {
		return fmt.Errorf("failed to reset failure count in redis: %w", err)
	}
	return nil
}

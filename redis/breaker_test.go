package redis

import (
	"context"
	"os"
	"testing"
	"time"

	"magicrouter/core"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestBreakerService(t *testing.T) {
	t.Parallel()
	if os.Getenv("REDIS_ADDR") == "" {
		t.Skip("redis not available")
	}
	breaker := &BreakerService{
		client: redis.NewClient(&redis.Options{
			Addr: os.Getenv("REDIS_ADDR"),
		}),
		cfg: core.BreakerConfig{
			MaxFailures:  10,
			ResetTimeout: time.Second * 1,
		},
	}
	ctx := context.Background()
	breakerID := "test"
	t.Cleanup(func() {
		breaker.client.Del(ctx, breakerID)
	})

	// Starts off closed
	// { "failures": 0, "last_failure": null, "last_reset": null }
	state := breaker.GetState(ctx, breakerID)
	assert.Equal(t, core.BreakerStateClosed, state)

	// After 10 failures, it opens
	// { "failures": 10, "last_failure": t1, "last_reset": t1 }
	for i := 0; i < 10; i++ {
		breaker.ReportFailure(ctx, breakerID)
	}
	state = breaker.GetState(ctx, breakerID)
	assert.Equal(t, core.BreakerStateOpen, state)

	// After reset timeout it goes half-open
	// { "failures": 10, "last_failure": t1, "last_reset": t1 }
	// last_reset + reset_timeout > now ie. enough time has passed since last failure
	time.Sleep(time.Second * 1)
	state = breaker.GetState(ctx, breakerID)
	assert.Equal(t, core.BreakerStateHalfOpen, state)

	// After a failure it goes back to open
	// { "failures": 11, "last_failure": t2, "last_reset": t2 }
	// last_reset + reset_timeout < now ie. not enough time has passed since last failure
	breaker.ReportFailure(ctx, breakerID)
	state = breaker.GetState(ctx, breakerID)
	assert.Equal(t, core.BreakerStateOpen, state)

	// After reset timeout it goes half-open
	// { "failures": 11, "last_failure": t2, "last_reset": t2 }
	// last_reset + reset_timeout > now ie. enough time has passed since last failure
	time.Sleep(time.Second * 1)
	state = breaker.GetState(ctx, breakerID)
	assert.Equal(t, core.BreakerStateHalfOpen, state)

	// After a success it goes back to closed
	// { "failures": 0, "last_failure": t2, "last_reset": t2 }
	breaker.ReportSuccess(ctx, breakerID)
	state = breaker.GetState(ctx, breakerID)
	assert.Equal(t, core.BreakerStateClosed, state)
}

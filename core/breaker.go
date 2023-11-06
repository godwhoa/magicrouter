package core

import (
	"context"
	"time"
)

type BreakerState uint

const (
	BreakerStateClosed BreakerState = iota
	BreakerStateOpen
	BreakerStateHalfOpen
)

func (s BreakerState) ShouldAttempt() bool {
	return s == BreakerStateClosed || s == BreakerStateHalfOpen
}

func (s BreakerState) String() string {
	switch s {
	case BreakerStateClosed:
		return "closed"
	case BreakerStateOpen:
		return "open"
	case BreakerStateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

type BreakerConfig struct {
	// MaxFailures is the number of failures before the breaker opens.
	MaxFailures int
	// ResetTimeout is the amount of time before the breaker resets to closed.
	ResetTimeout time.Duration
}

type BreakerService interface {
	GetState(ctx context.Context, breakerID string) (BreakerState, error)
	ReportFailure(ctx context.Context, breakerID string) error
	ReportSuccess(ctx context.Context, breakerID string) error
}

type NoOpBreaker struct{}

func (n NoOpBreaker) GetState(ctx context.Context, breakerID string) (BreakerState, error) {
	return BreakerStateClosed, nil
}

func (n NoOpBreaker) ReportSuccess(ctx context.Context, breakerID string) error { return nil }

func (n NoOpBreaker) ReportFailure(ctx context.Context, breakerID string) error { return nil }

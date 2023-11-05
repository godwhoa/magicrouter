package core

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

type BreakerService interface {
	GetState(breakerID string) BreakerState
	ReportFailure(breakerID string)
	ReportSuccess(breakerID string)
}

type NoOpBreaker struct{}

func (n NoOpBreaker) GetState(breakerID string) BreakerState {
	return BreakerStateClosed
}

func (n NoOpBreaker) ReportFailure(breakerID string) {}

func (n NoOpBreaker) ReportSuccess(breakerID string) {}

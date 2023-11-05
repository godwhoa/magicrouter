// Code generated by mockery v2.36.0. DO NOT EDIT.

package mocks

import (
	context "context"
	core "magicrouter/core"

	mock "github.com/stretchr/testify/mock"
)

// BreakerService is an autogenerated mock type for the BreakerService type
type BreakerService struct {
	mock.Mock
}

// GetState provides a mock function with given fields: ctx, breakerID
func (_m *BreakerService) GetState(ctx context.Context, breakerID string) core.BreakerState {
	ret := _m.Called(ctx, breakerID)

	var r0 core.BreakerState
	if rf, ok := ret.Get(0).(func(context.Context, string) core.BreakerState); ok {
		r0 = rf(ctx, breakerID)
	} else {
		r0 = ret.Get(0).(core.BreakerState)
	}

	return r0
}

// ReportFailure provides a mock function with given fields: ctx, breakerID
func (_m *BreakerService) ReportFailure(ctx context.Context, breakerID string) {
	_m.Called(ctx, breakerID)
}

// ReportSuccess provides a mock function with given fields: ctx, breakerID
func (_m *BreakerService) ReportSuccess(ctx context.Context, breakerID string) {
	_m.Called(ctx, breakerID)
}

// NewBreakerService creates a new instance of BreakerService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBreakerService(t interface {
	mock.TestingT
	Cleanup(func())
}) *BreakerService {
	mock := &BreakerService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
